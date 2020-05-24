From golang:1.12.9-stretch

MAINTAINER free5GC <support@free5gc.org>

ENV GO111MODULE=off

RUN apt-get update
RUN apt-get -y install gcc cmake autoconf libtool pkg-config libmnl-dev libyaml-dev
RUN apt-get -y install netcat tcpdump iproute2 netbase
RUN apt-get clean

# Get Free5GC
RUN cd $GOPATH/src \
    && git clone https://github.com/free5gc/free5gc.git \
    && cd free5gc \
    && git submodule update --init --jobs `nproc` \
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

WORKDIR $GOPATH/src/free5gc/bin

