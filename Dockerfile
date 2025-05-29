# Build stage
FROM golang:1.23.9-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Deployment stage
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080

CMD ["./main"]