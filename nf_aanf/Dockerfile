FROM golang:1.16 as builder

WORKDIR /free5gc
COPY . .

RUN cd src/aanf && \
    go build -o aanf

FROM ubuntu:20.04

WORKDIR /free5gc
COPY --from=builder /free5gc/src/aanf/aanf /free5gc/aanf
COPY config/ config/
COPY cert/ cert/

ENTRYPOINT ["/free5gc/aanf"]