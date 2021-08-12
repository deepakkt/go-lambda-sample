package validate_test

import (
	"deployment-notifications/pkg/validate"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestSourceValidatePass(t *testing.T) {
	sampleEvent := `
{                                                                         
   "version": "0",                                                        
   "id": "ddca6449-b258-46c0-8653-e0e3a6EXAMPLE",                         
   "detail-type": "ECS Deployment State Change",                          
   "source": "AWS.ECS",                                                   
   "account": "111122223333",                                             
   "time": "2020-05-23T12:31:14Z",                                        
   "region": "us-west-2",                                                 
   "resources": [                                                         
        "arn:aws:ecs:us-west-2:111122223333:service/shure-content-api"    
   ],                                                                     
   "detail": {                                                            
        "eventType": "INFO",                                              
        "eventName": "SERVICE_DEPLOYMENT_COMPLETED",                      
        "deploymentId": "ecs-svc/123",                                    
        "updatedAt": "2020-05-23T11:11:11Z",                              
        "reason": "ECS deployment deploymentId in progress."              
   }                                                                      
}                                                                         
`
	var cloudwatchEvent events.CloudWatchEvent
	err := json.Unmarshal([]byte(sampleEvent), &cloudwatchEvent)
	assert.Nil(t, err)

	validateMessage, err := validate.SourceValidate(cloudwatchEvent)

	assert.Equal(t, "", validateMessage)
	assert.Nil(t, err)
}

func TestSourceValidateFail(t *testing.T) {
	sampleEvent := `
{                                                                         
   "version": "0",                                                        
   "id": "ddca6449-b258-46c0-8653-e0e3a6EXAMPLE",                         
   "detail-type": "ECS Deployment State Change",                          
   "source": "AWS.EC2",                                                   
   "account": "111122223333",                                             
   "time": "2020-05-23T12:31:14Z",                                        
   "region": "us-west-2",                                                 
   "resources": [                                                         
        "arn:aws:ecs:us-west-2:111122223333:service/shure-content-api"    
   ],                                                                     
   "detail": {                                                            
        "eventType": "INFO",                                              
        "eventName": "SERVICE_DEPLOYMENT_COMPLETED",                      
        "deploymentId": "ecs-svc/123",                                    
        "updatedAt": "2020-05-23T11:11:11Z",                              
        "reason": "ECS deployment deploymentId in progress."              
   }                                                                      
}                                                                         
`
	var cloudwatchEvent events.CloudWatchEvent
	err := json.Unmarshal([]byte(sampleEvent), &cloudwatchEvent)
	assert.Nil(t, err)

	validateMessage, err := validate.SourceValidate(cloudwatchEvent)

	assert.NotEqual(t, "", validateMessage)
	assert.NotNil(t, err)
}

func TestDetailValidatePass(t *testing.T) {
	sampleEvent := `
{                                                                         
   "version": "0",                                                        
   "id": "ddca6449-b258-46c0-8653-e0e3a6EXAMPLE",                         
   "detail-type": "ECS Deployment State Change",                          
   "source": "AWS.ECS",                                                   
   "account": "111122223333",                                             
   "time": "2020-05-23T12:31:14Z",                                        
   "region": "us-west-2",                                                 
   "resources": [                                                         
        "arn:aws:ecs:us-west-2:111122223333:service/shure-content-api"    
   ],                                                                     
   "detail": {                                                            
        "eventType": "INFO",                                              
        "eventName": "SERVICE_DEPLOYMENT_COMPLETED",                      
        "deploymentId": "ecs-svc/123",                                    
        "updatedAt": "2020-05-23T11:11:11Z",                              
        "reason": "ECS deployment deploymentId in progress."              
   }                                                                      
}                                                                         
`
	var cloudwatchEvent events.CloudWatchEvent
	err := json.Unmarshal([]byte(sampleEvent), &cloudwatchEvent)
	assert.Nil(t, err)

	validateMessage, err := validate.DetailValidate(cloudwatchEvent)

	assert.Equal(t, "", validateMessage)
	assert.Nil(t, err)
}

func TestDetailValidateFail(t *testing.T) {
	sampleEvent := `
{                                                                         
   "version": "0",                                                        
   "id": "ddca6449-b258-46c0-8653-e0e3a6EXAMPLE",                         
   "detail-type": "EC2 Deployment State Change",                          
   "source": "AWS.ECS",                                                   
   "account": "111122223333",                                             
   "time": "2020-05-23T12:31:14Z",                                        
   "region": "us-west-2",                                                 
   "resources": [                                                         
        "arn:aws:ecs:us-west-2:111122223333:service/shure-content-api"    
   ],                                                                     
   "detail": {                                                            
        "eventType": "INFO",                                              
        "eventName": "SERVICE_DEPLOYMENT_COMPLETED",                      
        "deploymentId": "ecs-svc/123",                                    
        "updatedAt": "2020-05-23T11:11:11Z",                              
        "reason": "ECS deployment deploymentId in progress."              
   }                                                                      
}                                                                         
`
	var cloudwatchEvent events.CloudWatchEvent
	err := json.Unmarshal([]byte(sampleEvent), &cloudwatchEvent)
	assert.Nil(t, err)

	validateMessage, err := validate.DetailValidate(cloudwatchEvent)

	assert.NotEqual(t, "", validateMessage)
	assert.NotNil(t, err)
}

func TestEnvValidatePass(t *testing.T) {
	os.Setenv("SSM_PARAMETER_NAME_NEW_RELIC", "param1")
	os.Setenv("SSM_PARAMETER_NAME_SLACK", "param2")
	os.Setenv("SSM_PARAMETER_MESSAGE_SLACK", "param3")
	os.Setenv("NEW_RELIC_API_TOKEN", "param4")
	os.Setenv("NEW_RELIC_BASE_DOMAIN", "param6")

	defer os.Unsetenv("SSM_PARAMETER_NAME_NEW_RELIC")
	defer os.Unsetenv("SSM_PARAMETER_NAME_SLACK")
	defer os.Unsetenv("SSM_PARAMETER_MESSAGE_SLACK")
	defer os.Unsetenv("NEW_RELIC_API_TOKEN")
	defer os.Unsetenv("NEW_RELIC_BASE_DOMAIN")

	result, err := validate.EnvValidate()

	assert.Nil(t, err)

	assert.Equal(t, "param1", result["SSM_PARAMETER_NAME_NEW_RELIC"])
	assert.Equal(t, "param2", result["SSM_PARAMETER_NAME_SLACK"])
	assert.Equal(t, "param3", result["SSM_PARAMETER_MESSAGE_SLACK"])
	assert.Equal(t, "param4", result["NEW_RELIC_API_TOKEN"])
	assert.Equal(t, "param6", result["NEW_RELIC_BASE_DOMAIN"])
}

func TestEnvValidatePassNoDomain(t *testing.T) {
	os.Setenv("SSM_PARAMETER_NAME_NEW_RELIC", "param1")
	os.Setenv("SSM_PARAMETER_NAME_SLACK", "param2")
	os.Setenv("SSM_PARAMETER_MESSAGE_SLACK", "param3")
	os.Setenv("NEW_RELIC_API_TOKEN", "param4")

	defer os.Unsetenv("SSM_PARAMETER_NAME_NEW_RELIC")
	defer os.Unsetenv("SSM_PARAMETER_NAME_SLACK")
	defer os.Unsetenv("SSM_PARAMETER_MESSAGE_SLACK")
	defer os.Unsetenv("NEW_RELIC_API_TOKEN")

	result, err := validate.EnvValidate()

	assert.Nil(t, err)

	assert.Equal(t, "param1", result["SSM_PARAMETER_NAME_NEW_RELIC"])
	assert.Equal(t, "param2", result["SSM_PARAMETER_NAME_SLACK"])
	assert.Equal(t, "param3", result["SSM_PARAMETER_MESSAGE_SLACK"])
	assert.Equal(t, "param4", result["NEW_RELIC_API_TOKEN"])
	assert.Equal(t, "api.eu.newrelic.com", result["NEW_RELIC_BASE_DOMAIN"])
}

func TestEnvValidateFailSSMNewRelic(t *testing.T) {
	os.Setenv("SSM_PARAMETER_NAME_SLACK", "param2")
	os.Setenv("SSM_PARAMETER_MESSAGE_SLACK", "param3")
	os.Setenv("NEW_RELIC_API_TOKEN", "param4")

	defer os.Unsetenv("SSM_PARAMETER_NAME_SLACK")
	defer os.Unsetenv("SSM_PARAMETER_MESSAGE_SLACK")
	defer os.Unsetenv("NEW_RELIC_API_TOKEN")

	_, err := validate.EnvValidate()

	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "SSM_PARAMETER_NAME_NEW_RELIC"))
}

