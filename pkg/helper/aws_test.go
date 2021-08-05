package helper_test

import (
	"deployment-notifications/pkg/helper"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)


func TestAWSDefaultRegionLambda(t *testing.T) {
	os.Setenv("AWS_REGION", "ap-south-1")
	defer os.Unsetenv("AWS_REGION")

	myRegion := helper.GetAwsDefaultRegion()
	assert.Equal(t, "ap-south-1", myRegion)
}


func TestAWSDefaultRegionLambdaLocal(t *testing.T) {
	os.Unsetenv("AWS_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "ap-south-1")
	defer os.Unsetenv("AWS_DEFAULT_REGION")

	myRegion := helper.GetAwsDefaultRegion()
	assert.Equal(t, "ap-south-1", myRegion)
}


func TestAWSDefaultRegionAllMissing(t *testing.T) {
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_DEFAULT_REGION")

	myRegion := helper.GetAwsDefaultRegion()
	assert.Equal(t, "", myRegion)
}


func TestFetchServiceFromARN(t *testing.T) {
	inputARN := "arn:aws:ecs:us-west-2:111122223333:service/shure-content-api"
	serviceName, err := helper.GetServiceNameFromARN(inputARN)
	assert.Equal(t, "shure-content-api", serviceName)
	assert.Equal(t, nil, err)

	inputARN = "arn:aws:ecs:us-west-2:111122223333:services/shure-content-api"
	serviceName, err = helper.GetServiceNameFromARN(inputARN)
	assert.NotNil(t, err)
}


func TestParseEventDetailSuccess(t *testing.T) {
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

	if err != nil {
		t.Log("We could not unmarshal the sample event. This must not happen!")
		t.FailNow()
	}

	parseOutput, err := helper.ParseEventDetails(cloudwatchEvent)

	assert.Equal(t, "SERVICE_DEPLOYMENT_COMPLETED", parseOutput["eventName"])
	assert.Equal(t, "ecs-svc/123", parseOutput["deploymentId"])
	assert.Equal(t, "2020-05-23T11:11:11Z", parseOutput["updatedAt"])
	assert.Equal(t, "ECS deployment deploymentId in progress.", parseOutput["reason"])
	assert.Nil(t, err)
}


func TestParseEventDetailMissingEventName(t *testing.T) {
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
        "deploymentId": "ecs-svc/123",                                    
        "updatedAt": "2020-05-23T11:11:11Z",                              
        "reason": "ECS deployment deploymentId in progress."              
   }                                                                      
}                                                                         
`

	var cloudwatchEvent events.CloudWatchEvent
	err := json.Unmarshal([]byte(sampleEvent), &cloudwatchEvent)

	if err != nil {
		t.Log("We could not unmarshal the sample event. This must not happen!")
		t.FailNow()
	}

	_, err = helper.ParseEventDetails(cloudwatchEvent)

	assert.NotNil(t, err)
}


func TestParseEventDetailMissingDeploymentID(t *testing.T) {
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
        "updatedAt": "2020-05-23T11:11:11Z",                              
        "reason": "ECS deployment deploymentId in progress."              
   }                                                                      
}                                                                         
`

	var cloudwatchEvent events.CloudWatchEvent
	err := json.Unmarshal([]byte(sampleEvent), &cloudwatchEvent)

	if err != nil {
		t.Log("We could not unmarshal the sample event. This must not happen!")
		t.FailNow()
	}

	_, err = helper.ParseEventDetails(cloudwatchEvent)

	assert.NotNil(t, err)
}


func TestParseEventDetailMissingTimestamp(t *testing.T) {
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
        "reason": "ECS deployment deploymentId in progress."              
   }                                                                      
}                                                                         
`

	var cloudwatchEvent events.CloudWatchEvent
	err := json.Unmarshal([]byte(sampleEvent), &cloudwatchEvent)

	if err != nil {
		t.Log("We could not unmarshal the sample event. This must not happen!")
		t.FailNow()
	}

	_, err = helper.ParseEventDetails(cloudwatchEvent)

	assert.NotNil(t, err)
}