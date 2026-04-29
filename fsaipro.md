# 🚀 DEPLOY PAPABASE TO fsai.pro - FINAL CHECKLIST

**Status**: ✅ Stripe Configured | ✅ Products Created | ✅ Ready to Deploy

---

## ✅ Pre-Deployment Checklist

- [x] **Stripe Account**: Live mode active
- [x] **Stripe Key Rotated**: New key configured
- [x] **Products Created**: All 4 tiers with new key
- [x] **Webhook Secret**: `whsec_REDACTED`
- [x] **Frontend Updated**: New Price IDs in PricingView.tsx
- [ ] **Domain**: fsai.pro DNS configured
- [ ] **Hosting**: VPS or Hybrid ready

---

## 🔑 Your Stripe Configuration (Updated)

### Secret Key (Rotated)
```
sk_live_REDACTED
```

### Webhook Secret
```
whsec_REDACTED
```

### New Price IDs
| Plan | Monthly | Yearly |
|------|---------|--------|
| Starter | `price_1TR6gt6X8IBUtLKfsGUeNliQ` | `price_1TR6gt6X8IBUtLKf2qc2Yxci` |
| Studio | `price_1TR6gu6X8IBUtLKffqgXoEWO` | `price_1TR6gu6X8IBUtLKffUhXngpr` |
| Agency | `price_1TR6gu6X8IBUtLKfLx7JOVKL` | `price_1TR6gv6X8IBUtLKfOtPyJMh7` |

---

## 📋 DEPLOYMENT OPTION 1: VPS (Recommended)

### Step 1: SSH Into Your VPS
```bash
ssh root@fsai.pro
# Or your username if different
```

### Step 2: Run Deployment Script
```bash
# Download script
curl -fsSL https://raw.githubusercontent.com/robs46859-eng/stack-arkham/main/deploy-vps.sh -o deploy.sh

# Make executable
chmod +x deploy.sh

# Run deployment
sudo ./deploy.sh
```

### Step 3: Configure Environment
The script will prompt you to edit `.env`. Update these values:
```bash
nano /opt/stack-arkham/.env

# Update:
DATABASE_PASSWORD=YourSecurePassword123!
JWT_SECRET=$(openssl rand -base64 64)
STRIPE_SECRET_KEY=sk_live_REDACTED
STRIPE_WEBHOOK_SECRET=whsec_REDACTED
GOOGLE_API_KEY=Your_Gemini_API_Key_Here
```

### Step 4: Start Services
```bash
cd /opt/stack-arkham
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f
```

### Step 5: Configure DNS
Point fsai.pro to your VPS IP:
```
Type: A
Name: @
Value: YOUR_VPS_IP

Type: A  
Name: www
Value: YOUR_VPS_IP

Type: A
Name: api
Value: YOUR_VPS_IP
```

### Step 6: Set Up Stripe Webhook
1. Go to https://dashboard.stripe.com/webhooks
2. Add endpoint: `https://fsai.pro/api/v1/billing/webhook`
3. Select events: `checkout.session.completed`, `customer.subscription.*`, `invoice.*`
4. Copy signing secret (already have it: `whsec_REDACTED`)

---

## 📋 DEPLOYMENT OPTION 2: Hybrid (Shared Hosting + Cloud)

### Part A: Frontend on Hostinger

```bash
# 1. Build static export
cd ~/Desktop/DevStudio/stack-arkham/apps/web
npm install
npm run build

# 2. Upload to Hostinger via FTP
# Host: ftp.fsai.pro
# Username: your_username
# Password: your_password
# 
# Upload contents of /out folder to:
# Remote: /public_html/
```

### Part B: Backend on Railway

```bash
# Install Railway CLI
npm i -g @railway/cli

# Login
railway login

# Deploy Papabase service
cd ~/Desktop/DevStudio/stack-arkham/services/papabase
railway init
railway up

# Set environment variables
railway variables set STRIPE_SECRET_KEY=sk_live_REDACTED
railway variables set STRIPE_WEBHOOK_SECRET=whsec_REDACTED
```

### Part C: Database on Neon

```bash
# 1. Go to https://neon.tech
# 2. Create free project
# 3. Copy connection string
# 4. Add to Railway:
railway variables set DATABASE_URL=postgresql://...
```

---

## 🧪 Test Your Deployment

### After deploying, verify:

```bash
# 1. Check frontend loads
curl https://fsai.pro

# 2. Check API works
curl https://api.fsai.pro/health

# 3. Check pricing endpoint
curl https://api.fsai.pro/api/v1/pricing/plans

# 4. Test Stripe checkout
# Go to fsai.pro/dashboard/pricing
# Click "Upgrade" on any plan
# Use test card: 4242 4242 4242 4242
```

---

## 🔒 SSL/HTTPS Setup

### For VPS (Caddy auto-SSL):
Caddy will automatically get SSL certificate from Let's Encrypt.
Just ensure ports 80 and 443 are open.

### For Hostinger Shared:
1. hPanel → SSL
2. Install free Let's Encrypt
3. Enable "Force HTTPS"

---

## 📊 Monitor Your Deployment

### Check Service Health
```bash
# VPS
docker compose ps
docker compose logs -f papabase
docker compose logs -f billing

# Railway
railway logs
```

### Check Stripe Payments
- Dashboard: https://dashboard.stripe.com/test/payments
- Switch to "Live Mode" when ready

---

## 🆘 Troubleshooting

### Services Won't Start
```bash
# Check Docker status
docker compose ps

# View logs
docker compose logs -f

# Restart
docker compose restart
```

### Can't Access fsai.pro
```bash
# Check DNS propagation
nslookup fsai.pro

# Wait up to 24 hours for DNS to propagate
# Or use CloudFlare for faster propagation
```

### Stripe Payment Fails
```bash
# Check webhook logs
# Stripe Dashboard → Developers → Webhooks → Events

# Test webhook locally
stripe listen --forward-to localhost:8086/webhook
stripe trigger checkout.session.completed
```

---

## ✅ Post-Deployment Checklist

After deployment, verify:

- [ ] Frontend loads at https://fsai.pro
- [ ] Login/Signup works
- [ ] Onboarding questionnaire loads
- [ ] Dashboard accessible
- [ ] Pricing page shows correct prices
- [ ] "Upgrade" button redirects to Stripe
- [ ] Test payment completes successfully
- [ ] Webhook events received (check Stripe Dashboard)
- [ ] Dad AI website generation works
- [ ] CRM leads/tasks functional

---

## 📞 Support Files

Created during this deployment:
- `DEPLOY_NOW.md` - Main deployment guide
- `DEPLOYMENT_HOSTINGER.md` - Hostinger-specific guide
- `STRIPE_COMPLETE.md` - Stripe configuration
- `deploy-vps.sh` - Automated VPS deployment script
- `.env.production.example` - Environment template

---

## 🎉 You're Ready to Deploy!

**Choose your deployment method and run the commands above.**

Need help? Tell me:
- "VPS" - if deploying to Hostinger VPS
- "SHARED" - if using Hostinger shared hosting
- "RAILWAY" - if deploying backend to Railway

And I'll give you the exact next commands!
