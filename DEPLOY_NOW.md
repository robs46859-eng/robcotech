# 🚀 Papabase Deployment Guide - fsai.pro

**Domain**: fsai.pro  
**Host**: Hostinger  
**Status**: Ready to Deploy

---

## ⚡ Quick Deploy Options

### Option 1: Hostinger VPS (Recommended - Full Control)

If you have Hostinger VPS hosting, run this single command:

```bash
# SSH into your VPS
ssh your-username@fsai.pro

# Download and run deployment script
curl -O https://raw.githubusercontent.com/robs46859-eng/stack-arkham/main/deploy-vps.sh
chmod +x deploy-vps.sh
sudo ./deploy-vps.sh
```

This will:
- ✅ Install Docker & Docker Compose
- ✅ Clone Papabase repository
- ✅ Set up PostgreSQL, Redis, all services
- ✅ Configure environment variables
- ✅ Start all 11 services

**Time**: ~10 minutes  
**Cost**: $8.99/mo (Hostinger VPS)

---

### Option 2: Hostinger Shared Hosting (Hybrid Approach)

If you have shared hosting, we'll split the deployment:

| Component | Where | Cost |
|-----------|-------|------|
| Frontend (Next.js) | Hostinger Shared | $0 (included) |
| Backend API | Railway.app | $5/mo |
| Database | Neon | Free |
| Redis | Upstash | Free |

#### Step 1: Deploy Frontend to Hostinger

```bash
# Build static export
cd ~/Desktop/DevStudio/stack-arkham/apps/web
npm install
npm run build

# Upload to Hostinger via FTP
# 1. Go to hPanel → File Manager
# 2. Navigate to public_html/
# 3. Upload contents of /out folder
```

#### Step 2: Deploy Backend to Railway

```bash
# Install Railway CLI
npm i -g @railway/cli

# Login
railway login

# Deploy each service
cd services/gateway && railway init && railway up
cd services/billing && railway init && railway up
cd services/papabase && railway init && railway up
```

#### Step 3: Set Up Database

```bash
# Go to https://neon.tech
# Create free PostgreSQL database
# Copy connection string
# Add to Railway environment variables
```

**Time**: ~30 minutes  
**Cost**: ~$5/mo

---

## 📋 Pre-Deployment Checklist

Before deploying, make sure you have:

- [ ] **Stripe Account**: ✅ Done (products created)
- [ ] **Stripe Webhook**: Set up at https://dashboard.stripe.com/webhooks
  - Endpoint: `https://fsai.pro/api/v1/billing/webhook`
  - Events: `checkout.session.completed`, `customer.subscription.*`, `invoice.*`
- [ ] **Google AI API Key**: Get from https://makersuite.google.com/app/apikey
- [ ] **Domain DNS**: fsai.pro pointing to your hosting
- [ ] **SSL Certificate**: Auto-installed on Hostinger

---

## 🔐 Environment Variables

### For VPS Deployment (.env file)

```bash
# Database
DATABASE_HOST=localhost
DATABASE_USER=papabase
DATABASE_PASSWORD=YOUR_SECURE_PASSWORD_HERE
DATABASE_NAME=papabase

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT Auth
JWT_SECRET=GENERATE_RANDOM_64_CHARS_HERE

# Stripe (already configured)
STRIPE_SECRET_KEY=sk_live_REDACTED
STRIPE_WEBHOOK_SECRET=whsec_REDACTED_WEBHOOK_SECRET_HERE

# Google AI (for Dad AI)
GOOGLE_API_KEY=YOUR_GEMINI_API_KEY_HERE

# Domain
FRONTEND_URL=https://fsai.pro
API_URL=https://api.fsai.pro
```

---

## 🌐 DNS Configuration

### For VPS:

Point your domain to VPS IP:

1. **Get VPS IP**: Check Hostinger dashboard
2. **Update DNS** (at Hostinger or where domain is registered):
   ```
   Type: A
   Name: @
   Value: YOUR_VPS_IP
   TTL: Automatic
   ```
   
   ```
   Type: A
   Name: www
   Value: YOUR_VPS_IP
   TTL: Automatic
   ```
   
   ```
   Type: A
   Name: api
   Value: YOUR_VPS_IP
   TTL: Automatic
   ```

### For Shared Hosting:

DNS is already configured. Just upload files via FTP.

---

## 🔒 SSL/HTTPS Setup

### Hostinger VPS (with Caddy):

Caddy auto-configures SSL. Just ensure port 443 is open.

### Hostinger Shared:

1. hPanel → SSL
2. Install free Let's Encrypt certificate
3. Enable "Force HTTPS"

---

## 📊 Post-Deployment Verification

After deployment, verify everything is working:

```bash
# Check services are running
docker compose ps

# Test health endpoints
curl https://fsai.pro/health
curl https://api.fsai.pro/health
curl https://api.fsai.pro/api/v1/pricing/plans

# Test Stripe webhook
stripe trigger checkout.session.completed --url https://fsai.pro/api/v1/billing/webhook
```

---

## 🆘 Troubleshooting

### Services Won't Start
```bash
# Check logs
docker compose logs -f

# Restart services
docker compose restart

# Rebuild images
docker compose build
docker compose up -d
```

### Domain Not Working
```bash
# Check DNS propagation
nslookup fsai.pro
# or use: https://dnschecker.org/

# Wait up to 24 hours for full propagation
```

### Stripe Webhook Failing
```bash
# Check webhook logs in Stripe Dashboard
# Verify STRIPE_WEBHOOK_SECRET is set correctly
# Test locally first with stripe listen
```

---

## 📞 Support Resources

- **Full Deployment Guide**: `DEPLOYMENT_HOSTINGER.md`
- **Stripe Setup**: `STRIPE_SETUP.md`
- **VPS Script**: `deploy-vps.sh`
- **Environment Template**: `.env.production.example`

---

## ✅ Deployment Complete!

After deployment, your Papabase will be live at:

- **Frontend**: https://fsai.pro
- **API**: https://api.fsai.pro
- **Dashboard**: https://fsai.pro/dashboard

**Monthly Costs:**
- Hostinger VPS: $8.99
- Stripe: 2.9% + 30¢ per transaction
- Google AI: ~$0.50 per 1000 requests
- **Total**: ~$10-15/mo to start

---

**Ready to deploy? Choose Option 1 (VPS) or Option 2 (Hybrid) above!**
