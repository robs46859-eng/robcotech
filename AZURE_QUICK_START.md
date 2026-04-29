# 🚀 Papabase Azure Quick Deploy

**Deploy Papabase to Azure in 3 commands**

---

## ⚡ Quick Start

```bash
# 1. Login to Azure
az login

# 2. Navigate to project
cd ~/Desktop/DevStudio/stack-arkham

# 3. Run deployment
./deploy-to-azure.sh
```

**That's it!** The agent handles everything else.

---

## 📊 What Gets Deployed

| Service | Azure Resource | Purpose |
|---------|---------------|---------|
| Papabase API | Container Apps | Main application |
| Gateway | Container Apps | AI model routing |
| Database | PostgreSQL Flexible | Data storage |
| Cache | Redis Cache | Session/performance |
| Secrets | Key Vault | Secure configuration |
| Registry | Container Registry | Docker images |

---

## ⏱️ Timeline

| Step | Time |
|------|------|
| Resource provisioning | 5 min |
| Database setup | 3 min |
| Build & push images | 5 min |
| Deploy containers | 2 min |
| **Total** | **~15 min** |

---

## 🎯 Post-Deployment

### Get Your API URL
```bash
az containerapp show \
  --name papabase-api \
  --resource-group papabase-rg \
  --query "properties.configuration.ingress.fqdn" \
  --output tsv
```

### Test Health
```bash
curl https://[YOUR_URL]/health
```

### View Logs
```bash
az containerapp logs show \
  --name papabase-api \
  --resource-group papabase-rg \
  --follow
```

---

## 🔧 Configure Stripe Webhook

1. **Get webhook URL** from deployment output
2. **Add to Stripe**: https://dashboard.stripe.com/webhooks
3. **Events**: `checkout.session.completed`, `customer.subscription.*`, `invoice.*`

---

## 💰 Estimated Cost

**~$62/month** for production setup

Breakdown:
- Container Apps: $15
- PostgreSQL: $25
- Redis: $16
- Registry + Key Vault: $6

---

## 🆘 Need Help?

**During Deployment:**
- Watch the console output
- Check Azure Portal → Resource Group → papabase-rg
- View activity logs for errors

**After Deployment:**
- Check logs: `az containerapp logs show --name papabase-api -g papabase-rg --follow`
- Test endpoints: `curl https://[YOUR_URL]/health`
- Monitor: Azure Portal → Container Apps

---

## 📁 Deployment Output

After completion, you'll get:
- `azure-deployment-summary.txt` - All URLs and credentials
- Resource Group: `papabase-rg`
- Container App URL
- Database connection string
- Redis connection string

---

**Ready? Run the deploy agent!**

```bash
./deploy-to-azure.sh
```
