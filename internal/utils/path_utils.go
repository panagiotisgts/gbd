package utils

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func ExtractFileName(path string) string {
	_, file := filepath.Split(path)
	return file
}

func FindAndReplace(keyPath []string, value any, m map[string]any) {
	if len(keyPath) == 1 {
		if isIndex(keyPath[0]) {
			fmt.Println(m)
			return
		}
		m[keyPath[0]] = value
		return
	}
	path := keyPath[1:]
	if len(path) == 1 && isIndex(path[0]) {
		idx, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(path[0], "["), "]"))
		if err != nil {
			panic(err)
		}
		findAndReplaceArray(idx, value, m[keyPath[0]])
	} else {
		FindAndReplace(path, value, m[keyPath[0]].(map[string]any))
	}
}

func findAndReplaceArray(idx int, value any, m any) {
	switch m.(type) {
	case []string:
		m.([]string)[idx] = value.(string)
	case []int:
		m.([]int)[idx] = value.(int)
	case []float64:
		m.([]float64)[idx] = value.(float64)
	case []bool:
		m.([]bool)[idx] = value.(bool)
	}
}

func isIndex(key string) bool {
	return regexp.MustCompile(`^\[\d+\]$`).MatchString(key)
}
