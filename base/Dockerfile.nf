#
# Dockerfile responsible to compile specific NF from free5gc sources on the host
#

FROM free5gc/base as my-base

ENV DEBIAN_FRONTEND noninteractive
ARG F5GC_MODULE

# Get Free5GC
COPY free5gc/ $GOPATH/src/free5gc/

RUN cd $GOPATH/src/free5gc \
    && make ${F5GC_MODULE}

# Alpine is used for debug purpose. You can use scratch for a smaller footprint.
FROM alpine:3.15

WORKDIR /free5gc
RUN mkdir -p cert/ public

# Copy executables
COPY --from=my-base /go/src/free5gc/bin/${F5GC_MODULE} ./

# Copy configuration files (not used for now)
COPY --from=my-base /go/src/free5gc/config/* ./config/

# Copy default certificates (not used for now)
COPY --from=my-base /go/src/free5gc/cert/* ./cert/
