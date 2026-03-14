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
    go build -o /app/audit-worker ./audit

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/audit-worker /app/audit-worker
COPY audit/etc/audit.yaml /app/etc/audit.yaml
CMD ["/app/audit-worker", "-f", "etc/audit.yaml"]
