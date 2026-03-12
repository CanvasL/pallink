FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/activity-rpc ./activity

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/activity-rpc /app/activity-rpc
COPY activity/etc/activity.yaml /app/etc/activity.yaml
EXPOSE 8004
CMD ["/app/activity-rpc", "-f", "etc/activity.yaml"]
