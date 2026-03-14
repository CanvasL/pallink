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
    go build -o /app/notification-rpc ./notification

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/notification-rpc /app/notification-rpc
COPY notification/etc/notification.yaml /app/etc/notification.yaml
EXPOSE 8006
CMD ["/app/notification-rpc", "-f", "etc/notification.yaml"]
