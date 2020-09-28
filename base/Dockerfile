FROM golang:1.14.4-stretch AS builder

LABEL maintainer="Free5GC <support@free5gc.org>"

RUN apt-get update
RUN apt-get -y install gcc cmake autoconf libtool pkg-config libmnl-dev libyaml-dev
RUN apt-get clean

# Get Free5GC
RUN cd $GOPATH/src \
    && git clone --recursive -b v3.0.4 -j 33 https://github.com/free5gc/free5gc.git \
    && cd free5gc \
    && go mod download

COPY Makefile $GOPATH/src/free5gc/Makefile

# Build NFs (AMF, AUSF, N3IWF, NRF, NSSF, PCF, SMF, UDM, UDR, UPF)
RUN cd $GOPATH/src/free5gc \
    && make all

# Build WebUI
RUN cd $GOPATH/src/free5gc/webconsole \
    && CGO_ENABLED=0 go build -a -installsuffix nocgo -o webui -x server.go

# Alpine is used for debug purpose. You can use scratch for a smaller footprint.
FROM alpine

WORKDIR /free5gc
RUN mkdir -p config/ support/TLS/ public

# Copy executables
COPY --from=builder /go/src/free5gc/bin/* ./
COPY --from=builder /go/src/free5gc/src/upf/build/bin/* ./
COPY --from=builder /go/src/free5gc/webconsole/webui ./

# Copy static files (webui frontend)
COPY --from=builder /go/src/free5gc/webconsole/public ./public

# Copy linked libs
COPY --from=builder /go/src/free5gc/src/upf/build/updk/src/third_party/libgtp5gnl/lib/libgtp5gnl.so.0 ./
COPY --from=builder /go/src/free5gc/src/upf/build/utlt_logger/liblogger.so ./

# Copy configuration files (not used for now)
COPY --from=builder /go/src/free5gc/config/* ./config/
COPY --from=builder /go/src/free5gc/src/upf/build/config/* ./config/

# Copy default certificates (not used for now)
COPY --from=builder /go/src/free5gc/support/TLS/* ./support/TLS/
