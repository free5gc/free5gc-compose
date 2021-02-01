FROM my5gcore/base:latest AS builder
FROM alpine:3.8

LABEL description="My5Gcore open source 5G Core Network" \
    version="Stage 3"

ENV My5GC_MODULE webui
ARG DEBUG_TOOLS

# Install debug tools ~ 100MB (if DEBUG_TOOLS is set to true)
RUN if [ "$DEBUG_TOOLS" = "true" ] ; then apk add -U vim strace net-tools curl netcat-openbsd ; fi

# Set working dir
WORKDIR /my5Gcore
RUN mkdir -p config/ webconsole/publicoo

# Copy executable, frontend static files and default configuration
COPY --from=builder /my5Gcore/${My5GC_MODULE} ./webconsole
COPY --from=builder /my5Gcore/public ./webconsole/public

# Move to the executable location
WORKDIR /my5Gcore/webconsole

# Config files volume
VOLUME [ "/my5Gcore/config" ]

# WebUI uses the port 5000
EXPOSE 5000
