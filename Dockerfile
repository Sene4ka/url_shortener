FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o shortener ./cmd/shortener

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/shortener .

COPY .env .env

ARG SERVER_PORT=8080
ENV SERVER_PORT=${SERVER_PORT}
EXPOSE ${SERVER_PORT}

CMD ["./shortener"]