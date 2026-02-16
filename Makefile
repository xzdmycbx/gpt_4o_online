.PHONY: help build-frontend build-backend build-all dev prod clean test init download-geoip

# Default target
help:
	@echo "AI Chat System - Build & Deployment Commands"
	@echo ""
	@echo "Development:"
	@echo "  make init           - Initialize project (install dependencies, download GeoIP)"
	@echo "  make dev            - Start development environment"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Building:"
	@echo "  make build-frontend - Build frontend only"
	@echo "  make build-backend  - Build backend only (includes frontend)"
	@echo "  make build-all      - Build everything"
	@echo ""
	@echo "Production:"
	@echo "  make prod           - Start production environment"
	@echo "  make prod-stop      - Stop production environment"
	@echo "  make prod-logs      - View production logs"
	@echo ""
	@echo "Database:"
	@echo "  make db-migrate     - Run database migrations"
	@echo "  make db-reset       - Reset database (WARNING: deletes all data)"
	@echo ""
	@echo "Utilities:"
	@echo "  make download-geoip - Download GeoIP2 database"
	@echo "  make test           - Run tests"

# Initialize project
init: download-geoip
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "Downloading Go dependencies..."
	cd backend && go mod download
	@echo "Initialization complete!"

# Download GeoIP2 database (GeoLite2 free version)
download-geoip:
	@echo "Downloading GeoIP2 database..."
	@mkdir -p backend/data
	@echo "Please download GeoLite2-Country.mmdb from https://dev.maxmind.com/geoip/geolite2-free-geolocation-data"
	@echo "and place it in backend/data/"

# Build frontend
build-frontend:
	@echo "Building frontend..."
	cd frontend && npm install && npm run build
	@echo "Copying frontend build to backend..."
	@mkdir -p backend/web/dist
	@cp -r frontend/dist/* backend/web/dist/
	@echo "Frontend build complete!"

# Build backend (includes frontend)
build-backend: build-frontend
	@echo "Building backend..."
	cd backend && go build -o ../bin/ai-chat cmd/server/main.go
	@echo "Backend build complete! Binary: bin/ai-chat"

# Build everything
build-all: build-backend
	@echo "Complete build finished!"

# Development environment
dev:
	@echo "Starting development environment..."
	@docker-compose -f docker-compose.dev.yml up --build

# Development environment (detached)
dev-detached:
	@echo "Starting development environment (detached)..."
	@docker-compose -f docker-compose.dev.yml up -d --build

# Stop development environment
dev-stop:
	@docker-compose -f docker-compose.dev.yml down

# Production environment
prod:
	@echo "Starting production environment..."
	@docker-compose up -d --build
	@echo "Production environment started!"
	@echo "Access the application at http://localhost:8080"

# Stop production environment
prod-stop:
	@echo "Stopping production environment..."
	@docker-compose down

# View production logs
prod-logs:
	@docker-compose logs -f

# Database migrations
db-migrate:
	@echo "Running database migrations..."
	@docker-compose exec postgres psql -U ai_chat_user -d ai_chat_db -f /docker-entrypoint-initdb.d/001_init.sql

# Reset database (WARNING: deletes all data)
db-reset:
	@echo "WARNING: This will delete all data!"
	@read -p "Are you sure? (yes/no): " confirm && [ "$$confirm" = "yes" ] || exit 1
	@docker-compose down -v
	@docker-compose up -d postgres
	@sleep 5
	@make db-migrate
	@echo "Database reset complete!"

# Note: Super admin is created automatically on first startup
# Configure in .env file:
#   SUPER_ADMIN_USERNAME=admin
#   SUPER_ADMIN_PASSWORD=your_password
#   SUPER_ADMIN_EMAIL=admin@example.com

# Run tests
test:
	@echo "Running backend tests..."
	cd backend && go test -v ./...
	@echo "Running frontend tests..."
	cd frontend && npm test

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf frontend/dist
	@rm -rf frontend/node_modules/.vite
	@rm -rf backend/web/dist
	@rm -rf backend/tmp
	@rm -rf bin/
	@echo "Clean complete!"

# Clean everything (including dependencies)
clean-all: clean
	@echo "Removing all dependencies..."
	@rm -rf frontend/node_modules
	@rm -rf backend/vendor
	@echo "Deep clean complete!"

# Generate go.sum
go-tidy:
	cd backend && go mod tidy

# Format code
format:
	@echo "Formatting Go code..."
	cd backend && go fmt ./...
	@echo "Formatting TypeScript code..."
	cd frontend && npm run lint --fix
