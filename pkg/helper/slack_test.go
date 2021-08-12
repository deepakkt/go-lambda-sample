package helper_test

import (
	"deployment-notifications/pkg/helper"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSlackMappingDecode(t *testing.T) {
	mapTemplate := `
{
	"input1": ["a", "b", "c"],
	"input2": ["d", "e", "f"]
}
`

	mapOutput, err := helper.DecodeSlackMapping(mapTemplate)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, mapOutput["input1"])
	assert.Equal(t, []string{"d", "e", "f"}, mapOutput["input2"])

	mapTemplate = `
{
	"input1": ["a", "b", "c"]
	"input2": ["d", "e", "f"]
}
`

	mapOutput, err = helper.DecodeSlackMapping(mapTemplate)
	assert.NotNil(t, err)
}

func TestSlackDefaultWebhook(t *testing.T) {
	mapTemplate := `
{
	"default-service": ["http://www.google.com"],
	"input2": ["d", "e", "f"]
}
`
	mapOutput, _ := helper.DecodeSlackMapping(mapTemplate)
	webhook := helper.GetDefaultWebhook(mapOutput)
	assert.Equal(t, "http://www.google.com", webhook)

	mapTemplate = `
{
	"default-services": ["http://www.google.com"],
	"input2": ["d", "e", "f"]
}
`
	mapOutput, _ = helper.DecodeSlackMapping(mapTemplate)
	webhook = helper.GetDefaultWebhook(mapOutput)
	assert.Equal(t, "", webhook)

}

func TestSlackNotificationStructParse(t *testing.T) {
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

	slackStruct := helper.GenerateSlackNotificationStruct(cloudwatchEvent)

	assert.Equal(t, "shure-content-api", slackStruct.ServiceName)
	assert.Equal(t, "ecs-svc/123", slackStruct.DeploymentRevision)
	assert.Equal(t, "ddca6449-b258-46c0-8653-e0e3a6EXAMPLE", slackStruct.AWSReference)
	assert.Equal(t, "111122223333", slackStruct.AWSAccount)
	assert.Equal(t, "us-west-2", slackStruct.AWSRegion)
	assert.Equal(t, "2020-05-23T11:11:11Z", slackStruct.DeploymentTimestamp)
	assert.Equal(t, "ECS deployment deploymentId in progress.", slackStruct.DeploymentDescription)
}

func TestSlackPayloadDequote(t *testing.T) {
	sampleMessage := `{
   "attachments":[
      {
         "fallback":"New deployment notification for <backquote><varbegin>.ServiceName<varend><backquote>",
         "pretext":"New deployment notification for <backquote><varbegin>.ServiceName<varend><backquote>",
         "color":"#D00000",
         "fields":[
            {
               "title":"Service",
               "value":"<varbegin>.ServiceName<varend>",
               "short":false
            },		 
            {
               "title":"Revision",
               "value":"<varbegin>.DeploymentRevision<varend>",
               "short":false
            },
            {
               "title":"AWS Reference",
               "value":"<varbegin>.AWSReference<varend>",
               "short":false
            },			
            {
               "title":"AWS Account",
               "value":"<varbegin>.AWSAccount<varend>",
               "short":false
            },			
            {
               "title":"AWS Region",
               "value":"<varbegin>.AWSRegion<varend>",
               "short":false
            },						
            {
               "title":"Description",
               "value":"<varbegin>.DeploymentDescription<varend>",
               "short":false
            },									
            {
               "title":"Timestamp",
               "value":"<varbegin>.DeploymentTimestamp<varend>",
               "short":false
            },						
         ]
      }	  
   ]
}`

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

	slackStruct := helper.GenerateSlackNotificationStruct(cloudwatchEvent)

	parsedMessage, err := helper.GeneratePayload(sampleMessage, slackStruct, true)
	assert.Nil(t, err)

	expectedOutput := `{
   "attachments":[
      {
         "fallback":"New deployment notification for ` + "`" + `shure-content-api` + "`" + `",
         "pretext":"New deployment notification for ` + "`" + `shure-content-api` + "`" + `",
         "color":"#D00000",
         "fields":[
            {
               "title":"Service",
               "value":"shure-content-api",
               "short":false
            },		 
            {
               "title":"Revision",
               "value":"ecs-svc/123",
               "short":false
            },
            {
               "title":"AWS Reference",
               "value":"ddca6449-b258-46c0-8653-e0e3a6EXAMPLE",
               "short":false
            },			
            {
               "title":"AWS Account",
               "value":"111122223333",
               "short":false
            },			
            {
               "title":"AWS Region",
               "value":"us-west-2",
               "short":false
            },						
            {
               "title":"Description",
               "value":"ECS deployment deploymentId in progress.",
               "short":false
            },									
            {
               "title":"Timestamp",
               "value":"2020-05-23T11:11:11Z",
               "short":false
            },						
         ]
      }	  
   ]
}`

	assert.Equal(t, expectedOutput, parsedMessage)
}

func TestSlackPayloadNoDequote(t *testing.T) {
	sampleMessage := `{
   "attachments":[
      {
         "fallback":"New deployment notification for <backquote><varbegin>.ServiceName<varend><backquote>",
         "pretext":"New deployment notification for <backquote><varbegin>.ServiceName<varend><backquote>",
         "color":"#D00000",
         "fields":[
            {
               "title":"Service",
               "value":"<varbegin>.ServiceName<varend>",
               "short":false
            },		 
            {
               "title":"Revision",
               "value":"<varbegin>.DeploymentRevision<varend>",
               "short":false
            },
            {
               "title":"AWS Reference",
               "value":"<varbegin>.AWSReference<varend>",
               "short":false
            },			
            {
               "title":"AWS Account",
               "value":"<varbegin>.AWSAccount<varend>",
               "short":false
            },			
            {
               "title":"AWS Region",
               "value":"<varbegin>.AWSRegion<varend>",
               "short":false
            },						
            {
               "title":"Description",
               "value":"<varbegin>.DeploymentDescription<varend>",
               "short":false
            },									
            {
               "title":"Timestamp",
               "value":"<varbegin>.DeploymentTimestamp<varend>",
               "short":false
            },						
         ]
      }	  
   ]
}`

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

	slackStruct := helper.GenerateSlackNotificationStruct(cloudwatchEvent)

	parsedMessage, err := helper.GeneratePayload(sampleMessage, slackStruct, false)
	assert.Nil(t, err)

	expectedOutput := `{
   "attachments":[
      {
         "fallback":"New deployment notification for ` + "<backquote>" + `shure-content-api` + "<backquote>" + `",
         "pretext":"New deployment notification for ` + "<backquote>" + `shure-content-api` + "<backquote>" + `",
         "color":"#D00000",
         "fields":[
            {
               "title":"Service",
               "value":"shure-content-api",
               "short":false
            },		 
            {
               "title":"Revision",
               "value":"ecs-svc/123",
               "short":false
            },
            {
               "title":"AWS Reference",
               "value":"ddca6449-b258-46c0-8653-e0e3a6EXAMPLE",
               "short":false
            },			
            {
               "title":"AWS Account",
               "value":"111122223333",
               "short":false
            },			
            {
               "title":"AWS Region",
               "value":"us-west-2",
               "short":false
            },						
            {
               "title":"Description",
               "value":"ECS deployment deploymentId in progress.",
               "short":false
            },									
            {
               "title":"Timestamp",
               "value":"2020-05-23T11:11:11Z",
               "short":false
            },						
         ]
      }	  
   ]
}`

	assert.Equal(t, expectedOutput, parsedMessage)
}
