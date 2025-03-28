#!/bin/bash
ARCH=${1-"x86_64"}
TAG=${2-"latest"}

NF_LIST="nrf amf smf udr pcf udm nssf ausf n3iwf upf chf tngf nef"

cd base

if [ 'xlatest' == "x$TAG" ]; then
    git clone --recursive -j `nproc` https://github.com/free5gc/free5gc.git
else
    TAG=`echo "$TAG" | sed -e "s/refs\/tags\///g"`
    git clone --recursive -b ${TAG} -j `nproc` https://github.com/free5gc/free5gc.git
fi;

cd -

make all
if [ ${ARCH} == "aarch64" ]; then
    docker compose -f docker-compose-build.yaml build --build-arg TARGET_ARCH=${ARCH}
else
    docker compose -f docker-compose-build.yaml build
fi;

for NF in ${NF_LIST}; do
    docker tag free5gc-compose_free5gc-${NF}:latest free5gc/${NF}:${TAG}
    docker push free5gc/${NF}:${TAG}
done


docker tag free5gc-compose_free5gc-webui:latest free5gc/webui:${TAG}
docker tag free5gc-compose_ueransim:latest free5gc/ueransim:${TAG}
docker tag free5gc-compose_n3iwue:latest free5gc/n3iwue:${TAG}

docker push free5gc/webui:${TAG}
docker push free5gc/ueransim:${TAG}
docker push free5gc/n3iwue:${TAG}
