//go:build unit

package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindAndReplace(t *testing.T) {
	cfg := map[string]any{
		"levelOne": map[string]any{
			"stringValue": "localhost",
			"intValue":    5432,
			"floatValue":  10.32,
			"boolValue":   true,
			"arrayValue": []string{
				"one",
				"two",
				"three",
			},
			"mapValue": map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	tests := []struct {
		path         []string
		replaceValue any
		jsonpath     string
	}{
		{[]string{"levelOne", "stringValue"}, "127.0.0.1", "$.levelOne.stringValue"},
		{[]string{"levelOne", "intValue"}, 15432, "$.levelOne.intValue"},
		{[]string{"levelOne", "floatValue"}, 11.32, "$.levelOne.floatValue"},
		{[]string{"levelOne", "boolValue"}, false, "$.levelOne.boolValue"},
		{[]string{"levelOne", "mapValue", "key1"}, "value3", "$.levelOne.mapValue.key1"},
		{[]string{"levelOne", "arrayValue", "[0]"}, "one1", "$.levelOne.arrayValue[0]"},
	}

	for _, c := range tests {
		FindAndReplace(c.path, c.replaceValue, cfg)
		bcfg, err := json.Marshal(cfg)
		require.Nil(t, err)
		val, err := FindValueInJson(bcfg, c.jsonpath)
		require.Nil(t, err)
		require.Equal(t, c.replaceValue, val)

	}
}

func TestFindValueInJson(t *testing.T) {
	cfg := map[string]any{
		"levelOne": map[string]any{
			"stringValue": "localhost",
			"intValue":    5432,
			"floatValue":  10.32,
			"boolValue":   true,
			"arrayValue": []string{
				"one",
				"two",
				"three",
			},
			"mapValue": map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	b, err := json.Marshal(cfg)
	require.Nil(t, err)

	tests := []struct {
		path string
		val  any
	}{
		{"$.levelOne.stringValue", "localhost"},
		{"$.levelOne.intValue", 5432},
		{"$.levelOne.boolValue", true},
		{"$.levelOne.floatValue", 10.32},
		{"$.levelOne.arrayValue[0]", "one"},
		{`$.levelOne.mapValue["key1"]`, "value1"},
	}

	for _, c := range tests {
		val, err := FindValueInJson(b, c.path)
		require.Nil(t, err)
		require.Equal(t, c.val, val)
	}
}
