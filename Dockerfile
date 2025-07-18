FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o reservio-bot main.go

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/reservio-bot .

ENTRYPOINT ["/app/reservio-bot"]
