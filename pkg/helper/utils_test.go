package helper_test

import (
	"deployment-notifications/pkg/helper"
	"github.com/stretchr/testify/assert"
	"os"

	"testing"
)


func TestDecodeStringJSON(t *testing.T) {
	sampleMap := `
{
	"key1": "value1",
	"key2": "value2"
}	
`

	outputMap, err := helper.DecodeStringJSON(sampleMap)
	expectedMap := make(map[string]string)

	expectedMap["key1"] = "value1"
	expectedMap["key2"] = "value2"

	assert.Equal(t, nil, err)
	assert.Equal(t, expectedMap, outputMap)

	sampleMap = `
{
	"key1": value1",
	"key2": "value2"
}	
`
	outputMap, err = helper.DecodeStringJSON(sampleMap)
	assert.NotNil(t, err)
}


func TestGetStringEnv(t *testing.T) {
	os.Setenv("ENV_VAR", "value")
	defer os.Unsetenv("ENV_VAR")

	envOut := helper.GetStringEnv("ENV_VAR", "")
	assert.Equal(t, "value", envOut)

	envOut = helper.GetStringEnv("ENV_VAR_MISSING", "missing")
	assert.Equal(t, "missing", envOut)
}


func TestLocateValue(t *testing.T) {
	expectedMap := make(map[string]string)

	expectedMap["key1"] = "value1"
	expectedMap["key2"] = "value2"

	result := helper.LocateValue("key1", expectedMap)
	assert.Equal(t, "value1", result)
	result = helper.LocateValue("key2", expectedMap)
	assert.Equal(t, "value2", result)
	result = helper.LocateValue("key3", expectedMap)
	assert.Equal(t, "", result)
}


func TestLocateValueMultiple(t *testing.T) {
	expectedMap := make(map[string][]string)

	expectedMap["key1"] = []string{"1", "2", "3"}
	expectedMap["key2"] = []string{"4", "5", "6"}

	result := helper.LocateValueMultiple("key1", expectedMap)
	assert.Equal(t, []string{"1", "2", "3"}, result)
	result = helper.LocateValueMultiple("key2", expectedMap)
	assert.Equal(t, []string{"4", "5", "6"}, result)
	result = helper.LocateValueMultiple("key3", expectedMap)
	assert.Equal(t, []string{}, result)
}