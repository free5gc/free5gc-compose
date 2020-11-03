FROM golang:1.14.4-stretch AS builder

LABEL maintainer="My5Gcore <support@labora.inf.ufg.br>"

ENV MY5GC_BRANCH master
ENV UE_BRANCH fix_release_3.0.4

RUN apt-get update
RUN apt-get -y install gcc cmake autoconf libtool pkg-config libmnl-dev libyaml-dev ssh
RUN apt-get clean

# Get My5Gcore
RUN cd $GOPATH/src \
    && git clone --recursive -b ${MY5GC_BRANCH} -j 33 https://github.com/LABORA-INF-UFG/my5Gcore.git \
    && cd my5Gcore \
    && go mod download

## ssh key configuration
RUN mkdir /root/.ssh
COPY .ssh/id_rsa  /root/.ssh/id_rsa
RUN chown 400 /root/.ssh/id_rsa
#RUN ssh-add /root/.ssh/id_rsa.pub

RUN touch /root/.ssh/known_hosts
RUN ssh-keyscan github.com >> /root/.ssh/known_hosts

# Get UE Code
RUN cd $GOPATH/src/my5Gcore/src \
    && git clone -b ${UE_BRANCH} git@github.com:LABORA-INF-UFG/UE-IoT-non3GPP.git ue \
    && go get -u github.com/calee0219/fatal

# Build UE
RUN cd $GOPATH/src/my5Gcore/src \
    && CGO_ENABLED=0 go build -a -installsuffix nocgo -o ../bin/ue -x ue/ue.go

# Alpine is used for debug purpose. You can use scratch for a smaller footprint.
FROM alpine

WORKDIR /my5Gcore
RUN mkdir -p config/ support/TLS/

# Copy executables
COPY --from=builder /go/src/my5Gcore/bin/* ./

# Copy configuration files (not used for now)
COPY --from=builder /go/src/my5Gcore/config/* ./config/

# Copy default certificates (not used for now)
COPY --from=builder /go/src/my5Gcore/support/TLS/* ./support/TLS/
