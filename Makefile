DOCKER_IMAGE_OWNER ?= free5gc
DOCKER_IMAGE_TAG ?= latest
MAKE_JOBS ?= 1

NF_MODULES := amf ausf bsf nrf nssf pcf scp smf udm udr n3iwf upf chf tngf nef
BASE_IMAGE_ALIASES := $(addsuffix -base,$(NF_MODULES)) webconsole-base

.PHONY: all base all-base $(NF_MODULES) webconsole

all: all-base

base:
	docker build -t $(DOCKER_IMAGE_OWNER)/base:$(DOCKER_IMAGE_TAG) ./base
	docker image ls $(DOCKER_IMAGE_OWNER)/base:$(DOCKER_IMAGE_TAG)

all-base: base
	docker build \
		--build-arg BASE_IMAGE=$(DOCKER_IMAGE_OWNER)/base:$(DOCKER_IMAGE_TAG) \
		--build-arg MAKE_JOBS=$(MAKE_JOBS) \
		-t $(DOCKER_IMAGE_OWNER)/all-base:$(DOCKER_IMAGE_TAG) \
		-f ./base/Dockerfile.all ./base
	@set -e; \
	for image in $(BASE_IMAGE_ALIASES); do \
		docker tag \
			$(DOCKER_IMAGE_OWNER)/all-base:$(DOCKER_IMAGE_TAG) \
			$(DOCKER_IMAGE_OWNER)/$$image:$(DOCKER_IMAGE_TAG); \
	done
	docker image ls $(DOCKER_IMAGE_OWNER)/all-base:$(DOCKER_IMAGE_TAG)

$(NF_MODULES): %: base
	docker build \
		--build-arg F5GC_MODULE=$@ \
		-t $(DOCKER_IMAGE_OWNER)/$@-base:$(DOCKER_IMAGE_TAG) \
		-f ./base/Dockerfile.nf ./base

webconsole: base
	docker build \
		-t $(DOCKER_IMAGE_OWNER)/webconsole-base:$(DOCKER_IMAGE_TAG) \
		-f ./base/Dockerfile.nf.webconsole ./base
