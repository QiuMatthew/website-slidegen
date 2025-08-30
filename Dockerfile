# Development Dockerfile - Lightweight version
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod .
COPY main.go .
RUN go build -o slidegen-server main.go

FROM alpine:latest

WORKDIR /app

# Copy the Go server
COPY --from=builder /app/slidegen-server .

# Create static directory
RUN mkdir -p static

EXPOSE 8081

CMD ["./slidegen-server"]