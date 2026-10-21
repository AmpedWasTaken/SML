package sml

import (
	"fmt"
	"os"
)

// ParseFile parses a standalone .sml view from a regular local file.
// Includes and media assets require an asset root via View.WithAssets.
func ParseFile(path string) (View, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return View{}, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxFileSize {
		return View{}, fmt.Errorf("invalid SML view file: expected regular file up to %d bytes", maxFileSize)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return View{}, err
	}
	return Parse(data)
}
