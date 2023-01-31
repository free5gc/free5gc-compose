DOCKER_IMAGE_OWNER = 'free5gc'
DOCKER_IMAGE_NAME = 'base'
DOCKER_IMAGE_NAME_SLIM = 'base-slim'
DOCKER_IMAGE_TAG = 'latest'

.PHONY: base
all: base-slim amf ausf nrf nssf pcf smf udm udr n3iwf upf webconsole

base:
	docker build -t ${DOCKER_IMAGE_OWNER}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ./base
	docker image ls ${DOCKER_IMAGE_OWNER}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}

base-slim:
	docker build -t ${DOCKER_IMAGE_OWNER}/${DOCKER_IMAGE_NAME_SLIM}:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.base-slim ./base
	docker image ls ${DOCKER_IMAGE_OWNER}/${DOCKER_IMAGE_NAME_SLIM}:${DOCKER_IMAGE_TAG}

smf: base-slim
	docker build --build-arg F5GC_MODULE=smf -t ${DOCKER_IMAGE_OWNER}/smf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.base-slim.nf ./base
amf: base-slim
	docker build --build-arg F5GC_MODULE=amf -t ${DOCKER_IMAGE_OWNER}/amf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.base-slim.nf ./base
upf: base-slim
	docker build --build-arg F5GC_MODULE=upf -t ${DOCKER_IMAGE_OWNER}/upf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.base-slim.nf ./base
udr: base-slim
	docker build --build-arg F5GC_MODULE=udr -t ${DOCKER_IMAGE_OWNER}/udr-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.base-slim.nf ./base
udm: base-slim
	docker build --build-arg F5GC_MODULE=udm -t ${DOCKER_IMAGE_OWNER}/udm-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.base-slim.nf ./base
nrf: base-slim
	docker build --build-arg F5GC_MODULE=nrf -t ${DOCKER_IMAGE_OWNER}/nrf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.base-slim.nf ./base
nssf: base-slim
	docker build --build-arg F5GC_MODULE=nssf -t ${DOCKER_IMAGE_OWNER}/nssf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.base-slim.nf ./base
n3iwf: base-slim
	docker build --build-arg F5GC_MODULE=n3iwf -t ${DOCKER_IMAGE_OWNER}/n3iwf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.base-slim.nf ./base
pcf: base-slim
	docker build --build-arg F5GC_MODULE=pcf -t ${DOCKER_IMAGE_OWNER}/pcf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.base-slim.nf ./base
ausf: base-slim
	docker build --build-arg F5GC_MODULE=ausf -t ${DOCKER_IMAGE_OWNER}/ausf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.base-slim.nf ./base

webconsole: base-slim
	docker build -t ${DOCKER_IMAGE_OWNER}/webconsole-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.base-slim.webconsole ./base
