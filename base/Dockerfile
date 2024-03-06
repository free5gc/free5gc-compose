#
# Dockerfile responsible to create only the base image without compiling free5gc sources
#

FROM golang:1.18.10-bullseye AS builder

LABEL maintainer="Free5GC <support@free5gc.org>"

ENV DEBIAN_FRONTEND noninteractive

# Install dependencies
RUN apt-get update \
    && apt-get -y install gcc cmake autoconf libtool pkg-config libmnl-dev libyaml-dev apt-transport-https ca-certificates

# Install Nodejs20
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt update \
    && apt install nodejs -y

# Clean apt cache
RUN apt-get clean
