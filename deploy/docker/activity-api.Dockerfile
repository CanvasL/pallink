FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/activity-api ./activity/api

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/activity-api /app/activity-api
COPY activity/api/etc/activityapi.yaml /app/etc/activityapi.yaml
EXPOSE 8003
CMD ["/app/activity-api", "-f", "etc/activityapi.yaml"]
