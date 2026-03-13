FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/notify-rpc ./notify

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/notify-rpc /app/notify-rpc
COPY notify/etc/notify.yaml /app/etc/notify.yaml
EXPOSE 8006
CMD ["/app/notify-rpc", "-f", "etc/notify.yaml"]
