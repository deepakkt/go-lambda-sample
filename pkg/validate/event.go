package validate

import (
	"deployment-notifications/pkg/helper"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

func SourceValidate(request events.CloudWatchEvent) (string, error) {
	// ignore non ECS events
	if strings.ToLower(request.Source) != "aws.ecs" {
		outMessage := fmt.Sprintf("Event '%s' received. We only respond to 'aws.ecs' events", request.Source)
		return outMessage, helper.WrapError(outMessage, nil)
	}

	return "", nil
}

func DetailValidate(request events.CloudWatchEvent) (string, error) {
	// ignore non deployment ECS events
	if strings.ToLower(request.DetailType) != "ecs deployment state change" {
		outMessage := fmt.Sprintf("ECS Event '%s' received. We only respond to 'ECS Deployment State Change' events", request.DetailType)
		return outMessage, helper.WrapError(outMessage, nil)
	}

	return "", nil
}

func EnvValidate() (map[string]string, error) {
	result := make(map[string]string)

	ssmParameterNameNewRelic := helper.GetStringEnv("SSM_PARAMETER_NAME_NEW_RELIC", "")
	ssmParameterNameSlack := helper.GetStringEnv("SSM_PARAMETER_NAME_SLACK", "")
	ssmParameterMessageSlack := helper.GetStringEnv("SSM_PARAMETER_MESSAGE_SLACK", "")
	newRelicAPITokenARN := helper.GetStringEnv("NEW_RELIC_API_TOKEN", "")
	newRelicBaseDomain := helper.GetStringEnv("NEW_RELIC_BASE_DOMAIN", "api.eu.newrelic.com")

	switch {
	case ssmParameterNameNewRelic == "":
		return result, helper.WrapError("Env var SSM_PARAMETER_NAME_NEW_RELIC is missing", nil)
	case ssmParameterNameSlack == "":
		return result, helper.WrapError("Env var SSM_PARAMETER_NAME_SLACK is missing", nil)
	case ssmParameterMessageSlack == "":
		return result, helper.WrapError("Env var SSM_PARAMETER_MESSAGE_SLACK is missing", nil)
	case newRelicAPITokenARN == "":
		return result, helper.WrapError("Env Var NEW_RELIC_API_TOKEN is missing", nil)
	}

	result["SSM_PARAMETER_NAME_NEW_RELIC"] = ssmParameterNameNewRelic
	result["SSM_PARAMETER_NAME_SLACK"] = ssmParameterNameSlack
	result["SSM_PARAMETER_MESSAGE_SLACK"] = ssmParameterMessageSlack
	result["NEW_RELIC_API_TOKEN"] = newRelicAPITokenARN
	result["NEW_RELIC_BASE_DOMAIN"] = newRelicBaseDomain

	return result, nil
}
