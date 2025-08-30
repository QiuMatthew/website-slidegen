# Development Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY main.go .
RUN go build -o slidegen-server main.go

FROM node:18-alpine

# Install reveal-md
RUN npm install -g reveal-md

WORKDIR /app

# Copy the Go server
COPY --from=builder /app/slidegen-server .

# Create slides directory
RUN mkdir -p slides

EXPOSE 8081 1948

CMD ["./slidegen-server"]