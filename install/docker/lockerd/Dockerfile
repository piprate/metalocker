FROM golang:1.25.7-alpine3.22 AS build-stage
RUN apk add --no-cache musl-dev gcc make git openssh-client libsodium-dev

RUN mkdir -p "$GOPATH/pkg/mod"

WORKDIR /build

COPY Makefile .
COPY go.mod .
COPY go.sum .
RUN --mount=type=cache,target=$GOPATH/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make install-deps

COPY . .

RUN --mount=type=cache,target=$GOPATH/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make install

#-----------

FROM alpine:3.22.3 AS run-stage
RUN apk add --no-cache libsodium-dev

WORKDIR /app
COPY --from=build-stage build/bin/lockerd /app/
RUN mkdir ~/.metalocker
RUN mkdir ~/.metalocker/state

ENTRYPOINT ["./lockerd"]
