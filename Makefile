.PHONY: setup run build test clean

# Install Homebrew dependencies and create a local .env from the template.
setup:
	@chmod +x scripts/install.sh
	@./scripts/install.sh
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "==> Created .env from template — edit it and add your TELEGRAM_BOT_TOKEN"; \
	else \
		echo "==> .env already exists — leaving it untouched"; \
	fi

# Run the service (reads TELEGRAM_BOT_TOKEN from .env).
run:
	@go run .

# Build a standalone binary.
build:
	@go build -o music-queue .

# Run tests.
test:
	@go test ./...

# Remove build artifacts.
clean:
	@rm -f music-queue
