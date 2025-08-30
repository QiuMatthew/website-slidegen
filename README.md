# Website Slide Generation Service

A microservice for converting markdown files to interactive presentations using reveal-md.

## Features

- Upload markdown files via HTTP API
- Automatic conversion to reveal.js presentations
- Live preview of slides
- RESTful API for integration with frontend

## API Endpoints

- `POST /upload` - Upload a markdown file
- `GET /health` - Health check
- `GET /` - Access the slide presentation

## Local Development

```bash
# Run locally
go run main.go

# Build Docker container
docker build -f Dockerfile -t website-slidegen:test .

# Run container
docker run -p 8081:8081 -p 1948:1948 website-slidegen:test
```

## Production Build

```bash
# Build and push to GitHub Container Registry
./build.sh
```

## Markdown Syntax

- Use `---` to separate slides
- Use `#` for headings
- Use `-` for bullet points
- Supports code blocks, images, tables, and more

## Architecture

This service uses:
- Go HTTP server for handling uploads and routing
- reveal-md (Node.js) for rendering presentations
- Single container approach for simplicity