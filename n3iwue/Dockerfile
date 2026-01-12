FROM golang:1.25.5-trixie AS builder

LABEL maintainer="Free5GC <support@free5gc.org>"

ENV DEBIAN_FRONTEND=noninteractive

# Install dependencies
RUN apt-get update \
    && apt-get install libsctp-dev lksctp-tools iproute2 -y \
    && apt-get install make -y \
    && git clone https://github.com/free5gc/n3iwue.git \
    && cd ./n3iwue \
    && make 

FROM bitnami/minideb:trixie

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install libsctp-dev lksctp-tools iproute2 iputils-ping procps psmisc tcpdump sudo -y \
    && apt-get clean

WORKDIR /n3iwue
RUN mkdir -p config/
#Install runtime dependencies + ping

COPY --from=builder go/n3iwue/n3iwue .
COPY --from=builder go/n3iwue/run.sh .
COPY --from=builder go/n3iwue/clean.sh .

VOLUME [ "/n3iwue/config" ]
