ARG MODULE=github.com/ofstudio/dancegobot
ARG VERSION=latest

FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go test ./...
ARG MODULE
ARG VERSION
RUN go build -trimpath \
      -ldflags "-s -w -X ${MODULE}/internal/config.version=${VERSION}" \
      -o /build/dancegobot ./cmd/dancegobot

FROM alpine:3.20
LABEL org.opencontainers.image.source='https://github.com/ofstudio/dancegobot'
LABEL org.opencontainers.image.description='Telegram bot for finding a partner for dance events'
LABEL org.opencontainers.image.licenses='Apache License 2.0'
COPY --from=builder /build/dancegobot /
EXPOSE 8080
VOLUME ["/data"]
ENV DB_FILEPATH=/data/dancegobot.db
CMD ["/dancegobot"]
