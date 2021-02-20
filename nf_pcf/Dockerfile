FROM free5gc/base:latest AS builder
FROM alpine:3.13

LABEL description="Free5GC open source 5G Core Network" \
    version="Stage 3"

ENV F5GC_MODULE pcf
ARG DEBUG_TOOLS

# Install debug tools ~ 100MB (if DEBUG_TOOLS is set to true)
RUN if [ "$DEBUG_TOOLS" = "true" ] ; then apk add -U vim strace net-tools curl netcat-openbsd ; fi

# Set working dir
WORKDIR /free5gc
RUN mkdir -p config/ log/ support/TLS/ ${F5GC_MODULE}/

# Copy executable and default certs
COPY --from=builder /free5gc/${F5GC_MODULE} ./${F5GC_MODULE}
COPY --from=builder /free5gc/support/TLS/${F5GC_MODULE}.pem ./support/TLS/
COPY --from=builder /free5gc/support/TLS/${F5GC_MODULE}.key ./support/TLS/

# Move to the binary path
WORKDIR /free5gc/${F5GC_MODULE}

# Config files volume
VOLUME [ "/free5gc/config" ]

# Certificates (if not using default) volume
VOLUME [ "/free5gc/support/TLS" ]

# Exposed ports
EXPOSE 8000
