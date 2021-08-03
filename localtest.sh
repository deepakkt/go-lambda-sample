## TODO: Make this script idemopotent

# download lambda run time emulator
# uncomment this if needed
# this needs to run only once

#mkdir -p ~/.aws-lambda-rie \
#&& curl -Lo ~/.aws-lambda-rie/aws-lambda-rie \
#https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie \
#&& chmod +x ~/.aws-lambda-rie/aws-lambda-rie

# change this after local docker build
IMAGE_ID=da0fe572ebad

docker run -d -v ~/.aws-lambda-rie:/aws-lambda \
--env-file ~/.aws/env-file \
--entrypoint /aws-lambda/aws-lambda-rie  \
-p 9000:8080 ${IMAGE_ID} /main

