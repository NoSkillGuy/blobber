#!/usr/bin/env bash
ZCB="blobber"
ZCV="validator"
read -p "Provide the docker image tag name: " TAG
read -p "Provide the github organisation name[default:-0chaintest]: " organisation
echo "${organisation:-0chaintest}/${ZCB}:$TAG"

REGISTRY_BLOBBER="${organisation:-0chaintest}/${ZCB}"
REGISTRY_VALIDATOR="${organisation:-0chaintest}/${ZCV}"
if [[ $? -ne 0 ]]; then
  docker login
fi

if [ -n "$TAG" ]; then
echo " $TAG is the tage name provided"
echo -e "${ZCB}: Docker image build is started.. \n"
sudo docker build -t ${REGISTRY_BLOBBER}:${TAG} -f docker.local/Dockerfile .
sudo docker pull ${REGISTRY_BLOBBER}:latest
sudo docker tag ${REGISTRY_BLOBBER}:latest ${REGISTRY_BLOBBER}:stable_latest
echo "Re-tagging the remote latest tag to stable_latest"
sudo docker push ${REGISTRY_BLOBBER}:stable_latest
sudo docker tag ${REGISTRY_BLOBBER}:${TAG} ${REGISTRY_BLOBBER}:latest
echo "Pushing the new latest tag to dockerhub"
sudo docker push ${REGISTRY_BLOBBER}:latest
echo "Pushing the new tag to dockerhub tagged as ${REGISTRY_BLOBBER}:${TAG}"
sudo docker push ${REGISTRY_BLOBBER}:${TAG}

echo -e "${ZCB}: Docker image build is started.. \n"
sudo docker build -t ${REGISTRY_VALIDATOR}:${TAG} -f docker.local/build.validator/Dockerfile .
sudo docker pull ${REGISTRY_VALIDATOR}:latest
sudo docker tag ${REGISTRY_VALIDATOR}:latest ${REGISTRY_VALIDATOR}:stable_latest
echo "Re-tagging the remote latest tag to stable_latest"
sudo docker push ${REGISTRY_VALIDATOR}:stable_latest
sudo docker tag ${REGISTRY_VALIDATOR}:${TAG} ${REGISTRY_VALIDATOR}:latest
echo "Pushing the new latest tag to dockerhub"
sudo docker push ${REGISTRY_VALIDATOR}:latest
echo "Pushing the new tag to dockerhub tagged as ${REGISTRY_VALIDATOR}:${TAG}"
sudo docker push ${REGISTRY_VALIDATOR}:${TAG}

fi
