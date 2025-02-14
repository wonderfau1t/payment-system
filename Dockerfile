FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/cmd/payment-system

RUN CGO_ENABLED=0 go build -o payment-system
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/cmd/payment-system/payment-system .
COPY config/config.yaml ./config/

RUN apk --no-cache add \
    sqlite \
    ca-certificates

EXPOSE 8080
ENV CONFIG_PATH=./config/config.yaml
CMD ["./payment-system"]