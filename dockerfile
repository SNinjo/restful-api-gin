FROM golang:1.23.0 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4
RUN swag init --output ./docs
RUN go build -o main .

FROM debian:bookworm-slim
WORKDIR /root/

RUN apt-get update && apt-get install -y ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs

EXPOSE 8080
CMD ["./main"]
