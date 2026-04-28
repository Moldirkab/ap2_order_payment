FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o order-service ./cmd/order

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/order-service .

CMD ["./order-service"]