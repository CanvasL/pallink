FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/user-api ./user/api

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/user-api /app/user-api
COPY user/api/etc/userapi.yaml /app/etc/userapi.yaml
EXPOSE 8001
CMD ["/app/user-api", "-f", "etc/userapi.yaml"]
