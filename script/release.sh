#!/bin/bash
TAG=${1-"latest"}

NF_LIST="nrf amf smf udr pcf udm nssf ausf n3iwf"

cd base
git clone --recursive -b v3.2.1 -j `nproc` https://github.com/free5gc/free5gc.git
cd -

make all
docker compose -f docker-compose-build.yaml build

for NF in ${NF_LIST}; do
    # If $TAG not equal to latest
    if [ "${TAG}" != "latest" ]; then
        docker tag free5gc/${NF}:latest free5gc/${NF}:${TAG}
    fi
    docker push free5gc/${NF}:${TAG}
done

if [ "${TAG}" != "latest" ]; then
    docker tag free5gc/webconsole:latest free5gc/webconsole:${TAG}
fi

docker push free5gc/webconsole:${TAG}
docker push free5gc/ueransim:${TAG}