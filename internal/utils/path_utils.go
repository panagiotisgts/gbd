package utils

import (
	"path/filepath"
)

func ExtractFileName(path string) string {
	_, file := filepath.Split(path)
	return file
}

func FindAndReplace(keyPath []string, value any, m map[string]any) {
	if len(keyPath) == 1 {
		m[keyPath[0]] = value
		return
	}
	FindAndReplace(keyPath[1:], value, m[keyPath[0]].(map[string]any))
}
