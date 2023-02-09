DOCKER_IMAGE_OWNER = 'free5gc'
DOCKER_IMAGE_NAME = 'base'
DOCKER_IMAGE_NAME_SLIM = 'base-slim'
DOCKER_IMAGE_TAG = 'latest'

.PHONY: base
all: base

base:
	docker build -t ${DOCKER_IMAGE_OWNER}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ./base
	docker image ls ${DOCKER_IMAGE_OWNER}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}