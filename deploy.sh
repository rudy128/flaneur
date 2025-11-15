#!/bin/bash

# Docker Deployment Script for Social API
# This script will build and deploy the entire stack

set -e  # Exit on error

echo "=========================================="
echo "  Social API Docker Deployment"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    print_error "Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

print_success "Docker and Docker Compose are installed"
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    print_info "Creating .env file with default values..."
    cat > .env << 'EOF'
# Backend Environment Variables
DATABASE_URL=postgres://ripper:ripper123@postgres:5432/ripper?sslmode=disable
JWT_SECRET=change-this-to-a-secure-random-string-in-production
WHATSAPP_MICROSERVICE_URL=http://whatsapp-service:8083
PORT=8080
GIN_MODE=release
EOF
    print_success "Created .env file"
else
    print_info ".env file already exists, skipping creation"
fi

echo ""
print_info "Step 1: Stopping existing containers (if any)..."
docker-compose down 2>/dev/null || true
print_success "Stopped existing containers"

echo ""
print_info "Step 2: Building Docker images..."
docker-compose build --no-cache
print_success "Built all Docker images"

echo ""
print_info "Step 3: Starting services..."
docker-compose up -d
print_success "Started all services"

echo ""
print_info "Waiting for services to be healthy..."
sleep 5

# Check service health
echo ""
print_info "Checking service status..."

# Check postgres
if docker-compose ps | grep -q "postgres.*Up"; then
    print_success "PostgreSQL: Running"
else
    print_error "PostgreSQL: Failed to start"
fi

# Check whatsapp-service
if docker-compose ps | grep -q "whatsapp.*Up"; then
    print_success "WhatsApp Service: Running"
else
    print_error "WhatsApp Service: Failed to start"
fi

# Check backend
if docker-compose ps | grep -q "backend.*Up"; then
    print_success "Backend API: Running"
else
    print_error "Backend API: Failed to start"
fi

echo ""
echo "=========================================="
echo "  Deployment Complete!"
echo "=========================================="
echo ""
echo "Service URLs:"
echo "  Backend API:       http://localhost:8080"
echo "  WhatsApp Service:  http://localhost:8083"
echo "  PostgreSQL:        localhost:5432"
echo ""
echo "Useful commands:"
echo "  View logs:         docker-compose logs -f"
echo "  View specific:     docker-compose logs -f backend"
echo "  Stop services:     docker-compose down"
echo "  Restart service:   docker-compose restart backend"
echo "  View status:       docker-compose ps"
echo ""
print_info "To view real-time logs, run: docker-compose logs -f"
echo ""
