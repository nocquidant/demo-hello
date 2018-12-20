package api

import (
	"encoding/json"

	"github.com/google/logger"
)

func kvAsJson(key string, val string) string {
	m := make(map[string]string)
	m[key] = val
	data, err := json.Marshal(m)
	if err != nil {
		logger.Errorf("Error while serializing to json: %s", err)
		return ""
	}
	return string(data)
}

func kmAsJson(k string, v string, key string, val map[string]interface{}) string {
	m := make(map[string]interface{})
	m[k] = v
	m[key] = val
	data, err := json.Marshal(m)
	if err != nil {
		logger.Errorf("Error while serializing to json: %s", err)
		return ""
	}
	return string(data)
}

func mapAsJson(m map[string]interface{}) string {
	data, err := json.Marshal(m)
	if err != nil {
		logger.Errorf("Error while serializing to json: %s", err)
		return ""
	}
	return string(data)
}
