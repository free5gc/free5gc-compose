#
# Dockerfile responsible to create only the base image without compiling free5gc sources
#

FROM golang:1.18.10-bullseye AS builder

LABEL maintainer="Free5GC <support@free5gc.org>"

ENV DEBIAN_FRONTEND noninteractive

# Install dependencies
RUN apt-get update \
    && apt-get -y install gcc cmake autoconf libtool pkg-config libmnl-dev libyaml-dev apt-transport-https ca-certificates \
    && curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/master/install.sh | bash \
    && bash -c 'source ~/.bashrc && nvm install --lts && npm install -g yarn' \
    && bash -c 'source ~/.bashrc && ln -s $(which node) $(which yarn) /usr/bin/'

# Clean apt cache
RUN apt-get clean
