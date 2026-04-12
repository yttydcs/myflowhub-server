FROM --platform=$BUILDPLATFORM golang:1.25-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/hub_server ./cmd/hub_server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates \
    && addgroup -S myflowhub \
    && adduser -S -D -H -h /data -G myflowhub myflowhub \
    && mkdir -p /data \
    && chown myflowhub:myflowhub /data

COPY --from=build /out/hub_server /usr/local/bin/hub_server

ENV HUB_ADDR=:9000 \
    HUB_QUIC_ADDR=:9000 \
    HUB_WORKDIR=/data

WORKDIR /data
USER myflowhub:myflowhub

EXPOSE 9000/tcp
EXPOSE 9000/udp

ENTRYPOINT ["/usr/local/bin/hub_server"]
