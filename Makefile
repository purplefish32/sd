.PHONY: all build run generate clean dev

# Default target
all: generate build

# Generate templ files
generate:
	@echo "Generating templ files..."
	@go generate ./cmd/web/views/...
	@echo "Building Tailwind CSS..."
	@cd cmd/web && npx tailwindcss -i ./assets/css/input.css -o ./static/css/style.css --minify

# Build the web application
build: generate
	@echo "Building web application..."
	@go build -o bin/web cmd/web/main.go

# Run the web application
run: generate
	@echo "Running web application..."
	@go run cmd/web/main.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@find . -type f -name '*_templ.go' -delete
	@rm -f cmd/web/static/css/style.css

# Development mode with auto-reload (requires air)
dev:
	@echo "Starting development server..."
	@air -c .air.toml