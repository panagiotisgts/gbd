package utils

import (
	"encoding/json"
	"math"

	"github.com/AsaiYusuke/jsonpath"
)

func FindValueInJson(source []byte, path string) (any, error) {

	config := jsonpath.Config{}
	config.SetAccessorMode()

	var inputPayload interface{}
	if err := json.Unmarshal(source, &inputPayload); err != nil {
		return nil, err
	}

	jsonNode, err := jsonpath.Retrieve(path, inputPayload, config)
	if err != nil {
		return nil, err
	}

	return parseValue(jsonNode[0].(jsonpath.Accessor).Get()), nil
}

func parseValue(val any) any {
	switch val.(type) {
	case string, bool, []any:
		return val
	case float64, float32:
		if math.Trunc(val.(float64)) == val.(float64) {
			return int(val.(float64))
		}
		return val.(float64)
	default:
		return nil
	}
}
