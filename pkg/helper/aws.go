package helper

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/pkg/errors"
	"os"
	"strings"
)


func GetECSServiceARN(serviceName, awsRegion, awsAccountNumber string) string {
	return fmt.Sprintf("arn:aws:ecs:%s:%s:service/%s",
		awsRegion, serviceName, awsAccountNumber)
}


func GetAwsDefaultRegion() string {
	val, exists := os.LookupEnv("AWS_REGION")
	if !exists || len(val) < 1 {
		val = os.Getenv("AWS_DEFAULT_REGION")
	}

	return val
}


func ReadAWSSecret(secretID string) (string, error) {
	awsSession := session.Must(session.NewSession())
	*awsSession.Config.Region = GetAwsDefaultRegion()

	session := secretsmanager.New(awsSession)

	secretValue, err := session.GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: &secretID})

	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Error getting secret from ID '%s'", secretID))
	}

	return *secretValue.SecretString, nil
}


func ReadAWSParameter(paramID string) (string, error) {
	awsSession := session.Must(session.NewSession())
	*awsSession.Config.Region = GetAwsDefaultRegion()

	session := ssm.New(awsSession)

	param, err := session.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(paramID),
		WithDecryption: aws.Bool(false),
	})

	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Error getting param value from ID '%s'", paramID))
	}

	return *param.Parameter.Value, nil
}


func GetServiceNameFromARN(arnString string) (string, error) {
	arnSplit := strings.Split(arnString, "service/")

	if len(arnSplit) < 2 {
		return "", WrapError(fmt.Sprintf("Input '%s' did not have expected 'service/' prefix", arnString), nil)
	}

	return arnSplit[1], nil
}

func ParseEventDetails(request events.CloudWatchEvent) (map[string]string, error) {
	type EventInfo struct {
		EventName    string `json:eventName`
		DeploymentID string `json:deploymentId`
		UpdatedAt    string `json:updatedAt`
		Comments     string `json:reason`
	}

	var eventInfo EventInfo
	result := make(map[string]string)

	err := json.Unmarshal(request.Detail, &eventInfo)

	if err != nil {
		return result, WrapError("Unexpected error unmarshaling event details", err)
	}

	if eventInfo.EventName == "" {
		return result, WrapError("'eventName' attribute not found on payload", nil)
	}
	if eventInfo.DeploymentID == "" {
		return result, WrapError("'deploymentId' attribute not found on payload", nil)
	}
	if eventInfo.UpdatedAt == "" {
		return result, WrapError("'updatedAt' attribute not found on payload", nil)
	}

	result["eventName"] = eventInfo.EventName
	result["deploymentId"] = eventInfo.DeploymentID
	result["updatedAt"] = eventInfo.UpdatedAt
	result["reason"] = eventInfo.Comments

	return result, nil
}


