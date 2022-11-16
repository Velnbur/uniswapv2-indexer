FROM golang:1.18-alpine as buildbase

RUN apk add git build-base

WORKDIR /go/src/github.com/Velnbur/uniswapv2-indexer
COPY vendor .
COPY . .

RUN GOOS=linux go build  -o /usr/local/bin/uniswapv2-indexer /go/src/github.com/Velnbur/uniswapv2-indexer


FROM alpine:3.9

COPY --from=buildbase /usr/local/bin/uniswapv2-indexer /usr/local/bin/uniswapv2-indexer
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["uniswapv2-indexer"]
