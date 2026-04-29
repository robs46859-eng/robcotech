#!/bin/bash
# Papabase Azure Deployment Script - Simple Version
# Completes deployment with existing infrastructure

set -e

echo "=============================================="
echo "  Papabase Azure Deployment Agent"
echo "  Completing Deployment..."
echo "=============================================="
echo ""

# Configuration
RESOURCE_GROUP="papabase-rg-2"
LOCATION="westus"
APP_NAME="papabase"
ACR_NAME="papabaseacr5b6c70ba"
POSTGRES_NAME="papabase-psql"
REDIS_NAME="papabase-redis"

# Get PostgreSQL details
echo "Getting PostgreSQL connection..."
DB_HOST=$(az postgres flexible-server show -n $POSTGRES_NAME -g $RESOURCE_GROUP --query "fullyQualifiedDomainName" -o tsv)
echo "✓ PostgreSQL: $DB_HOST"

# Get Redis details
echo "Getting Redis connection..."
REDIS_HOST=$(az redis show -n $REDIS_NAME -g $RESOURCE_GROUP --query "hostName" -o tsv)
REDIS_KEY=$(az redis list-keys -n $REDIS_NAME -g $RESOURCE_GROUP --query "primaryKey" -o tsv)
echo "✓ Redis: $REDIS_HOST"

# Required secrets are read from the environment so they do not get committed.
: "${DB_PASSWORD:?Set DB_PASSWORD before running deploy-to-azure.sh}"
: "${GOOGLE_API_KEY:?Set GOOGLE_API_KEY before running deploy-to-azure.sh}"
: "${STRIPE_SECRET_KEY:?Set STRIPE_SECRET_KEY before running deploy-to-azure.sh}"
: "${STRIPE_WEBHOOK_SECRET:?Set STRIPE_WEBHOOK_SECRET before running deploy-to-azure.sh}"
JWT_SECRET=${JWT_SECRET:-$(openssl rand -hex 32)}

echo ""
echo "Building Gateway Docker image..."
cd services/gateway
az acr build --registry $ACR_NAME --image gateway:latest . 
cd ../..
echo "✓ Gateway image build complete"

echo ""
echo "Building Papabase Docker image..."
cd services/papabase
az acr build --registry $ACR_NAME --image papabase:latest . 
cd ../..
echo "✓ Papabase image build complete"
echo ""

# Get ACR credentials
ACR_USER=$(az acr credential show -n $ACR_NAME --query username -o tsv)
ACR_PASS=$(az acr credential show -n $ACR_NAME --query "passwords[0].value" -o tsv)

# Create Container Apps Environment
echo "Creating Container Apps Environment..."
az containerapp env show -n ${APP_NAME}-env -g $RESOURCE_GROUP &>/dev/null || {
    az containerapp env create \
      --name ${APP_NAME}-env \
      --resource-group $RESOURCE_GROUP \
      --location $LOCATION \
      --output table
}
echo "✓ Container Apps Environment ready"
echo ""

# Deploy Gateway
echo "Deploying Gateway to Container Apps..."
GATEWAY_ENV=(
    "GOOGLE_API_KEY=$GOOGLE_API_KEY"
    "DATABASE_HOST=$DB_HOST"
    "DATABASE_USER=papabase"
    "DATABASE_PASSWORD=$DB_PASSWORD"
    "DATABASE_NAME=fullstackarkham"
    "DATABASE_SSL_MODE=require"
    "REDIS_HOST=$REDIS_HOST"
    "REDIS_PORT=6379"
    "REDIS_PASSWORD=$REDIS_KEY"
)

# For clean reliable deployment, we delete and recreate
az containerapp delete --name gateway --resource-group $RESOURCE_GROUP --yes &>/dev/null || true

