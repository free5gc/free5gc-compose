FROM golang:1.12.9-stretch AS builder

MAINTAINER free5GC <support@free5gc.org>

ENV GO111MODULE=off

RUN apt-get update
RUN apt-get -y install gcc cmake autoconf libtool pkg-config libmnl-dev libyaml-dev
RUN apt-get clean

# Get Free5GC
RUN cd $GOPATH/src \
    && git clone --recursive -b v3.0.2 -j 33 https://github.com/free5gc/free5gc.git \
    && cd free5gc \
    && git submodule update --init -j 33 \
    && ./install_env.sh -j `nproc`

# Build NF (AMF, AUSF, N3IWF, NRF, NSSF, PCF, SMF, UDM, UDR)
RUN cd $GOPATH/src/free5gc/src \
    && for d in * ; do if [ -f "$d/$d.go" ] ; then go build -o ../bin/"$d" -x "$d/$d.go" ; fi ; done ;

# Build UPF
#RUN go get -u -v "github.com/sirupsen/logrus"
RUN cd $GOPATH/src/free5gc/src/upf \
    && mkdir -p build \
    && cd build \
    && cmake .. \
    && make -j `nproc`

# Build WebUI
RUN cd $GOPATH/src/free5gc/webconsole \
    && go build -o webui -x server.go

# Alpine is used for debug purpose. You can use scratch for a smaller footprint.
FROM alpine

WORKDIR /free5gc
RUN mkdir -p config/ support/TLS/

# Copy executables
COPY --from=builder /go/src/free5gc/bin/* ./
COPY --from=builder /go/src/free5gc/src/upf/build/bin/* ./
COPY --from=builder /go/src/free5gc/webconsole/webui ./

# Copy linked libs
COPY --from=builder /go/src/free5gc/src/upf/build/libgtp5gnl/lib/libgtp5gnl.so.0 ./
COPY --from=builder /go/src/free5gc/src/upf/build/utlt_logger/liblogger.so ./

# Copy configuration files (not used for now)
COPY --from=builder /go/src/free5gc/config/* ./config/
COPY --from=builder /go/src/free5gc/src/upf/build/config/* ./config/

# Copy default certificates (not used for now)
COPY --from=builder /go/src/free5gc/support/TLS/* ./support/TLS/
