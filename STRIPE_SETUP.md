# Stripe Webhook Setup for Papabase

## Your Stripe Configuration

**Products Created:**
| Plan | Product ID | Monthly Price | Yearly Price |
|------|------------|---------------|--------------|
| Starter | `prod_UPvrpWJ34fgUYa` | `price_1TR5zy6X8IBUtLKfylNCziIs` | `price_1TR5zz6X8IBUtLKfGVbMISjL` |
| Studio | `prod_UPvr2ZbWl3NWyZ` | `price_1TR5zz6X8IBUtLKfQ3OhIGhs` | `price_1TR5zz6X8IBUtLKfdJLVitiw` |
| Agency | `prod_UPvraQ3UZqCDZv` | `price_1TR6006X8IBUtLKfaAmvzJBm` | `price_1TR6006X8IBUtLKfpEFA5yd2` |
| Enterprise | `prod_UPvrsqXLhvmE3d` | Custom | Custom |

---

## Step 1: Configure Stripe Webhook

### Option A: Using Stripe CLI (Development)

```bash
# Install Stripe CLI
# Mac: brew install stripe/stripe-cli/stripe
# Linux: curl -s https://packages.stripe.dev/api/security/keypair/stripe-cli-gpg/public | gpg --dearmor | sudo tee /usr/share/keyrings/stripe.gpg
#        echo "deb [signed-by=/usr/share/keyrings/stripe.gpg] https://packages.stripe.dev/stripe-cli-debian-local stable main" | sudo tee -a /etc/apt/sources.list.d/stripe.list
#        sudo apt update && sudo apt install stripe

# Login to Stripe
stripe login

# Forward webhooks to local billing service
stripe listen --forward-to localhost:8086/webhook
```

### Option B: Production Webhook (fsai.pro)

1. **Go to Stripe Dashboard**: https://dashboard.stripe.com/test/webhooks

2. **Add Endpoint**:
   - URL: `https://fsai.pro/api/v1/billing/webhook`
   - Events to send:
     - ✅ `checkout.session.completed`
     - ✅ `customer.subscription.created`
     - ✅ `customer.subscription.updated`
     - ✅ `customer.subscription.deleted`
     - ✅ `invoice.payment_succeeded`
     - ✅ `invoice.payment_failed`
     - ✅ `payment_intent.succeeded`

3. **Copy Webhook Secret**:
   - After creating, click "Reveal" next to Signing secret
   - Copy it (starts with `whsec_`)
   - Add to `.env`:
   ```bash
   STRIPE_WEBHOOK_SECRET=whsec_REDACTED_webhook_secret_here
   ```

---

## Step 2: Update Frontend with Price IDs

Update the pricing page to use real Stripe Checkout:

```bash
# Add to apps/web/.env.local
NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_live_...  # Get from Stripe Dashboard
NEXT_PUBLIC_STRIPE_PRICE_STARTER_MONTHLY=price_1TR5zy6X8IBUtLKfylNCziIs
NEXT_PUBLIC_STRIPE_PRICE_STUDIO_MONTHLY=price_1TR5zz6X8IBUtLKfQ3OhIGhs
NEXT_PUBLIC_STRIPE_PRICE_AGENCY_MONTHLY=price_1TR6006X8IBUtLKfaAmvzJBm
```

---

## Step 3: Test Payment Flow

1. **Start billing service**:
   ```bash
   cd services/billing
   pip install -e .
   python -m uvicorn app.main:app --reload --port 8086
   ```

2. **Start webhook forwarding**:
   ```bash
   stripe listen --forward-to localhost:8086/webhook
   ```

3. **Test in browser**:
   - Go to http://localhost:3000/dashboard/pricing
   - Click "Upgrade" on any plan
   - Use Stripe test card: `4242 4242 4242 4242`

---

## Step 4: Verify Webhook Events

Check that webhooks are being received:

```bash
# View billing service logs
docker compose logs -f billing

# Or check Stripe Dashboard → Developers → Webhooks → Events
```

---

## Quick Reference

### Your Stripe Keys
- **Publishable Key**: Get from https://dashboard.stripe.com/apikeys (starts with `pk_live_`)
- **Secret Key**: `sk_live_REDACTED`

### Test Cards (Test Mode Only)
| Card Number | Description |
|-------------|-------------|
| 4242 4242 4242 4242 | Success |
| 4000 0000 0000 9995 | Declined |
| 4000 0025 0000 3155 | Requires authentication |

---

## Next Steps

1. ✅ Stripe products created
2. ⏳ Set up webhook (choose Option A or B above)
3. ⏳ Update frontend with publishable key
4. ⏳ Test checkout flow
5. ⏳ Deploy to fsai.pro
