FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go test ./...
RUN go build -ldflags "-s -w" -trimpath -o /build/dancegobot ./cmd/dancegobot

FROM alpine:3.20
COPY --from=builder /build/dancegobot /
EXPOSE 8080
VOLUME ["/data"]
ENV DB_FILEPATH=/data/dancegobot.db
CMD ["/dancegobot"]
