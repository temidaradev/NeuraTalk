run:
	@go run main.go

# Web-related targets
build-web:
	@./scripts/build-web.sh

server:
	@go run ./cmd/server

web: build-web
	@echo "Starting NeuraTalk server..."
	@echo "Open http://localhost:8080 in your browser"
	@./bin/neuratalk-server

clean-web:
	@rm -f web/main.wasm
	@rm -f bin/neuratalk-server

# Create directories
dirs:
	@mkdir -p bin web

.PHONY: run build-web server web clean-web dirs