DOCKER_IMAGE_OWNER = 'free5gc'
DOCKER_IMAGE_NAME = 'base'
DOCKER_IMAGE_TAG = 'latest'
DEBUG_ENABLE = 'false'

.PHONY: base
all: base amf ausf nrf nssf pcf smf udm udr n3iwf upf webconsole

base:
	docker build -t ${DOCKER_IMAGE_OWNER}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ./base
	docker image ls ${DOCKER_IMAGE_OWNER}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}

smf: base
	docker build --build-arg F5GC_MODULE=smf --build-arg DEBUG_TOOLS=${DEBUG_ENABLE} -t ${DOCKER_IMAGE_OWNER}/smf:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
amf: base
	docker build --build-arg F5GC_MODULE=amf --build-arg DEBUG_TOOLS=${DEBUG_ENABLE} -t ${DOCKER_IMAGE_OWNER}/amf:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
upf: base
	docker build --build-arg F5GC_MODULE=upf --build-arg DEBUG_TOOLS=false -t ${DOCKER_IMAGE_OWNER}/upf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
udr: base
	docker build --build-arg F5GC_MODULE=udr --build-arg DEBUG_TOOLS=${DEBUG_ENABLE} -t ${DOCKER_IMAGE_OWNER}/udr:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
udm: base
	docker build --build-arg F5GC_MODULE=udm --build-arg DEBUG_TOOLS=${DEBUG_ENABLE} -t ${DOCKER_IMAGE_OWNER}/udm:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
nrf: base
	docker build --build-arg F5GC_MODULE=nrf --build-arg DEBUG_TOOLS=${DEBUG_ENABLE} -t ${DOCKER_IMAGE_OWNER}/nrf:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
nssf: base
	docker build --build-arg F5GC_MODULE=nssf --build-arg DEBUG_TOOLS=${DEBUG_ENABLE} -t ${DOCKER_IMAGE_OWNER}/nssf:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
n3iwf: base
	docker build --build-arg F5GC_MODULE=n3iwf --build-arg DEBUG_TOOLS=${DEBUG_ENABLE} -t ${DOCKER_IMAGE_OWNER}/n3iwf:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
pcf: base
	docker build --build-arg F5GC_MODULE=pcf --build-arg DEBUG_TOOLS=${DEBUG_ENABLE} -t ${DOCKER_IMAGE_OWNER}/pcf:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
ausf: base
	docker build --build-arg F5GC_MODULE=ausf --build-arg DEBUG_TOOLS=${DEBUG_ENABLE} -t ${DOCKER_IMAGE_OWNER}/ausf:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base

webconsole: base
	docker build -t ${DOCKER_IMAGE_OWNER}/webconsole-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf.webconsole ./base
