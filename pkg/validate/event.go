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
	newRelicAPITokenARN := helper.GetStringEnv("NEW_RELIC_API_TOKEN", "")
	slackAPITokenARN := helper.GetStringEnv("SLACK_API_TOKEN", "")
	localExecution := helper.GetStringEnv("LOCAL_EXECUTION", "")
	awsAccountNumber := helper.GetStringEnv("AWS_ACCOUNT_NUMBER", "")
	newRelicBaseDomain := helper.GetStringEnv("NEW_RELIC_BASE_DOMAIN", "api.eu.newrelic.com")

	// placeholder env var in case
	// we want to test the Lambda locally
	if localExecution == "" {
		localExecution = "false"
	} else {
		localExecution = "true"
	}

	switch {
	case ssmParameterNameNewRelic == "":
		return result, helper.WrapError("Env var SSM_PARAMETER_NAME_NEW_RELIC is missing", nil)
	case ssmParameterNameSlack == "":
		return result, helper.WrapError("Env var SSM_PARAMETER_NAME_SLACK is missing", nil)
	case newRelicAPITokenARN == "":
		return result, helper.WrapError("Env Var NEW_RELIC_API_TOKEN is missing", nil)
	case slackAPITokenARN == "":
		return result, helper.WrapError("Env Var SLACK_API_TOKEN is missing", nil)
	case awsAccountNumber == "":
		return result, helper.WrapError("Env Var AWS_ACCOUNT_NUMBER is missing", nil)
	}

	result["SSM_PARAMETER_NAME_NEW_RELIC"] = ssmParameterNameNewRelic
	result["SSM_PARAMETER_NAME_SLACK"] = ssmParameterNameSlack
	result["NEW_RELIC_API_TOKEN"] = newRelicAPITokenARN
	result["SLACK_API_TOKEN"] = slackAPITokenARN
	result["AWS_ACCOUNT_NUMBER"] = awsAccountNumber
	result["LOCAL_EXECUTION"] = localExecution
	result["NEW_RELIC_BASE_DOMAIN"] = newRelicBaseDomain

	return result, nil
}
