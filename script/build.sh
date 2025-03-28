#!/bin/bash
ARCH=${1-"x86_64"}

cd base
git clone --recursive -j `nproc` https://github.com/free5gc/free5gc.git

cd -

make all

echo "Building docker images for ${ARCH}..."

if [ ${ARCH} == "aarch64" ]; then
    docker compose -f docker-compose-build.yaml build --build-arg TARGET_ARCH=${ARCH}
else
    docker compose -f docker-compose-build.yaml build
fi;

