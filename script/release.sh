#!/usr/bin/env bash
set -euo pipefail

ARCH="${1:-amd64}"
TAG="${2:-latest}"
FREE5GC_COMMIT="${3:-${FREE5GC_COMMIT:-}}"
MAKE_JOBS="${MAKE_JOBS:-1}"
CLONE_JOBS="${CLONE_JOBS:-1}"

NF_IMAGES=(nrf amf bsf scp smf udr pcf udm nssf ausf n3iwf upf chf tngf nef webui)
ADDITIONAL_IMAGES=(ueransim n3iwue)
IMAGES=("${NF_IMAGES[@]}" "${ADDITIONAL_IMAGES[@]}")
BUILD_SERVICES=(
    free5gc-nrf free5gc-amf free5gc-bsf free5gc-scp free5gc-smf free5gc-udr free5gc-pcf
    free5gc-udm free5gc-nssf free5gc-ausf free5gc-n3iwf free5gc-upf
    free5gc-chf free5gc-tngf free5gc-nef free5gc-webui
    ueransim n3iwue
)

# Docker Hub tags use OCI architecture names; TARGET_ARCH follows Kitware's
# CMake installer filename convention used by ueransim/Dockerfile.
case "${ARCH}" in
    amd64)
        TARGET_ARCH=x86_64
        ;;
    arm64)
        TARGET_ARCH=aarch64
        ;;
    *)
        echo "Unsupported architecture: ${ARCH}" >&2
        exit 1
        ;;
esac

TAG="${TAG#refs/tags/}"
ARCH_TAG="${TAG}-${ARCH}"

if [[ -n "${FREE5GC_COMMIT}" ]]; then
    git init -q base/free5gc
    git -C base/free5gc remote add origin https://github.com/free5gc/free5gc.git
    git -C base/free5gc fetch --depth 1 origin "${FREE5GC_COMMIT}"
    git -C base/free5gc checkout --detach FETCH_HEAD
    git -C base/free5gc submodule update --init --recursive --jobs "${CLONE_JOBS}"
else
    clone_args=(--recursive --jobs "${CLONE_JOBS}")
    if [[ "${TAG}" != "latest" ]]; then
        clone_args+=(--branch "${TAG}")
    fi
    git clone "${clone_args[@]}" https://github.com/free5gc/free5gc.git base/free5gc
fi

make all MAKE_JOBS="${MAKE_JOBS}"

echo "Building ${ARCH} runtime images with tag ${ARCH_TAG}..."
for service in "${BUILD_SERVICES[@]}"; do
    FREE5GC_IMAGE_TAG="${ARCH_TAG}" \
    FREE5GC_ADDITIONAL_IMAGE_TAG="${ARCH_TAG}" \
    docker compose -f docker-compose-build.yaml build \
        --build-arg TARGET_ARCH="${TARGET_ARCH}" \
        "${service}"
done

for image in "${IMAGES[@]}"; do
    source_ref="free5gc/${image}:${ARCH_TAG}"

    docker image inspect "${source_ref}" >/dev/null
    docker push "${source_ref}"
done
