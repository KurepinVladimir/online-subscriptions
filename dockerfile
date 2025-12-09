# build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o subscriptions-service ./cmd/app

# run stage
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/subscriptions-service /app/subscriptions-service
COPY configs ./configs
COPY migrations ./migrations

ENV APP_LOGGING_LEVEL=info

EXPOSE 8080

CMD ["/app/subscriptions-service"]
