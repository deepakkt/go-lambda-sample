#!/usr/bin/env bash
## Script to execute the lambda locally

## INSTALL AWS Lambda Runtime Simulator
mkdir -p ~/.aws-lambda-rie \
&& curl -Lo ~/.aws-lambda-rie/aws-lambda-rie \
https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie \
&& chmod +x ~/.aws-lambda-rie/aws-lambda-rie

# Fetch last commit from local Git repo
GIT_LAST_COMMIT=$(git rev-parse --short HEAD)

docker build -t deployment-notifications:"${GIT_LAST_COMMIT}" .

## Execute image
## Note: Update env-file with live AWS creds from SSO session
## change other params as suitable for testing
docker run -d -v ~/.aws-lambda-rie:/aws-lambda \
--env-file ./local-test/env-file \
--entrypoint /aws-lambda/aws-lambda-rie  \
-p 9000:8080 deployment-notifications:"${GIT_LAST_COMMIT}" /main

## Sample local execution (this is like a lambda execution)
curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d @local-test/ecsevent.txt

