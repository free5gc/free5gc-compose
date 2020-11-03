FROM my5gcore/ue:latest AS builder
FROM alpine:3.8

LABEL description="User Equipment in LoRaWAN Gateway" \
    version="Stage 3"

ENV MY5GC_MODULE ue
ARG DEBUG_TOOLS

# Install debug tools ~ 100MB (if DEBUG_TOOLS is set to true)
RUN if [ "$DEBUG_TOOLS" = "true" ] ; then apk add -U vim strace net-tools curl netcat-openbsd ; fi

# Install UE dependencies
RUN apk add -U iproute2 python3

# Set working dir
WORKDIR /my5gcore
RUN mkdir -p bin/ config/ log/ support/TLS/ 

# Copy executable and default certs
COPY --from=builder /my5Gcore/ue  ./bin/
COPY --from=builder /my5Gcore/support/TLS/${MY5GC_MODULE}.pem ./support/TLS/
COPY --from=builder /my5Gcore/support/TLS/${MY5GC_MODULE}.key ./support/TLS/

# Move to the binary path
WORKDIR /my5gcore/

# Config files volume
VOLUME [ "/my5gcore/config" ]

# Certificates (if not using default) volume
VOLUME [ "/my5gcore/support/TLS" ]
