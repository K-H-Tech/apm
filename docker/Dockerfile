FROM golang:1.23-alpine as builder

ARG VERSION
ARG COMMIT_SHA
ARG BUILD_DATE

WORKDIR /build

COPY . .

RUN go build --ldflags "-w -s -X 'main.Version=${VERSION}' -X 'main.CommitSHA=${COMMIT_SHA}' -X 'main.BuildDate=${BUILD_DATE}'" -v -o server cmd/server/server.go

FROM alpine:3.20

RUN apk update && \
    apk add --update --no-cache \
    tzdata \
    bash \
    ca-certificates \
    curl \
    redis \
    gcompat \
    bind-tools \
    yq \
    jq \
    busybox-extras \
    zip \
    unzip \
    tcpdump

COPY --from=builder /build/server /usr/bin/tm-server

CMD ["tm-server", "-c", "/etc/tax-manager/server.config.yaml"]
