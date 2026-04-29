#!/bin/bash
# Papabase VPS Deployment Script
# Run this on your VPS (Ubuntu/Debian)

set -e

echo "=============================================="
echo "  Papabase VPS Deployment Script"
echo "=============================================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "Please run as root (sudo ./deploy.sh)"
    exit 1
fi

# Update system
echo "[1/8] Updating system packages..."
apt update && apt upgrade -y

# Install Docker
echo "[2/8] Installing Docker..."
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
rm get-docker.sh

# Install Docker Compose
echo "[3/8] Installing Docker Compose..."
DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
mkdir -p $DOCKER_CONFIG/cli-plugins
curl -SL https://github.com/docker/compose/releases/download/v2.27.0/docker-compose-linux-x86_64 -o $DOCKER_CONFIG/cli-plugins/docker-compose
chmod +x $DOCKER_CONFIG/cli-plugins/docker-compose

# Clone repository
echo "[4/8] Cloning Papabase repository..."
cd /opt
git clone https://github.com/robs46859-eng/stack-arkham.git || echo "Repo already exists"
cd stack-arkham

# Create .env file
echo "[5/8] Creating environment file..."
if [ ! -f .env ]; then
    cp .env.production.example .env
    echo "⚠️  Please edit .env with your credentials before continuing"
    echo "   Required: DATABASE_PASSWORD, JWT_SECRET, STRIPE_SECRET_KEY, GOOGLE_API_KEY"
    read -p "Press Enter after you've edited .env..."
fi

# Generate JWT secret if not set
if grep -q "CHANGE_THIS" .env; then
    echo "⚠️  .env still contains placeholder values"
    echo "   Please edit .env before running docker compose up"
    exit 1
fi

# Create necessary directories
echo "[6/8] Creating data directories..."
mkdir -p data/postgres data/redis data/backups

# Pull and start services
echo "[7/8] Starting Papabase services..."
docker compose pull
docker compose up -d

# Wait for services to be ready
echo "[8/8] Waiting for services to start..."
sleep 30

# Check health
echo ""
echo "=============================================="
echo "  Checking service health..."
echo "=============================================="

# Check each service
for service in gateway papabase postgres redis; do
    if docker compose ps | grep -q "$service.*Up"; then
        echo "✓ $service is running"
    else
        echo "✗ $service is NOT running"
    fi
done

echo ""
echo "=============================================="
echo "  DEPLOYMENT COMPLETE!"
echo "=============================================="
echo ""
echo "Papabase is now running at:"
echo "  Frontend: http://$(curl -s ifconfig.me)"
echo "  API: http://$(curl -s ifconfig.me):8087"
echo ""
echo "Next steps:"
echo "1. Set up SSL with Caddy (see DEPLOYMENT_HOSTINGER.md)"
echo "2. Configure your domain (fsai.pro) DNS"
echo "3. Set up Stripe webhook"
echo "4. Run Stripe products script:"
echo "   python scripts/create_stripe_products.py"
echo ""
echo "Useful commands:"
echo "  docker compose ps          # Check status"
echo "  docker compose logs -f     # View logs"
echo "  docker compose restart     # Restart all"
echo "  docker compose down        # Stop all"
echo ""
