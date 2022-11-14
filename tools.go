package main

import "gopkg.in/yaml.v2"

func fieldValueIdenticalToDefaultValue(fieldName string, value interface{}) bool {
	if defaultValue, ok := defaultSpecValues[fieldName]; ok {
		if _, ok = value.([]interface{}); ok {
			if slicesIdentical(value.([]interface{}), defaultValue.([]interface{})) {
				return true
			}
		} else if value == defaultValue {
			return true
		}
		return false
	} else {
		return false
	}
}

func convertToMapItem(key string, value interface{}) yaml.MapItem {
	return yaml.MapItem{
		Key:   key,
		Value: value,
	}
}

func contains(source []string, target string) bool {
	for _, value := range source {
		if value == target {
			return true
		}
	}
	return false
}

func slicesIdentical(source []interface{}, target []interface{}) bool {
	if len(source) != len(target) {
		return false
	}
	for _, sourceValue := range source {
		for _, targetValue := range source {
			if sourceValue == targetValue {
				continue
			} else {
				return false
			}
		}
	}
	return true
}

func convertToMapString(i interface{}) map[string]interface{} {
	objMap := map[string]interface{}{}
	if result, ok := i.(map[string]interface{}); ok {
		objMap = result
	} else if result, ok := i.(map[interface{}]interface{}); ok {
		for key, value := range result {
			objMap[key.(string)] = value
		}
	}

	return objMap
}