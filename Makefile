DOCKER_IMAGE_NAME = 'free5gc-base-v3'
DOCKER_IMAGE_TAG = 'latest'

.PHONY: base
base:
	docker build -t ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ./base
	docker image ls ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}
