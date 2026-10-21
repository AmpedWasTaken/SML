package engine

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maxMediaFile   = 1024 * 1024
	maxMediaPixels = 640 * 480
	maxMediaFrames = 40
)

var mediaName = regexp.MustCompile(`^[A-Za-z0-9_-]{1,48}\.(png|jpg|jpeg|gif)$`)

type mediaEntry struct {
	modified time.Time
	size     int64
	frames   []image.Image
	delays   []int // GIF delays in 1/100 seconds
}

var mediaCache sync.Map

func validateMediaNode(node Node) error {
	if node.Kind != "image" && node.Kind != "gif" {
		return nil
	}
	if len(node.Children) > 0 || strings.TrimSpace(node.Text) != "" {
		return errors.New("media elements must be empty")
	}
	allowed := map[string]bool{"src": true, "width": true, "height": true, "mode": true, "alt": true}
	if node.Kind == "gif" {
		allowed["fps"] = true
	}
	for key := range node.Attrs {
		if !allowed[key] {
			return fmt.Errorf("invalid media attribute %q", key)
		}
	}
	name := node.Attr("src", "")
	if !mediaName.MatchString(name) {
		return errors.New("invalid media filename (use local media/ assets)")
	}
	if node.Kind == "gif" && !strings.HasSuffix(name, ".gif") {
		return errors.New("gif requires .gif asset")
	}
	w, err := strconv.Atoi(node.Attr("width", "32"))
	if err != nil || w < 2 || w > 64 {
		return errors.New("media width must be 2..64 columns")
	}
	h, err := strconv.Atoi(node.Attr("height", "10"))
	if err != nil || h < 1 || h > 24 {
		return errors.New("media height must be 1..24 rows")
	}
	mode := node.Attr("mode", "ascii")
	if mode != "ascii" && mode != "blocks" {
		return errors.New("media mode must be ascii or blocks")
	}
	if len(node.Attr("alt", "")) > 120 {
		return errors.New("media alt text too long")
	}
	if node.Kind == "gif" {
		fps, err := strconv.Atoi(node.Attr("fps", "8"))
		if err != nil || fps < 1 || fps > 12 {
			return errors.New("gif fps must be between 1 and 12")
		}
	}
	return nil
}

func cloneRGBA(img *image.RGBA) *image.RGBA {
	clone := image.NewRGBA(img.Bounds())
	copy(clone.Pix, img.Pix)
	return clone
}

func decodeMedia(data []byte, animated bool) ([]image.Image, []int, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, nil, errors.New("unrecognized local image")
	}
	if cfg.Width < 1 || cfg.Height < 1 || cfg.Width*cfg.Height > maxMediaPixels {
		return nil, nil, errors.New("media image dimensions are too large")
	}
	if animated {
		if format != "gif" {
			return nil, nil, errors.New("animation asset is not a GIF")
		}
		decoded, err := gif.DecodeAll(bytes.NewReader(data))
		if err != nil || len(decoded.Image) == 0 || len(decoded.Image) > maxMediaFrames {
			return nil, nil, errors.New("invalid or oversized GIF frame sequence")
		}
		canvas := image.NewRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
		snapshots := make([]image.Image, 0, len(decoded.Image))
		var backup *image.RGBA
		for i, frame := range decoded.Image {
			if i > 0 && len(decoded.Disposal) > i-1 {
				switch decoded.Disposal[i-1] {
				case gif.DisposalBackground:
					bounds := decoded.Image[i-1].Bounds()
					draw.Draw(canvas, bounds, image.Transparent, image.Point{}, draw.Src)
				case gif.DisposalPrevious:
					if backup != nil {
						canvas = cloneRGBA(backup)
					}
				}
			}
			backup = cloneRGBA(canvas)
			draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)
			snapshots = append(snapshots, cloneRGBA(canvas))
		}
		return snapshots, decoded.Delay, nil
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, nil, errors.New("invalid local image")
	}
	return []image.Image{decoded}, nil, nil
}

