#!/bin/bash
# FullStackArkham - Quick Start Script
# Builds services locally and starts Docker Compose

set -e

echo "============================================================"
echo "FullStackArkham - Building Services"
echo "============================================================"

cd "$(dirname "$0")"

# Build Go Gateway locally (faster than Docker build)
echo ""
echo "[1/4] Building Gateway (Go)..."
cd services/gateway
if ! command -v go &> /dev/null; then
    echo "ERROR: Go is not installed. Install from https://go.dev/dl/"
    exit 1
fi
go build -o bin/gateway ./app
echo "✓ Gateway built successfully"
cd ../..

# Create bin directory for other services
mkdir -p services/bim_ingestion/bin
mkdir -p services/orchestration/bin
mkdir -p services/memory/bin
mkdir -p services/semantic-cache/bin
mkdir -p services/billing/bin
mkdir -p services/arkham/bin

# Build Python services (just create wrapper scripts that call python)
echo ""
echo "[2/4] Creating Python service wrappers..."

for service in arkham bim_ingestion orchestration memory semantic-cache billing; do
    cat > services/${service}/bin/start.sh << EOF
#!/bin/bash
cd /app
exec python -m uvicorn app.main:app --host 0.0.0.0 --port \${PORT:-8080}
EOF
    chmod +x services/${service}/bin/start.sh
    echo "  - ${service}"
done

echo "✓ Python service wrappers created"

# Verify Docker is running
echo ""
echo "[3/4] Checking Docker..."
if ! docker info &> /dev/null; then
    echo "ERROR: Docker is not running. Start Docker Desktop first."
    exit 1
fi
echo "✓ Docker is running"

# Stop existing containers
echo ""
echo "[4/4] Stopping existing containers..."
docker-compose down 2>/dev/null || true

# Start all services
echo ""
echo "============================================================"
echo "Starting all services..."
echo "============================================================"
docker-compose up -d --build

echo ""
echo "============================================================"
echo "Waiting for services to start..."
echo "============================================================"
sleep 15

# Check service status
echo ""
echo "============================================================"
echo "Service Status:"
echo "============================================================"
docker-compose ps

echo ""
echo "============================================================"
echo "Testing service health..."
echo "============================================================"

# Test each service
for i in {1..30}; do
    echo -n "Attempt $i/30: "
    
    GATEWAY_OK=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health 2>/dev/null || echo "000")
    ARKHAM_OK=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/health 2>/dev/null || echo "000")
    
    if [ "$GATEWAY_OK" = "200" ] && [ "$ARKHAM_OK" = "200" ]; then
        echo "✓ Services are ready!"
        echo ""
        echo "============================================================"
        echo "Running E2E Tests..."
        echo "============================================================"
        python3 tests/run_e2e_test.py
        exit $?
    fi
    
    sleep 2
done

echo ""
echo "⚠ Services didn't become ready in time"
echo ""
echo "Check logs with: docker-compose logs"
exit 1
