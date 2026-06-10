.PHONY: setup run build test clean check-deps

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

# Check for required runtime dependencies and prompt to run make setup if any are missing.
check-deps:
	@missing=""; \
	for dep in go mpv yt-dlp; do \
		if ! command -v $$dep >/dev/null 2>&1; then \
			missing="$$missing $$dep"; \
		fi; \
	done; \
	if [ -n "$$missing" ]; then \
		echo ""; \
		echo "  Missing dependencies:$$missing"; \
		echo ""; \
		echo "  Run \`make setup\` to install them, then try again."; \
		echo ""; \
		exit 1; \
	fi

# Run the service (reads TELEGRAM_BOT_TOKEN from .env).
run: check-deps
	@go run .

# Build a standalone binary.
build:
	@go build -o music-kwewe .

# Run tests.
test:
	@go test ./...

# Remove build artifacts.
clean:
	@rm -f music-kwewe
