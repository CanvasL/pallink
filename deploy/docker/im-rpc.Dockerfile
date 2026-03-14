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
    go build -o /app/im-rpc ./im

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/im-rpc /app/im-rpc
COPY im/etc/im.yaml /app/etc/im.yaml
EXPOSE 8005
CMD ["/app/im-rpc", "-f", "etc/im.yaml"]
