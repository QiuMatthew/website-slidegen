#!/bin/bash
set -e

echo "🔨 Building slide generation service..."

# Build Go binary for Linux
echo "Building Go binary for linux/amd64..."
GOOS=linux GOARCH=amd64 go build -o slidegen-server main.go

# Build Docker image
echo "Building Docker image..."
docker buildx build --platform linux/amd64 -t ghcr.io/qiumatthew/website-slidegen:latest -f Dockerfile.prod --push .

# Clean up
rm slidegen-server

echo "✅ Build complete and pushed to GitHub Container Registry!"
echo "Image: ghcr.io/qiumatthew/website-slidegen:latest"