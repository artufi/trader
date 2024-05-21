FROM golang:1.21-alpine AS deps

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY --from=deps /go/pkg /go/pkg
COPY --from=deps /go/bin /go/bin

COPY . .

RUN go build -o trader .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/trader .
COPY config/.env config/.env

EXPOSE 4000

CMD ["./trader"]