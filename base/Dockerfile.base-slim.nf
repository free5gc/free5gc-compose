#
# Dockerfile responsible to build a specific free5gc NF from the sources located at
# the host
#
# Prior to this build invocation you must clone free5gc sources into this folder (i.e.
# 'base') on the host
# E.g.:
# git clone --recursive -b v3.2.1 -j `nproc` https://github.com/free5gc/free5gc.git

FROM free5gc/base-slim as my-base

ENV DEBIAN_FRONTEND noninteractive
ARG F5GC_MODULE

# Get Free5GC
COPY free5gc/ $GOPATH/src/free5gc/

RUN cd $GOPATH/src/free5gc \
    && make all

# Alpine is used for debug purpose. You can use scratch for a smaller footprint.
FROM alpine:3.15

WORKDIR /free5gc
RUN mkdir -p config/TLS/ public

# Copy executables
COPY --from=my-base /go/src/free5gc/bin/${F5GC_MODULE} ./
COPY --from=my-base /go/src/free5gc/webconsole/bin/webconsole ./webui

# Copy configuration files (not used for now)
COPY --from=my-base /go/src/free5gc/config/* ./config/

# Copy default certificates (not used for now)
COPY --from=my-base /go/src/free5gc/config/TLS/* ./config/TLS/
