
mkdir -p ~/.aws-lambda-rie \
&& curl -Lo ~/.aws-lambda-rie/aws-lambda-rie \
https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie \
&& chmod +x ~/.aws-lambda-rie/aws-lambda-rie

GIT_LAST_COMMIT=$(git rev-parse --short HEAD)

docker build -t deployment-notifications:${GIT_LAST_COMMIT} .

docker run -d -v ~/.aws-lambda-rie:/aws-lambda \
--env-file ./env-file \
--entrypoint /aws-lambda/aws-lambda-rie  \
-p 9000:8080 deployment-notifications:${GIT_LAST_COMMIT} /main

