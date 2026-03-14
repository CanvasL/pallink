FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/notification-rpc ./notification

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/notification-rpc /app/notification-rpc
COPY notification/etc/notification.yaml /app/etc/notification.yaml
EXPOSE 8006
CMD ["/app/notification-rpc", "-f", "etc/notification.yaml"]
