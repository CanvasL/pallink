FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/audit-worker ./audit

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/audit-worker /app/audit-worker
COPY audit/etc/audit.yaml /app/etc/audit.yaml
CMD ["/app/audit-worker", "-f", "etc/audit.yaml"]
