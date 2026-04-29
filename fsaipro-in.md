# Papabase Deployment Guide - Hostinger (fsai.pro)

## Prerequisites

1. **Domain**: fsai.pro (already on Hostinger)
2. **Stripe Account**: For payment processing
3. **Google Cloud Account**: For Gemini API (Dad AI)
4. **VPS or Cloud Hosting**: Hostinger shared hosting won't work for the full stack

---

## ⚠️ Important: Hosting Requirements

**Papabase CANNOT run on Hostinger shared hosting** because it requires:
- Docker containers (11 services)
- PostgreSQL with pgvector
- Redis
- Python 3.11+ and Go 1.25+
- Long-running processes

**Recommended Options:**

### Option A: VPS (Recommended for Full Stack)
- **Hostinger VPS** ($8.99/mo) or
- **DigitalOcean Droplet** ($12/mo) or
- **Linode** ($10/mo)

Minimum specs:
- 4GB RAM
- 2 CPU cores
- 80GB SSD
- Docker support

### Option B: Hybrid (Cheaper)
- **Frontend**: Hostinger shared hosting (static Next.js export)
- **Backend**: Cloud Run / Railway / Render
- **Database**: Neon (free PostgreSQL)

---

## 🔧 Stripe Setup

### 1. Create Stripe Account
1. Go to https://stripe.com
2. Create account for fsai.pro
3. Get API keys from Dashboard → Developers → API keys

### 2. Configure Stripe Products & Prices

Run this script to create products:

```python
# scripts/create_stripe_products.py
import stripe

stripe.api_key = "sk_test_REDACTED"

# Create products
products = {
    "starter": {"name": "Papabase Starter", "price": 2900},  # $29 in cents
    "studio": {"name": "Papabase Studio", "price": 9900},    # $99
    "agency": {"name": "Papabase Agency", "price": 29900},   # $299
}

for key, data in products.items():
    product = stripe.Product.create(name=data["name"])
    price = stripe.Price.create(
        product=product.id,
        unit_amount=data["price"],
        currency="usd",
        recurring={"interval": "month"},
    )
    print(f"{key}: price_id={price.id}")
```

### 3. Update Environment Variables

```bash
# services/billing/.env
STRIPE_SECRET_KEY=sk_live_...  # Use live key for production
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_ID_STARTER=price_...
STRIPE_PRICE_ID_STUDIO=price_...
STRIPE_PRICE_ID_AGENCY=price_...
```

### 4. Configure Webhook

In Stripe Dashboard → Developers → Webhooks:
- Endpoint: `https://fsai.pro/api/v1/billing/webhook`
- Events: `invoice.payment_succeeded`, `customer.subscription.updated`, `payment_intent.succeeded`

---

## 🚀 Deployment Options

### Option 1: VPS Deployment (Full Control)

```bash
# SSH into your VPS
ssh root@fsai.pro

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Clone repo
git clone https://github.com/your-org/stack-arkham.git
cd stack-arkham

# Create .env file
cat > .env << EOF
# Database
DATABASE_HOST=postgres
DATABASE_USER=papabase
DATABASE_PASSWORD=YOUR_SECURE_PASSWORD
DATABASE_NAME=papabase

# Stripe
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...

# Google AI (for Dad AI)
GOOGLE_API_KEY=YOUR_GEMINI_API_KEY

# JWT
JWT_SECRET=YOUR_RANDOM_SECRET_64_CHARS
EOF

# Start services
docker compose up -d

# Check status
docker compose ps
```

### Option 2: Hybrid Deployment (Cheaper)

#### Frontend on Hostinger Shared Hosting

```bash
# Build static export
cd apps/web
npm run build

# Upload to Hostinger via FTP
# Upload contents of /out folder to public_html/
```

#### Backend on Cloud Services

**Railway.app** (easiest):
```bash
# Install Railway CLI
npm i -g @railway/cli

# Deploy each service
cd services/gateway && railway up
cd services/billing && railway up
# etc.
```

**Google Cloud Run**:
```bash
# Build and deploy each service
cd services/gateway
gcloud builds submit --tag gcr.io/PROJECT_ID/gateway
gcloud run deploy gateway --image gcr.io/PROJECT_ID/gateway
```

---

## 🔐 SSL/HTTPS Setup

### For VPS (Docker):
```yaml
# Add to docker-compose.yml
services:
  caddy:
    image: caddy:2-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
```

```caddy
# Caddyfile
fsai.pro {
    reverse_proxy gateway:8080
}

api.fsai.pro {
    reverse_proxy gateway:8080
}
```

### For Hostinger:
1. hPanel → SSL → Install SSL Certificate (free Let's Encrypt)
2. Force HTTPS in hPanel → SSL → HTTPS Enforce

---

## 📊 Production Checklist

### Before Going Live:

- [ ] Change all default passwords
- [ ] Generate secure JWT_SECRET (64 chars)
- [ ] Use Stripe LIVE keys (not test)
- [ ] Set up Google Cloud billing for Gemini API
- [ ] Configure backup strategy (daily DB backups)
- [ ] Set up monitoring (UptimeKuma, Grafana)
- [ ] Configure email sending (SendGrid/Resend)
- [ ] Set up domain email (contact@fsai.pro)
- [ ] Test payment flow end-to-end
- [ ] Configure rate limiting
- [ ] Set up error tracking (Sentry)

### Environment Variables (Production):

```bash
# Required for all services
DATABASE_URL=postgresql://user:pass@host:5432/papabase
REDIS_URL=redis://host:6379
JWT_SECRET=generate_openssl_rand_base64_64
GOOGLE_API_KEY=AIza...

# Stripe
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...

# Email (for notifications)
SENDGRID_API_KEY=SG....
FROM_EMAIL=noreply@fsai.pro

# Domain
FRONTEND_URL=https://fsai.pro
API_URL=https://api.fsai.pro
```

---

## 💰 Cost Estimate

| Service | Monthly Cost |
|---------|-------------|
| VPS (4GB) | $8-12 |
| Domain | $15/year |
| Stripe | 2.9% + 30¢ per transaction |
| Google AI | ~$0.0005/1K tokens |
| Email (SendGrid) | Free up to 100/day |
| **Total** | **~$15-25/mo** |

---

## 🆘 Quick Start Commands

```bash
# Check if services are running
docker compose ps

# View logs
docker compose logs -f papabase
docker compose logs -f gateway

# Restart services
docker compose restart

# Database backup
docker exec postgres pg_dump -U postgres papabase > backup.sql

# Restore database
docker exec -i postgres psql -U postgres < backup.sql
```

---

## 📞 Support

For issues:
1. Check logs: `docker compose logs`
2. Test API: `curl https://api.fsai.pro/health`
3. Check Stripe webhook logs in Dashboard
