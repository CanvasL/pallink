FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/im-rpc ./im

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/im-rpc /app/im-rpc
COPY im/etc/im.yaml /app/etc/im.yaml
EXPOSE 8005
CMD ["/app/im-rpc", "-f", "etc/im.yaml"]
