# 🚀 Run Papabase Azure Deployment - Step by Step

**Follow these steps on your Azure-connected machine**

---

## 📋 Step 1: Install Azure CLI

### On macOS:
```bash
brew update && brew install azure-cli
```

### On Linux (Ubuntu/Debian):
```bash
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
```

### On Windows:
```powershell
# Run in PowerShell as Administrator
Invoke-WebRequest -Uri https://aka.ms/installazurecliwindows -OutFile AzureCLI.msi
Start-Process msiexec.exe -Wait -ArgumentList '/I AzureCLI.msi /quiet'
```

---

## 📋 Step 2: Login to Azure

```bash
az login
```

This will open a browser window. Sign in with your Azure account.

If you have multiple subscriptions:
```bash
az account list --output table
az account set --subscription "YOUR_SUBSCRIPTION_ID"
```

---

## 📋 Step 3: Navigate to Project

```bash
cd ~/Desktop/DevStudio/stack-arkham
```

---

## 📋 Step 4: Run Deployment Agent

```bash
chmod +x deploy-to-azure.sh
./deploy-to-azure.sh
```

**⏱️ This will take ~15 minutes**

The script will:
1. Create Resource Group
2. Provision Azure Container Registry
3. Create PostgreSQL Database
4. Create Redis Cache
5. Create Key Vault & store secrets
6. Build Docker images
7. Deploy to Container Apps
8. Output DNS configuration

---

## 📋 Step 5: Watch Deployment Progress

You'll see output like:

```
==============================================
  Papabase Azure Deployment Agent
==============================================

Resource Group: papabase-rg
Location: eastus
App Name: papabase
Domain: fsai.pro

[1/10] Creating Resource Group...
✓ Resource Group created

[2/10] Creating Azure Container Registry...
✓ ACR created: papabaseacrXXXX.azurecr.io

[3/10] Creating Azure Database for PostgreSQL...
✓ PostgreSQL created: papabase-psql

[4/10] Creating Azure Redis Cache...
✓ Redis created: papabase-redis

[5/10] Creating Azure Key Vault...
✓ Key Vault created: papabase-kvXXXX
✓ Secrets stored

[6/10] Building and pushing Docker images...
✓ Building gateway...
✓ Building papabase...
✓ Building billing...
...

[7/10] Creating Container Apps Environment...
✓ Environment created

[8/10] Deploying Papabase...
✓ Papabase deployed

Papabase API URL: https://papabase-api.eastus.azurecontainerapps.io

[9/10] DNS Configuration

Add these DNS records to your domain registrar:

Type: CNAME
Name: api
Value: papabase-api.eastus.azurecontainerapps.io

Type: A
Name: @
Value: [Your Frontend IP]

[10/10] Creating deployment summary...
✓ Summary saved to azure-deployment-summary.txt

==============================================
  DEPLOYMENT COMPLETE!
==============================================
```

---

## 📋 Step 6: Configure DNS

### In Hostinger hPanel:

1. Go to: https://hpanel.hostinger.com
2. **Domains** → **fsai.pro** → **DNS/Namespace**
3. **Add DNS Records**:

| Type | Name | Value | TTL |
|------|------|-------|-----|
| `CNAME` | `api` | `[Container App URL from output]` | Auto |
| `A` | `@` | `[Your frontend IP]` | Auto |
| `CNAME` | `www` | `fsai.pro` | Auto |

4. Click **Save**

---

## 📋 Step 7: Set Up Stripe Webhook

1. **Copy webhook URL** from deployment output:
   ```
   https://[YOUR_CONTAINER_APP_URL]/api/v1/billing/webhook
   ```

2. **Go to Stripe Dashboard**: https://dashboard.stripe.com/webhooks

3. **Click "Add endpoint"**

4. **Enter webhook URL**:
   ```
   https://[YOUR_CONTAINER_APP_URL]/api/v1/billing/webhook
   ```

5. **Select events**:
   - ✅ `checkout.session.completed`
   - ✅ `customer.subscription.created`
   - ✅ `customer.subscription.updated`
   - ✅ `customer.subscription.deleted`
   - ✅ `invoice.payment_succeeded`
   - ✅ `invoice.payment_failed`

6. **Click "Add endpoint"**

7. **Copy the Signing Secret** (starts with `whsec_`)

8. **Update Key Vault** (optional - already configured):
   ```bash
   az keyvault secret set \
     --vault-name papabase-kvXXXX \
     --name "STRIPE-WEBHOOK-SECRET" \
     --value "whsec_REDACTED"
   ```

---

## 📋 Step 8: Test Your Deployment

### Test API Health:
```bash
curl https://api.fsai.pro/health
```

### Test Pricing Endpoint:
```bash
curl https://api.fsai.pro/api/v1/pricing/plans
```

### View Logs:
```bash
az containerapp logs show \
  --name papabase-api \
  --resource-group papabase-rg \
  --follow
```

---

## 📋 Step 9: Deploy Frontend (Optional)

### Option A: Azure Static Web Apps

```bash
# Install SWA CLI
npm install -g @azure/static-web-apps-cli

# Build frontend
cd apps/web
npm install
npm run build

# Deploy
swa deploy ./out --env production
```

### Option B: Keep on Hostinger

Upload `apps/web/out/*` to Hostinger via FTP.

---

## 📋 Step 10: Check Deployment Summary

```bash
cat azure-deployment-summary.txt
```

This file contains:
- All resource names
- Connection strings
- URLs
- DNS configuration
- Stripe webhook URL

---

## 🆘 Troubleshooting

### Azure CLI Not Found
```bash
# Install Azure CLI first
# See Step 1 above
```

### Login Fails
```bash
# Try device code authentication
az login --use-device-code
```

### Deployment Fails
```bash
# Check Azure Portal for errors
# https://portal.azure.com
# Resource groups → papabase-rg → Deployments

# View activity log
az monitor activity-log list \
  --resource-group papabase-rg \
  --max-events 20
```

### Can't Access API After Deployment
```bash
# Wait 5-10 minutes for DNS propagation
# Check container app status
az containerapp show \
  --name papabase-api \
  --resource-group papabase-rg

# View logs
az containerapp logs show \
  --name papabase-api \
  --resource-group papabase-rg \
  --follow
```

---

## ✅ Success Indicators

You'll know deployment succeeded when:

- ✅ `azure-deployment-summary.txt` is created
- ✅ You see "DEPLOYMENT COMPLETE!" message
- ✅ Container App URL is displayed
- ✅ All 10/10 steps complete without errors
- ✅ DNS records are shown

---

## 📞 Need Help?

**During Deployment:**
- Watch console output for errors
- Check Azure Portal → Resource Group → papabase-rg
- View deployment logs in Azure Portal

**After Deployment:**
- Test endpoints with curl
- Check Container App logs
- Monitor in Azure Portal

**Stripe Issues:**
- Check webhook logs in Stripe Dashboard
- Verify webhook secret in Key Vault
- Test with Stripe CLI locally

---

## 💰 Cost Tracking

Monitor your Azure costs:
```bash
az consumption budget list
az consumption usage list --start-date 2024-01-01
```

Or check: Azure Portal → Cost Management + Billing

---

**Ready to deploy? Start with Step 1!** 🚀

```bash
# Install Azure CLI
brew install azure-cli  # macOS
# or
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash  # Linux

# Then run deployment
./deploy-to-azure.sh
```
