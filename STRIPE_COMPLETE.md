# ✅ Stripe Setup Complete for Papabase (fsai.pro)

## What Was Done

### 1. Stripe Products Created ✓

Your Stripe account now has these products configured:

| Plan | Product ID | Monthly Price | Yearly Price |
|------|------------|---------------|--------------|
| **Starter** | `prod_UPvrpWJ34fgUYa` | `price_1TR5zy6X8IBUtLKfylNCziIs` ($29) | `price_1TR5zz6X8IBUtLKfGVbMISjL` ($290) |
| **Studio** | `prod_UPvr2ZbWl3NWyZ` | `price_1TR5zz6X8IBUtLKfQ3OhIGhs` ($99) | `price_1TR5zz6X8IBUtLKfdJLVitiw` ($990) |
| **Agency** | `prod_UPvraQ3UZqCDZv` | `price_1TR6006X8IBUtLKfaAmvzJBm` ($299) | `price_1TR6006X8IBUtLKfpEFA5yd2` ($2990) |
| **Enterprise** | `prod_UPvrsqXLhvmE3d` | Custom | Custom |

### 2. Environment Configuration ✓

Created `services/billing/.env` with:
- Your live Stripe secret key
- All product and price IDs
- Database and Redis configuration

### 3. Checkout API Endpoint ✓

Added checkout endpoint to billing service:
- `POST /api/v1/checkout/create` - Creates Stripe Checkout sessions
- `GET /api/v1/checkout/session/{id}` - Get session status

### 4. Frontend Integration ✓

Updated pricing page with:
- Real Stripe Price IDs
- Upgrade buttons with checkout flow
- Loading states during checkout

---

## 🔧 Next Steps to Complete Setup

### Step 1: Set Up Stripe Webhook

**For Development (Local Testing):**
```bash
# Install Stripe CLI
brew install stripe/stripe-cli/stripe  # Mac
# or download from https://github.com/stripe/stripe-cli

# Login
stripe login

# Forward webhooks to local billing service
stripe listen --forward-to localhost:8086/webhook
```

**For Production (fsai.pro):**
1. Go to https://dashboard.stripe.com/webhooks
2. Click "Add endpoint"
3. URL: `https://fsai.pro/api/v1/billing/webhook`
4. Events to select:
   - ✅ `checkout.session.completed`
   - ✅ `customer.subscription.created`
   - ✅ `customer.subscription.updated`
   - ✅ `customer.subscription.deleted`
   - ✅ `invoice.payment_succeeded`
   - ✅ `invoice.payment_failed`
5. Copy the **Signing Secret** (starts with `whsec_`)
6. Add to `.env`:
   ```bash
   STRIPE_WEBHOOK_SECRET=whsec_REDACTED_secret_here
   ```

### Step 2: Get Your Stripe Publishable Key

1. Go to https://dashboard.stripe.com/apikeys
2. Copy **Publishable key** (starts with `pk_live_`)
3. Add to `apps/web/.env.local`:
   ```bash
   NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_live_...
   ```

### Step 3: Test the Checkout Flow

1. **Start the billing service:**
   ```bash
   cd services/billing
   pip install -e .
   python -m uvicorn app.main:app --reload --port 8086
   ```

2. **Start the frontend:**
   ```bash
   cd apps/web
   npm install
   npm run dev
   ```

3. **Go to pricing page:**
   - http://localhost:3000/dashboard/pricing
   - Click "Upgrade" on any plan
   - Complete test payment with card: `4242 4242 4242 4242`

### Step 4: Deploy to fsai.pro

See `DEPLOYMENT_HOSTINGER.md` for full deployment guide.

Quick summary:
- **Option A**: VPS (recommended) - run `deploy-vps.sh`
- **Option B**: Hybrid - Frontend on Hostinger, backend on Cloud

---

## 📊 Your Stripe Dashboard

View your products and payments:
- **Products**: https://dashboard.stripe.com/test/products
- **Payments**: https://dashboard.stripe.com/test/payments
- **Customers**: https://dashboard.stripe.com/test/customers
- **Webhooks**: https://dashboard.stripe.com/test/webhooks

*(Switch to "Live Mode" in the toggle when ready for real payments)*

---

## 🔑 Important Keys Reference

### Stripe Keys
```bash
# Secret Key (server-side)
STRIPE_SECRET_KEY=sk_live_REDACTED

# Publishable Key (client-side) - GET FROM STRIPE DASHBOARD
NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_live_...

# Webhook Secret - SET UP IN STRIPE DASHBOARD
STRIPE_WEBHOOK_SECRET=whsec_...
```

### Price IDs (for checkout)
```bash
# Starter
STRIPE_PRICE_STARTER_MONTHLY=price_1TR5zy6X8IBUtLKfylNCziIs
STRIPE_PRICE_STARTER_YEARLY=price_1TR5zz6X8IBUtLKfGVbMISjL

# Studio
STRIPE_PRICE_STUDIO_MONTHLY=price_1TR5zz6X8IBUtLKfQ3OhIGhs
STRIPE_PRICE_STUDIO_YEARLY=price_1TR5zz6X8IBUtLKfdJLVitiw

# Agency
STRIPE_PRICE_AGENCY_MONTHLY=price_1TR6006X8IBUtLKfaAmvzJBm
STRIPE_PRICE_AGENCY_YEARLY=price_1TR6006X8IBUtLKfpEFA5yd2
```

---

## ✅ Checklist

- [x] Stripe account created
- [x] Products & prices created
- [x] Environment file configured
- [x] Checkout API endpoint added
- [x] Frontend pricing page updated
- [ ] Webhook configured (do this now)
- [ ] Publishable key added to frontend
- [ ] Test checkout flow
- [ ] Deploy to fsai.pro

---

## 🆘 Need Help?

**Check these files:**
- `STRIPE_SETUP.md` - Detailed webhook setup
- `DEPLOYMENT_HOSTINGER.md` - Deployment guide
- `scripts/create_stripe_products.py` - Re-run if needed

**Common Issues:**

1. **"Price ID not configured"**: Check STRIPE_PRICE_IDS in PricingView.tsx
2. **Webhook not working**: Run `stripe listen` for local dev
3. **Payment failed**: Use test card `4242 4242 4242 4242` in test mode

---

**Stripe is now connected to Papabase! 🎉**

Next: Set up your webhook and test the checkout flow.
