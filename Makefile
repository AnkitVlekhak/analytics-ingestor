.PHONY: all run-server run-client db-up db-down db-logs gen tidy help

# -----------------------------------------------------------------------------
# Configuration
# -----------------------------------------------------------------------------
# If your proto file is in a different folder, update this path
PROTO_FILE=proto/analytics.proto

# -----------------------------------------------------------------------------
# Main Commands
# -----------------------------------------------------------------------------

# Default command: setup deps and start db
all: tidy db-up

# Run the Ingestor Server
run-server:
	@echo "🚀 Starting Analytics Server..."
	go run cmd/server/main.go

# Run the Load Test Client
run-client:
	@echo "🧪 Starting Load Generator..."
	go run cmd/client/main.go

# -----------------------------------------------------------------------------
# Database (Docker)
# -----------------------------------------------------------------------------

# Start TimescaleDB in background
db-up:
	@echo "🐳 Starting Database..."
	docker-compose up -d

# Stop Database
db-down:
	@echo "🛑 Stopping Database..."
	docker-compose down

# Check DB Logs (Simulating 'tail -f')
db-logs:
	docker-compose logs -f timescaledb

# -----------------------------------------------------------------------------
# Development Utils
# -----------------------------------------------------------------------------

# Re-generate .pb.go files from your Proto definition
# NOTE: Ensure you have protoc installed
gen:
	@echo "⚙️  Generating Proto files..."
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       $(PROTO_FILE)

# Clean up Go modules
tidy:
	go mod tidy

# Show help menu
help:
	@echo "Make commands:"
	@echo "  make db-up       -> Start TimescaleDB (Docker)"
	@echo "  make db-down     -> Stop TimescaleDB"
	@echo "  make run-server  -> Run the gRPC Server"
	@echo "  make run-client  -> Run the Load Test Client"
	@echo "  make gen         -> Re-generate Proto files"