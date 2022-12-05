#!/bin/bash
TAG=${1-"lastest"}

make base

NF_LIST="nrf amf smf udr pcf udm nssf ausf n3iwf"

echo "Building docker images"
for NF in ${NF_LIST}; do
    cd nf_${NF}
    docker build --build-arg DEBUG_TOOLS=true -t free5gc/${NF}:${TAG} .
    docker push free5gc/${NF}:${TAG}
    cd -
done

cd webui
docker build --build-arg DEBUG_TOOLS=true -t free5gc/webconsole:${TAG} .
docker push free5gc/webconsole:${TAG}
cd -

cd ueransim
docker build --build-arg DEBUG_TOOLS=true -t free5gc/ueransim:latest .
docker image ls
docker push free5gc/ueransim:${TAG}