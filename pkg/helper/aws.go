package helper

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/pkg/errors"
	"os"
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