func (v View) mediaFrames(name string, animated bool) (mediaEntry, error) {
	if v.assets == "" {
		return mediaEntry{}, errors.New("media assets folder not configured")
	}
	mediaDir := filepath.Join(v.assets, "media")
	rootInfo, rootErr := os.Lstat(mediaDir)
	if rootErr != nil || !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return mediaEntry{}, errors.New("media directory missing or not regular")
	}
	file := filepath.Join(mediaDir, name)
	info, err := os.Lstat(file)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxMediaFile {
		return mediaEntry{}, fmt.Errorf("media asset %q missing or invalid", name)
	}
	if cached, ok := mediaCache.Load(file); ok {
		entry := cached.(mediaEntry)
		if entry.size == info.Size() && entry.modified.Equal(info.ModTime()) {
			return entry, nil
		}
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return mediaEntry{}, fmt.Errorf("media asset %q unreadable", name)
	}
	frames, delays, err := decodeMedia(data, animated)
	if err != nil {
		return mediaEntry{}, err
	}
	entry := mediaEntry{modified: info.ModTime(), size: info.Size(), frames: frames, delays: delays}
	mediaCache.Store(file, entry)
	return entry, nil
}

func selectedFrame(entry mediaEntry, fps int) image.Image {
	if len(entry.frames) == 0 {
		return nil
	}
	if len(entry.frames) == 1 {
		return entry.frames[0]
	}
	var current int
	if fps > 0 {
		current = int(time.Now().UnixMilli()/int64(1000/fps)) % len(entry.frames)
	} else {
		total := 0
		for _, d := range entry.delays {
			total += max(2, d)
		}
		if total <= 0 {
			return entry.frames[0]
		}
		tick := int(time.Now().UnixMilli()/10) % total
		for i, d := range entry.delays {
			tick -= max(2, d)
			if tick < 0 {
				return entry.frames[i]
			}
		}
	}
	return entry.frames[current]
}

// renderTerminalMedia converts local images and GIF frames into safe text cells.
// This fallback works in PuTTY. No network downloads or terminal escape
// sequences are generated from media; actual inline bitmaps need a terminal
// supporting a separate protocol (Kitty/Sixel), not baseline PuTTY.
func (v View) renderTerminalMedia(node Node, width int) ([]string, error) {
	if err := validateMediaNode(node); err != nil {
		return nil, err
	}
	name := node.Attr("src", "")
	frames, err := v.mediaFrames(name, node.Kind == "gif")
	if err != nil {
		return nil, err
	}
	fps := 0
	if node.Kind == "gif" {
		fps, _ = strconv.Atoi(node.Attr("fps", "8"))
	}
	img := selectedFrame(frames, fps)
	if img == nil {
		return nil, errors.New("empty image")
	}
	columns, _ := strconv.Atoi(node.Attr("width", "32"))
	rows, _ := strconv.Atoi(node.Attr("height", "10"))
	columns = min(width, columns)
	if columns < 2 {
		columns = 2
	}
	mode := node.Attr("mode", "ascii")
	bounds := img.Bounds()
	shades := []rune("@%#*+=-:. ")
	if mode == "blocks" {
		shades = []rune("█▓▒░ ")
	}
	result := make([]string, rows)
	for y := 0; y < rows; y++ {
		var b strings.Builder
		for x := 0; x < columns; x++ {
			px := bounds.Min.X + int(float64(x)+0.5)*bounds.Dx()/columns
			py := bounds.Min.Y + int(float64(y)+0.5)*bounds.Dy()/rows
			if px >= bounds.Max.X {
				px = bounds.Max.X - 1
			}
			if py >= bounds.Max.Y {
				py = bounds.Max.Y - 1
			}
			r, g, blue, alpha := img.At(px, py).RGBA()
			lum := (2126*r + 7152*g + 722*blue) / 10000
			// Transparent pixels remain blank.
			if alpha < 32768 {
				b.WriteRune(' ')
				continue
			}
			index := int(lum * uint32(len(shades)-1) / 65535)
			b.WriteRune(shades[index])
		}
		result[y] = b.String()
	}
	return result, nil
}
