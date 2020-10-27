FROM my5gcore/base:latest AS builder
FROM ubuntu:18.04

LABEL description="My5Gcore open source 5G Core Network" \
    version="Stage 3"

ENV My5GC_MODULE my5Gcore-upfd
ENV DEBIAN_FRONTEND noninteractive
ARG DEBUG_TOOLS

# Install debug tools ~ 100MB (if DEBUG_TOOLS is set to true)
RUN if [ "$DEBUG_TOOLS" = "true" ] ; then apt-get update && apt-get install -y vim strace net-tools iputils-ping curl netcat ; fi

# Install UPF dependencies
RUN apt-get update \
    && apt-get install -y libmnl0 libyaml-0-2 iproute2 \
    && apt-get clean

# Set working dir
WORKDIR /my5Gcore
RUN mkdir -p ${My5GC_MODULE}/config/

# Copy executable and default configuration
COPY --from=builder /my5Gcore/${My5GC_MODULE} ./${My5GC_MODULE}/

# Copy linked libs
COPY --from=builder /my5Gcore/libgtp5gnl.so.0 /usr/local/lib
COPY --from=builder /my5Gcore/liblogger.so /usr/local/lib

# Move to the executable location
WORKDIR /my5Gcore/${My5GC_MODULE}

# Update links
RUN ldconfig
