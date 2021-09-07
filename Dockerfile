FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOPROXY https://goproxy.cn,direct

WORKDIR /build
COPY . .
RUN go mod tidy
RUN go build -ldflags="-s -w" -o /app/main cmd/connect/main.go


FROM alpine

RUN apk update --no-cache
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache tzdata

ARG envType=local
ENV gim_env ${envType}

WORKDIR /app
COPY --from=builder /app/main /app/main

CMD ["./main"]
