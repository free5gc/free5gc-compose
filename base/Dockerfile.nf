#
# Dockerfile responsible to compile specific NF from free5gc sources on the host
#

FROM free5gc/base AS my-base

ENV DEBIAN_FRONTEND=noninteractive
ARG F5GC_MODULE

# Copy sources outside GOPATH to avoid module path conflicts
WORKDIR /workspace/free5gc
COPY free5gc/ ./

# Build target NF
RUN make ${F5GC_MODULE}

# Alpine is used for debug purpose. You can use scratch for a smaller footprint.
FROM alpine:3.15

ARG F5GC_MODULE

WORKDIR /free5gc
RUN mkdir -p cert/ public

# Copy executables
COPY --from=my-base /workspace/free5gc/bin/${F5GC_MODULE} ./

# Copy configuration files (not used for now)
COPY --from=my-base /workspace/free5gc/config/* ./config/

# Copy default certificates (not used for now)
COPY --from=my-base /workspace/free5gc/cert/* ./cert/
