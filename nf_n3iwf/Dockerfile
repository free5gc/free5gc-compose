FROM free5gc/n3iwf-base:latest AS builder
FROM alpine:3.15

LABEL description="Free5GC open source 5G Core Network" \
    version="Stage 3"

ENV F5GC_MODULE n3iwf
ARG DEBUG_TOOLS

# Install debug tools ~ 100MB (if DEBUG_TOOLS is set to true)
RUN if [ "$DEBUG_TOOLS" = "true" ] ; then apk add -U vim strace net-tools curl netcat-openbsd ; fi

# Install N3IWF dependencies
RUN apk add -U iproute2

# Set working dir
WORKDIR /free5gc
RUN mkdir -p config/ log/ cert/

# Copy executable and default certs
COPY --from=builder /free5gc/${F5GC_MODULE} ./
COPY --from=builder /free5gc/cert/${F5GC_MODULE}.pem ./cert/
COPY --from=builder /free5gc/cert/${F5GC_MODULE}.key ./cert/

# Config files volume
VOLUME [ "/free5gc/config" ]

# Certificates (if not using default) volume
VOLUME [ "/free5gc/cert" ]
