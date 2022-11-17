# syntax=docker/dockerfile:1

FROM golang:1.19-alpine as builder

WORKDIR /app

COPY . ./voice-plugin

WORKDIR /app/voice-plugin

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN go generate

RUN go build -o /go/bin/app github.com/openline-ai/openline-voice/packages/apps/voice-plugin

FROM alpine:3.14
COPY --chown=65534:65534 --from=builder /go/bin/app .
USER 65534

ENTRYPOINT [ "./app" ]
