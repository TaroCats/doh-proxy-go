FROM golang:1.26.1-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

RUN set -eux; \
	if [ "${TARGETARCH}${TARGETVARIANT}" = "armv7" ]; then export GOARM=7; fi; \
	CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" go build -trimpath -ldflags="-s -w" -o /out/doh-proxy .

FROM alpine:3.21

RUN addgroup -S app && adduser -S app -G app

ENV UPSTREAM_DNS=124.221.68.73:1053

COPY --from=build /out/doh-proxy /usr/local/bin/doh-proxy

USER app

EXPOSE 8053

ENTRYPOINT ["doh-proxy"]
