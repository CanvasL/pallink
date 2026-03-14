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
    go build -o /app/user-rpc ./user

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/user-rpc /app/user-rpc
COPY user/etc/user.yaml /app/etc/user.yaml
EXPOSE 8002
CMD ["/app/user-rpc", "-f", "etc/user.yaml"]
