DOCKER_IMAGE_OWNER = 'free5gc'
DOCKER_IMAGE_NAME = 'base'
DOCKER_IMAGE_NAME_SLIM = 'base-slim'
DOCKER_IMAGE_TAG = 'latest'

.PHONY: base
all: base-slim

base:
	docker build -t ${DOCKER_IMAGE_OWNER}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ./base
	docker image ls ${DOCKER_IMAGE_OWNER}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}

base-slim:
	docker build -t ${DOCKER_IMAGE_OWNER}/${DOCKER_IMAGE_NAME_SLIM}:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.base-slim ./base
	docker build -t ${DOCKER_IMAGE_OWNER}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.base-slim.nf ./base
	docker image ls ${DOCKER_IMAGE_OWNER}/${DOCKER_IMAGE_NAME_SLIM}:${DOCKER_IMAGE_TAG}