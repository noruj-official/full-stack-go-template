# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies including Node.js for CSS compilation
RUN apk add --no-cache git ca-certificates tzdata nodejs npm

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy package files for npm dependencies
COPY package.json ./
RUN npm install

# Copy source code
COPY . .

# Build CSS from source
RUN npm run build

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Build the application
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server

# Production stage
FROM alpine:3.20

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Copy web assets (templates and static files)
COPY --from=builder /app/web /app/web

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Run the application
CMD ["/app/server"]
