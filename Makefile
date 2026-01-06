.PHONY: dev build run css css-watch db-up db-down migrate test clean

# Development server with file watching
dev:
	go run github.com/a-h/templ/cmd/templ generate
	go run ./cmd/server

# Generate templ files
templ:
	go run github.com/a-h/templ/cmd/templ generate

# Build production binary
build:
	go build -o bin/server ./cmd/server

# Run the built binary
run:
	./bin/server

# Build Tailwind CSS
css:
	npx @tailwindcss/cli -i ./web/static/css/tailwind.css -o ./web/static/css/output.css --minify

# Watch Tailwind CSS for changes
css-watch:
	npx @tailwindcss/cli -i ./web/static/css/tailwind.css -o ./web/static/css/output.css --watch

# Start PostgreSQL container
db-up:
	docker-compose up -d

# Stop PostgreSQL container
db-down:
	docker-compose down

# Run database migrations
migrate:
	go run ./cmd/migrate

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf tmp/
