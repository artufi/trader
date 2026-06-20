FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o trader .

FROM alpine:3.24.1

WORKDIR /app

COPY --from=builder /app/trader .

EXPOSE 4000

CMD ["./trader"]