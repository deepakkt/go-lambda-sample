package validate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

func WrapError(errorMessage string, err error) error {
	if err == nil {
		return errors.New(errorMessage)
	}

	return fmt.Errorf("%s: %w", errorMessage, err)
}


func SourceValidate(request events.CloudWatchEvent) (string, error) {
	// ignore non ECS events
	if strings.ToLower(request.Source) != "aws.ecs" {
		outMessage := fmt.Sprintf("Event '%s' received. We only respond to 'aws.ecs' events", request.Source)
		return outMessage, WrapError(outMessage, nil)
	}

	return "", nil
}


func DetailValidate(request events.CloudWatchEvent) (string, error) {
	// ignore non deployment ECS events
	if strings.ToLower(request.DetailType) != "ecs deployment state change" {
		outMessage := fmt.Sprintf("ECS Event '%s' received. We only respond to 'ECS Deployment State Change' events", request.DetailType)
		return outMessage, WrapError(outMessage, nil)
	}

	return "", nil
}


func ParseEventName(request events.CloudWatchEvent) (string, error) {
	type EventInfo struct {
		EventName string `json:eventName`
	}

	var eventInfo EventInfo
	err := json.Unmarshal(request.Detail, &eventInfo)

	if err != nil {
		return "", WrapError("Unexpected error unmarshaling event name", err)
	}

	if eventInfo.EventName == "" {
		return "", WrapError("'eventName' attribute not found on payload", nil)
	}

	return eventInfo.EventName, nil
}

func getStringEnv(name string, defaultValue string) string {
	stringValue := os.Getenv(name)
	if stringValue == "" {
		return defaultValue
	}

	return stringValue
}


func EnvValidate() (map[string]string, error) {
	result := make(map[string]string)

	ssmParameterName := getStringEnv("SSM_PARAMETER_NAME", "")
	newRelicAPITokenARN := getStringEnv("NEW_RELIC_API_TOKEN", "")
	localExecution := getStringEnv("LOCAL_EXECUTION", "")

	if localExecution == "" {
		localExecution = "false"
	} else {
		localExecution = "true"
	}

	switch {
	case ssmParameterName == "":
		return result, WrapError("Env var SSM_PARAMETER_NAME is missing", nil)
	case newRelicAPITokenARN == "":
		return result, WrapError("Env Var NEW_RELIC_API_TOKEN is missing", nil)
	}

	result["SSM_PARAMETER_NAME"] = ssmParameterName
	result["NEW_RELIC_API_TOKEN"] = newRelicAPITokenARN
	result["LOCAL_EXECUTION"] = localExecution

	return result, nil
}
