DOCKER_IMAGE_OWNER = 'free5gc'
DOCKER_IMAGE_NAME = 'base'
DOCKER_IMAGE_TAG = 'latest'

.PHONY: base
all: base amf ausf nrf nssf pcf smf udm udr n3iwf upf chf webconsole

base:
	docker build -t ${DOCKER_IMAGE_OWNER}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ./base
	docker image ls ${DOCKER_IMAGE_OWNER}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}

smf: base
	docker build --build-arg F5GC_MODULE=smf -t ${DOCKER_IMAGE_OWNER}/smf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
amf: base
	docker build --build-arg F5GC_MODULE=amf -t ${DOCKER_IMAGE_OWNER}/amf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
upf: base
	docker build --build-arg F5GC_MODULE=upf -t ${DOCKER_IMAGE_OWNER}/upf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
udr: base
	docker build --build-arg F5GC_MODULE=udr -t ${DOCKER_IMAGE_OWNER}/udr-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
udm: base
	docker build --build-arg F5GC_MODULE=udm -t ${DOCKER_IMAGE_OWNER}/udm-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
nrf: base
	docker build --build-arg F5GC_MODULE=nrf -t ${DOCKER_IMAGE_OWNER}/nrf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
nssf: base
	docker build --build-arg F5GC_MODULE=nssf -t ${DOCKER_IMAGE_OWNER}/nssf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
n3iwf: base
	docker build --build-arg F5GC_MODULE=n3iwf -t ${DOCKER_IMAGE_OWNER}/n3iwf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
pcf: base
	docker build --build-arg F5GC_MODULE=pcf -t ${DOCKER_IMAGE_OWNER}/pcf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
ausf: base
	docker build --build-arg F5GC_MODULE=ausf -t ${DOCKER_IMAGE_OWNER}/ausf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base
chf: base
	docker build --build-arg F5GC_MODULE=chf -t ${DOCKER_IMAGE_OWNER}/chf-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf ./base

webconsole: base
	docker build -t ${DOCKER_IMAGE_OWNER}/webconsole-base:${DOCKER_IMAGE_TAG} -f ./base/Dockerfile.nf.webconsole ./base
