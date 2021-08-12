package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

func WrapError(errorMessage string, err error) error {
	if err == nil {
		return errors.New(errorMessage)
	}

	return fmt.Errorf("%s: %w", errorMessage, err)
}

func DecodeStringJSON(parameterString string) (map[string]string, error) {
	//this function assumes that the parameter string is a series of
	//key value pairs which are all string. Any other input type will
	//error out and the error is returned as is

	resultMap := make(map[string]string)
	err := json.Unmarshal([]byte(parameterString), &resultMap)

	if err != nil {
		return resultMap, WrapError(fmt.Sprintf("Could not decode parameter string\n%s", parameterString),
			nil)
	}

	return resultMap, nil
}

func GetStringEnv(name string, defaultValue string) string {
	stringValue := os.Getenv(name)
	if stringValue == "" {
		return defaultValue
	}

	return stringValue
}

func LocateValueMultiple(inputKey string, valuesMap map[string][]string) []string {
	for key, value := range valuesMap {
		if key == inputKey {
			return value
		}
	}

	return []string{}
}
