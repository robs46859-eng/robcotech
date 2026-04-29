# Papabase Azure Deployment Agent

**Autonomous deployment agent for Papabase to Microsoft Azure**

This agent automates the complete deployment of Papabase (fsai.pro) to Azure with production-ready infrastructure.

---

## 🤖 What This Agent Does

1. **Provisions Azure Infrastructure**
   - Resource Group
   - Azure Container Registry (ACR)
   - Azure Database for PostgreSQL (Flexible Server)
   - Azure Redis Cache
   - Azure Key Vault (secrets)
   - Azure Container Apps Environment
   - Azure Application Gateway (SSL)

2. **Builds & Deploys Containers**
   - Papabase API
   - Gateway Service
   - Billing Service
   - Arkham Security
   - All backend services

3. **Configures Everything**
   - Environment variables
   - Database schema
   - SSL certificates
   - Custom domain (fsai.pro)
   - Stripe webhook

---

## 📋 Prerequisites

```bash
# 1. Install Azure CLI
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

# 2. Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 3. Install Terraform (for IaC deployment)
wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/hashicorp.list
apt update && apt install terraform

# 4. Login to Azure
az login

# 5. Set subscription
az account set --subscription "YOUR_SUBSCRIPTION_ID"
```

---

## 🚀 Quick Deploy (Automated)

### Option 1: Bash Script (Fastest)

```bash
# Clone repository
cd ~/Desktop/DevStudio/stack-arkham

# Make script executable
chmod +x deploy-to-azure.sh

# Run deployment
./deploy-to-azure.sh

# Deployment takes ~15 minutes
# Coffee break! ☕
```

### Option 2: Terraform (Infrastructure as Code)

```bash
# Navigate to Terraform directory
cd ~/Desktop/DevStudio/stack-arkham/infra/terraform/azure

# Initialize Terraform
terraform init

# Review execution plan
terraform plan

# Apply infrastructure
terraform apply

# Build and push images
./deploy-images.sh
```

---

## 📊 Deployment Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Azure Front Door                      │
│                    (Global Load Balancer)                │
│                    fsai.pro                              │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│              Azure Application Gateway                   │
│              (SSL Termination, WAF)                      │
└─────────────────────────────────────────────────────────┘
                            │
            ┌───────────────┴───────────────┐
            ▼                               ▼
┌───────────────────────┐       ┌───────────────────────┐
│  Azure Static Web     │       │  Azure Container      │
│  Apps (Frontend)      │       │  Apps (Backend)       │
│  - Next.js            │       │  - Papabase API       │
│  - React              │       │  - Gateway            │
│                       │       │  - Billing            │
└───────────────────────┘       └───────────────────────┘
                                            │
                    ┌───────────────────────┼───────────────────────┐
                    ▼                       ▼                       ▼
        ┌───────────────────┐   ┌───────────────────┐   ┌───────────────────┐
        │   Azure Database  │   │   Azure Redis     │   │   Azure Key       │
        │   PostgreSQL      │   │   Cache           │   │   Vault           │
        │   (Flexible)      │   │                   │   │   (Secrets)       │
        └───────────────────┘   └───────────────────┘   └───────────────────┘
```

---

## 🔐 Configuration

### Environment Variables

Create `azure.env` with your configuration:

```bash
# Azure Configuration
AZURE_RESOURCE_GROUP="papabase-rg"
AZURE_LOCATION="eastus"
AZURE_SUBSCRIPTION_ID="your-subscription-id"

# Application Configuration
APP_NAME="papabase"
DOMAIN="fsai.pro"
FRONTEND_DOMAIN="www.fsai.pro"

# Stripe Configuration (Already Set)
STRIPE_SECRET_KEY="sk_live_REDACTED"
STRIPE_WEBHOOK_SECRET="whsec_REDACTED"

# Google AI (Get from https://makersuite.google.com/app/apikey)
GOOGLE_API_KEY="YOUR_GEMINI_API_KEY"

# Database Configuration
DB_ADMIN_USER="papabase"
DB_ADMIN_PASSWORD="GENERATE_SECURE_PASSWORD"

# Redis Configuration
REDIS_SKU="Basic"
REDIS_SIZE="C0"
```

---

## 📁 Files Created by Agent

```
stack-arkham/
├── deploy-to-azure.sh              # Main deployment script
├── azure-deployment-summary.txt    # Deployment output
├── infra/
│   └── terraform/
│       └── azure/
│           ├── main.tf             # Azure resources
│           ├── variables.tf        # Input variables
│           ├── outputs.tf          # Deployment outputs
│           └── deploy-images.sh    # Image build script
└── services/
    └── papabase/
        └── Dockerfile              # Container image
```

---

## 🎯 Deployment Steps (Detailed)

### Step 1: Create Resource Group

```bash
az group create \
  --name papabase-rg \
  --location eastus
```

### Step 2: Create Azure Container Registry

```bash
az acr create \
  --resource-group papabase-rg \
  --name papabaseacr$(openssl rand -hex 4) \
  --sku Basic \
  --admin-enabled true
```

### Step 3: Create PostgreSQL Database

```bash
az postgres flexible-server create \
  --resource-group papabase-rg \
  --name papabase-psql \
  --location eastus \
  --sku-name Standard_B1ms \
  --admin-user papabase \
  --admin-password "SecurePass123!" \
  --public-access all
