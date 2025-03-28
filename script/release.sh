#!/bin/bash
ARCH=${1-"x86_64"}
TAG=${2-"latest"}

NF_LIST="nrf amf smf udr pcf udm nssf ausf n3iwf upf chf tngf nef webui"
ADDITIONAL_IMAGES="ueransim n3iwue"

cd base

if [ "xlatest" == "x$TAG" ]; then
    git clone --recursive -j "$(nproc)" https://github.com/free5gc/free5gc.git
else
    TAG=$(echo "$TAG" | sed -e "s/refs\/tags\///g")
    git clone --recursive -b "${TAG}" -j "$(nproc)" https://github.com/free5gc/free5gc.git
fi

cd -

make all

# Build images for the specified architecture
if [ "$ARCH" == "aarch64" ]; then
    docker compose -f docker-compose-build.yaml build --build-arg TARGET_ARCH="$ARCH"
else
    docker compose -f docker-compose-build.yaml build
fi

# Tag and push images for each network function
for IMAGE in $NF_LIST; do
    docker tag "free5gc-compose_free5gc-${IMAGE}:latest" "free5gc/${IMAGE}:${TAG}-${ARCH}"
    docker push "free5gc/${IMAGE}:${TAG}-${ARCH}"
done

for IMAGE in $ADDITIONAL_IMAGES; do
    docker tag "free5gc-compose_${IMAGE}:latest" "free5gc/${IMAGE}:${TAG}-${ARCH}"
    docker push "free5gc/${IMAGE}:${TAG}-${ARCH}"
done

# Wait for the images to be pushed
sleep 60

# Create and push multi-architecture manifests
for IMAGE in $NF_LIST $ADDITIONAL_IMAGES; do
    docker manifest create "free5gc/${IMAGE}:${TAG}" \
        "free5gc/${IMAGE}:${TAG}-x86_64" \
        "free5gc/${IMAGE}:${TAG}-aarch64"
    docker manifest push "free5gc/${IMAGE}:${TAG}"
done