# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api

FROM alpine:3.21
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /out/api /app/api

ENV HTTP_ADDR=:8080
EXPOSE 8080

USER app
ENTRYPOINT ["/app/api"]
