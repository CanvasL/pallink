FROM golang:1.25-alpine AS builder

WORKDIR /src
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod,id=pallink-go-mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,id=pallink-go-build \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod,id=pallink-go-mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,id=pallink-go-build \
    go build -o /app/gateway-api ./gateway

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/gateway-api /app/gateway-api
COPY gateway/etc/gatewayapi.yaml /app/etc/gatewayapi.yaml
EXPOSE 8080
CMD ["/app/gateway-api", "-f", "etc/gatewayapi.yaml"]
