FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/gateway-api ./gateway

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/gateway-api /app/gateway-api
COPY gateway/etc/gatewayapi.yaml /app/etc/gatewayapi.yaml
EXPOSE 8080
CMD ["/app/gateway-api", "-f", "etc/gatewayapi.yaml"]