```

### Step 4: Create Redis Cache

```bash
az redis create \
  --location eastus \
  --name papabase-redis \
  --resource-group papabase-rg \
  --sku Basic \
  --vm-size C0
```

### Step 5: Create Key Vault & Store Secrets

```bash
# Create Key Vault
az keyvault create \
  --resource-group papabase-rg \
  --name papabase-kv$(openssl rand -hex 4)

# Store Stripe secrets
az keyvault secret set \
  --vault-name papabase-kvXXXX \
  --name "STRIPE-SECRET-KEY" \
  --value "sk_live_REDACTED"

az keyvault secret set \
  --vault-name papabase-kvXXXX \
  --name "STRIPE-WEBHOOK-SECRET" \
  --value "whsec_REDACTED"
```

### Step 6: Build & Push Docker Images

```bash
# Login to ACR
az acr login --name papabaseacr

# Build Papabase
cd services/papabase
az acr build \
  --registry papabaseacr \
  --image papabase:latest \
  .
```

### Step 7: Deploy to Container Apps

```bash
# Create environment
az containerapp env create \
  --name papabase-env \
  --resource-group papabase-rg \
  --location eastus

# Deploy app
az containerapp create \
  --name papabase-api \
  --resource-group papabase-rg \
  --environment papabase-env \
  --image papabaseacr.azurecr.io/papabase:latest \
  --target-port 8087 \
  --ingress external \
  --cpu 0.5 \
  --memory 1.0 \
  --min-replicas 1
```

### Step 8: Configure DNS

Add these records to your domain registrar:

```
Type: CNAME
Name: api
Value: [YOUR_CONTAINER_APP_URL]

Type: A
Name: @
Value: [YOUR_FRONTEND_IP]
```

### Step 9: Set Up Stripe Webhook

1. Get your webhook URL from deployment output
2. Go to https://dashboard.stripe.com/webhooks
3. Add endpoint: `https://[YOUR_URL]/api/v1/billing/webhook`
4. Select events: `checkout.session.completed`, `customer.subscription.*`, `invoice.*`

---

## 📈 Post-Deployment

### Monitor Your Application

```bash
# View logs
az containerapp logs show \
  --name papabase-api \
  --resource-group papabase-rg \
  --follow

# Check health
curl https://api.fsai.pro/health

# Check metrics
az monitor metrics list \
  --resource [RESOURCE_ID] \
  --metric Requests \
  --interval PT1H
```

### Scale Your Application

```bash
# Scale up (more CPU/RAM)
az containerapp update \
  --name papabase-api \
  --resource-group papabase-rg \
  --cpu 1.0 \
  --memory 2.0

# Scale out (more replicas)
az containerapp update \
  --name papabase-api \
  --resource-group papabase-rg \
  --min-replicas 2 \
  --max-replicas 10
```

---

## 💰 Cost Estimate

| Service | Configuration | Monthly Cost |
|---------|--------------|-------------|
| Container Apps | 0.5 CPU, 1GB RAM | ~$15 |
| PostgreSQL | Burstable B1ms, 32GB | ~$25 |
| Redis Cache | Basic C0 (250MB) | ~$16 |
| Container Registry | Basic | ~$5 |
| Key Vault | Standard | ~$1 |
| Static Web Apps | Free tier | $0 |
| **Total** | | **~$62/mo** |

---

## 🆘 Troubleshooting

### Container Won't Start

```bash
# Check logs
az containerapp logs show \
  --name papabase-api \
  --resource-group papabase-rg \
  --follow

# Common issues:
# 1. Database connection string incorrect
# 2. Environment variables missing
# 3. Image pull failed
```

### Can't Access API

```bash
# Check ingress is external
az containerapp show \
  --name papabase-api \
  --query "properties.configuration.ingress.external"

# Test health endpoint
curl https://[YOUR_URL]/health

# Check firewall
az postgres flexible-server firewall-rule list \
  --resource-group papabase-rg \
  --name papabase-psql
```

### Database Connection Fails

```bash
# Allow Azure services
az postgres flexible-server firewall-rule create \
  --resource-group papabase-rg \
  --name papabase-psql \
  --rule-name AllowAzureServices \
  --start-ip-address 0.0.0.0 \
  --end-ip-address 0.0.0.0
```

---

## ✅ Deployment Checklist

- [ ] Azure CLI installed and logged in
- [ ] Subscription set correctly
- [ ] Run `./deploy-to-azure.sh`
- [ ] Save deployment summary
- [ ] Configure DNS for fsai.pro
- [ ] Set up Stripe webhook
- [ ] Deploy frontend to Static Web Apps
- [ ] Test payment flow
- [ ] Monitor in Azure Portal

---

## 📞 Support

**Deployment Issues:**
- Check Azure Portal → Resource Group → papabase-rg
- View activity logs for errors
- Check Container App log stream

**Stripe Issues:**
- Check Stripe Dashboard → Developers → Webhooks
- Verify webhook secret in Key Vault
- Test with Stripe CLI locally

**Database Issues:**
- Check PostgreSQL server status
- Verify connection string
- Check firewall rules

---

**Ready to deploy? Run the agent!**

```bash
./deploy-to-azure.sh
```

**Estimated time: 15-20 minutes** ☕
