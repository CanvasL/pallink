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
    go build -o /app/activity-rpc ./activity

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/activity-rpc /app/activity-rpc
COPY activity/etc/activity.yaml /app/etc/activity.yaml
EXPOSE 8004
CMD ["/app/activity-rpc", "-f", "etc/activity.yaml"]
