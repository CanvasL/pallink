FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/user-rpc ./user

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/user-rpc /app/user-rpc
COPY user/etc/user.yaml /app/etc/user.yaml
EXPOSE 8002
CMD ["/app/user-rpc", "-f", "etc/user.yaml"]