echo "Creating Gateway..."
az containerapp create \
  --name gateway \
  --resource-group $RESOURCE_GROUP \
  --environment ${APP_NAME}-env \
  --image ${ACR_NAME}.azurecr.io/gateway:latest \
  --registry-server ${ACR_NAME}.azurecr.io \
  --registry-username $ACR_USER \
  --registry-password "$ACR_PASS" \
  --target-port 8080 \
  --ingress external \
  --cpu 0.5 \
  --memory 1.0 \
  --min-replicas 1 \
  --env-vars "${GATEWAY_ENV[@]}" \
  --output table

GATEWAY_URL=$(az containerapp show -n gateway -g $RESOURCE_GROUP --query "properties.configuration.ingress.fqdn" -o tsv)
echo "✓ Gateway deployed at: https://$GATEWAY_URL"

# Deploy Container App
echo "Deploying Papabase to Container Apps..."
ENV_VARS=(
    "DATABASE_HOST=$DB_HOST"
    "DATABASE_USER=papabase"
    "DATABASE_PASSWORD=$DB_PASSWORD"
    "DATABASE_NAME=fullstackarkham"
    "REDIS_HOST=$REDIS_HOST"
    "REDIS_PORT=6379"
    "REDIS_PASSWORD=$REDIS_KEY"
    "JWT_SECRET=$JWT_SECRET"
    "GATEWAY_URL=https://$GATEWAY_URL"
    "STRIPE_SECRET_KEY=$STRIPE_SECRET_KEY"
    "STRIPE_WEBHOOK_SECRET=$STRIPE_WEBHOOK_SECRET"
)

az containerapp delete --name ${APP_NAME}-api --resource-group $RESOURCE_GROUP --yes &>/dev/null || true

echo "Creating Papabase API..."
az containerapp create \
  --name ${APP_NAME}-api \
  --resource-group $RESOURCE_GROUP \
  --environment ${APP_NAME}-env \
  --image ${ACR_NAME}.azurecr.io/papabase:latest \
  --registry-server ${ACR_NAME}.azurecr.io \
  --registry-username $ACR_USER \
  --registry-password "$ACR_PASS" \
  --target-port 8087 \
  --ingress external \
  --cpu 0.5 \
  --memory 1.0 \
  --min-replicas 1 \
  --env-vars "${ENV_VARS[@]}" \
  --output table

# Get the app URL
echo ""
echo "Getting your API URL..."
sleep 10
APP_URL=$(az containerapp show \
  --name ${APP_NAME}-api \
  --resource-group $RESOURCE_GROUP \
  --query "properties.configuration.ingress.fqdn" \
  --output tsv)

echo ""
echo "=============================================="
echo "  🎉 DEPLOYMENT COMPLETE!"
echo "=============================================="
echo ""
echo "Your Papabase API is running at:"
echo "  https://$APP_URL"
echo ""
echo "Test it:"
echo "  curl https://$APP_URL/health"
echo ""
echo "Stripe Webhook URL:"
echo "  https://$APP_URL/api/v1/billing/webhook"
echo ""
echo "DNS Configuration:"
echo "  Add this CNAME record:"
echo "    Name: api"
echo "    Value: $APP_URL"
echo ""

# Save summary
cat > azure-deployment-summary.txt << EOF
============================================
Papabase Azure Deployment Complete!
============================================
Date: $(date)

Resource Group: $RESOURCE_GROUP
Location: $LOCATION

API URL: https://$APP_URL
Gateway URL: https://$GATEWAY_URL

Database:
- Host: $DB_HOST
- User: papabase
- Password: [redacted]

Redis:
- Host: $REDIS_HOST
- Port: 6379
- Password: [redacted]

Stripe Configuration:
- Secret Key: [redacted]
- Webhook Secret: [redacted]
- Webhook URL: https://$APP_URL/api/v1/billing/webhook

DNS Configuration:
Type: CNAME
Name: api
Value: $APP_URL

Next Steps:
1. Test: curl https://$APP_URL/health
2. Configure DNS for fsai.pro
3. Set up Stripe webhook
4. Deploy frontend

============================================
EOF

echo "Summary saved to: azure-deployment-summary.txt"
echo ""
