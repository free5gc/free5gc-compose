FROM golang:1.14.4-stretch AS builder

LABEL maintainer="My5Gcore <support@labora.inf.ufg.br>"

RUN apt-get update
RUN apt-get -y install gcc cmake autoconf libtool pkg-config libmnl-dev libyaml-dev curl
RUN apt-get clean

# Get My5Gcore
RUN cd $GOPATH/src \
    && git clone -b vSBrT2020 --recursive -j 33 https://github.com/LABORA-INF-UFG/my5Gcore.git \
    && cd my5Gcore \
    && go mod download

# Build NF (AMF, AUSF, N3IWF, NRF, NSSF, PCF, SMF, UDM, UDR)
RUN cd $GOPATH/src/my5Gcore/src \
    && for d in * ; do if [ -f "$d/$d.go" ] ; then CGO_ENABLED=0 go build -a -installsuffix nocgo -o ../bin/"$d" -x "$d/$d.go" ; fi ; done ;

# Build UPF
#RUN go get -u -v "github.com/sirupsen/logrus"
RUN cd $GOPATH/src/my5Gcore/src/upf \
    && mkdir -p build \
    && cd build \
    && cmake .. \
    && make -j `nproc`

#INSTALL YARN 
#RUN apt update \
#    && apt -y  install apt-transport-https ca-certificates \ 
#    && curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg |  apt-key add - \
#    && echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list \
#    && apt-get update \
#    && apt-get -y install yarn \
#    && apt-get clean 

# Build WebUI
RUN cd $GOPATH/src/my5Gcore/webconsole \
# && cd frontend \
# && yarn install \
# && yarn build \
# && rm -rf ../public \
# && cp -R build ../public \
# && cd ..\
 && CGO_ENABLED=0 go build -a -installsuffix nocgo -o webui -x server.go


# Alpine is used for debug purpose. You can use scratch for a smaller footprint.
FROM alpine

WORKDIR /my5Gcore
RUN mkdir -p config/ support/TLS/ public

# Copy executables
COPY --from=builder /go/src/my5Gcore/bin/* ./
COPY --from=builder /go/src/my5Gcore/src/upf/build/bin/* ./
COPY --from=builder /go/src/my5Gcore/webconsole/webui ./

# Copy static files (webui frontend)
COPY --from=builder /go/src/my5Gcore/webconsole/public ./public

# Copy linked libs
COPY --from=builder /go/src/my5Gcore/src/upf/build/updk/src/third_party/libgtp5gnl/lib/libgtp5gnl.so.0 ./
COPY --from=builder /go/src/my5Gcore/src/upf/build/utlt_logger/liblogger.so ./

# Copy configuration files (not used for now)
COPY --from=builder /go/src/my5Gcore/config/* ./config/
COPY --from=builder /go/src/my5Gcore/src/upf/build/config/* ./config/

# Copy default certificates (not used for now)
COPY --from=builder /go/src/my5Gcore/support/TLS/* ./support/TLS/
