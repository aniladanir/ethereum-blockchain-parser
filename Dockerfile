FROM golang:1.22 as builder

WORKDIR /app

ENV GO111MODULE on
ENV GOBIN=/usr/local/bin/go/bin

COPY go.mod .

RUN go mod download

COPY . ./

RUN go build -o ./bin/ethereum-blockchain-parser ./cmd/txparser

FROM ubuntu:22.04

COPY --from=builder /app/config.json /etc/ethereum-blockchain-parser/config/config.json
COPY --from=builder /app/bin/ethereum-blockchain-parser /opt/app/ethereum-blockchain-parser

RUN apt-get update && apt-get install -y ca-certificates
RUN update-ca-certificates
RUN mkdir /var/log/ethereum-blockchain-parser

ENTRYPOINT ["/opt/app/ethereum-blockchain-parser","--cfg","/etc/ethereum-blockchain-parser/config/config.json"]