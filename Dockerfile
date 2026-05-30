# Stage 1: Build the Vite frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /web
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# Stage 2: Build the Go backend
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy the built frontend assets into the embedded assets folder
COPY --from=frontend-builder /web/dist /app/pkg/api/web-dist
RUN go build -o clinlang ./cmd/clinlang

# Stage 3: Minimal runner image
FROM alpine:3.19
WORKDIR /app
COPY --from=backend-builder /app/clinlang /app/clinlang

# Expose default port
EXPOSE 8080

# Set default configuration variables for deployment
ENV CLINLANG_MODE=local
ENV CLINLANG_BIND=0.0.0.0:8080
ENV CLINLANG_WORKSPACE=/app/workspace

# Create workspace directory
RUN mkdir -p /app/workspace

# Start the server daemon
ENTRYPOINT ["/app/clinlang", "server"]
