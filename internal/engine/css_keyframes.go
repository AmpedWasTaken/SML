package engine

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type brightnessStop struct { position, brightness int }

func cssMilliseconds(s string) (int, error) {
	d, err := time.ParseDuration(s)
	if err != nil || d < 250*time.Millisecond || d > 10*time.Second || d%time.Millisecond != 0 {
		return 0, fmt.Errorf("animation duration must be 250ms..10s")
	}
	return int(d/time.Millisecond), nil
}

func cssKeyframes(body string) (string, error) {
	var stops []brightnessStop
	for strings.TrimSpace(body) != "" {
		head, declarations, tail, err := cssBlock(body)
		if err != nil { return "", err }
		body = tail
		position := 0
		switch head {
		case "from":
		case "to": position = 100
		default:
			position, err = strconv.Atoi(strings.TrimSuffix(head,"%"))
			if err != nil || !strings.HasSuffix(head,"%") || position < 0 || position > 100 {
				return "", fmt.Errorf("invalid keyframe position %q",head)
			}
		}
		attrs, err := cssDeclarations(declarations)
		if err != nil { return "", err }
		raw := attrs["brightness"]
		value, err := strconv.Atoi(strings.TrimSuffix(raw,"%"))
		if len(attrs) != 1 || err != nil || !strings.HasSuffix(raw,"%") || value < 30 || value > 100 {
			return "", fmt.Errorf("keyframes support brightness: 30%%..100%% only")
		}
		stops = append(stops, brightnessStop{position,value})
		if len(stops) > 8 { return "", fmt.Errorf("too many keyframe stops") }
	}
	sort.Slice(stops,func(i,j int) bool { return stops[i].position < stops[j].position })
	parts := make([]string,len(stops))
	for i,s := range stops { parts[i] = fmt.Sprintf("%d:%d",s.position,s.brightness) }
	curve := strings.Join(parts,",")
	if _, err := parseBrightnessCurve(curve); err != nil { return "",err }
	return curve,nil
}

func parseBrightnessCurve(curve string) ([]brightnessStop,error) {
	parts := strings.Split(curve,",")
	if len(parts) < 2 || len(parts) > 8 { return nil,fmt.Errorf("animation needs 2..8 keyframe stops") }
	stops := make([]brightnessStop,0,len(parts))
	for _,part := range parts {
		a,b,ok := strings.Cut(part,":")
		pos,e1 := strconv.Atoi(a)
		val,e2 := strconv.Atoi(b)
		if !ok || e1 != nil || e2 != nil || pos < 0 || pos > 100 || val < 30 || val > 100 || (len(stops)>0 && pos <= stops[len(stops)-1].position) {
			return nil,fmt.Errorf("invalid brightness keyframe %q",part)
		}
		stops = append(stops,brightnessStop{pos,val})
	}
	if stops[0].position != 0 || stops[len(stops)-1].position != 100 { return nil,fmt.Errorf("keyframes require 0%% and 100%% endpoints") }
	return stops,nil
}

func keyframeStrength(curve string, progress float64) float64 {
	stops,err := parseBrightnessCurve(curve)
	if err != nil { return 1 }
	p := progress*100
	for i:=1; i<len(stops); i++ {
		a,b:=stops[i-1],stops[i]
		if p <= float64(b.position) {
			t:=(p-float64(a.position))/float64(b.position-a.position)
			return (float64(a.brightness)+t*float64(b.brightness-a.brightness))/100
		}
	}
	return float64(stops[len(stops)-1].brightness)/100
}