func TestEnvValidateFailSSMSlack(t *testing.T) {
	os.Setenv("SSM_PARAMETER_NAME_NEW_RELIC", "param1")
	os.Setenv("SSM_PARAMETER_MESSAGE_SLACK", "param3")
	os.Setenv("NEW_RELIC_API_TOKEN", "param4")

	defer os.Unsetenv("SSM_PARAMETER_NAME_NEW_RELIC")
	defer os.Unsetenv("SSM_PARAMETER_MESSAGE_SLACK")
	defer os.Unsetenv("NEW_RELIC_API_TOKEN")

	_, err := validate.EnvValidate()

	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "SSM_PARAMETER_NAME_SLACK"))
}

func TestEnvValidateFailMessageSlack(t *testing.T) {
	os.Setenv("SSM_PARAMETER_NAME_NEW_RELIC", "param1")
	os.Setenv("SSM_PARAMETER_NAME_SLACK", "param2")
	os.Setenv("NEW_RELIC_API_TOKEN", "param4")

	defer os.Unsetenv("SSM_PARAMETER_NAME_NEW_RELIC")
	defer os.Unsetenv("SSM_PARAMETER_NAME_SLACK")
	defer os.Unsetenv("NEW_RELIC_API_TOKEN")

	_, err := validate.EnvValidate()

	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "SSM_PARAMETER_MESSAGE_SLACK"))
}

func TestEnvValidateFailNewRelicAPIToken(t *testing.T) {
	os.Setenv("SSM_PARAMETER_NAME_NEW_RELIC", "param1")
	os.Setenv("SSM_PARAMETER_NAME_SLACK", "param2")
	os.Setenv("SSM_PARAMETER_MESSAGE_SLACK", "param3")

	defer os.Unsetenv("SSM_PARAMETER_NAME_NEW_RELIC")
	defer os.Unsetenv("SSM_PARAMETER_NAME_SLACK")
	defer os.Unsetenv("SSM_PARAMETER_MESSAGE_SLACK")

	_, err := validate.EnvValidate()

	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "NEW_RELIC_API_TOKEN"))
}
