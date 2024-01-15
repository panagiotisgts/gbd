package utils

import (
	"encoding/json"

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

	return jsonNode[0].(jsonpath.Accessor).Get(), nil
}
