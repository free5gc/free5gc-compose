#
# Dockerfile responsible to create only the base image without compiling free5gc sources
#

FROM golang:1.24.5-bullseye AS builder

LABEL maintainer="Free5GC <support@free5gc.org>"

ENV DEBIAN_FRONTEND=noninteractive

# Install dependencies
RUN apt-get update \
    && apt-get -y install gcc cmake autoconf libtool pkg-config libmnl-dev libyaml-dev apt-transport-https ca-certificates

# Clean apt cache
RUN apt-get clean
