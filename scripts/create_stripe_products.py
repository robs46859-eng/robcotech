#!/usr/bin/env python3
"""
Stripe Products Setup Script

Creates Stripe products and prices for Papabase subscription tiers.

Usage:
    1. Set your Stripe secret key: export STRIPE_SECRET_KEY=sk_test_...
    2. Run: python create_stripe_products.py
    3. Copy the price IDs to your .env file
"""

import os
import stripe

# Get Stripe key from environment
stripe.api_key = os.environ.get("STRIPE_SECRET_KEY")

if not stripe.api_key:
    print("❌ Error: STRIPE_SECRET_KEY not set")
    print("Run: export STRIPE_SECRET_KEY=sk_test_...")
    exit(1)

# Papabase pricing tiers
PRODUCTS = {
    "starter": {
        "name": "Papabase Starter",
        "description": "For solo operators and side hustles",
        "price_monthly": 2900,  # $29.00 in cents
        "price_yearly": 29000,  # $290.00 (2 months free)
        "features": [
            "Single-page HTML websites",
            "Basic CRM (leads & tasks)",
            "3 AI generations per month",
            "1 GB storage",
            "Community support",
        ],
    },
    "studio": {
        "name": "Papabase Studio",
        "description": "For small studios (2-5 people)",
        "price_monthly": 9900,  # $99.00
        "price_yearly": 99000,  # $990.00
        "features": [
            "Multi-page React websites",
            "Full CRM with workflows",
            "15 AI generations per month",
            "10 GB storage",
            "Custom domain",
            "Priority support",
        ],
    },
    "agency": {
        "name": "Papabase Agency",
        "description": "For growing agencies (6-20 people)",
        "price_monthly": 29900,  # $299.00
        "price_yearly": 299000,  # $2,990.00
        "features": [
            "Full web applications",
            "Unlimited AI generations",
            "100 GB storage",
            "White-label option",
            "API access",
            "Dedicated support",
        ],
    },
    "enterprise": {
        "name": "Papabase Enterprise",
        "description": "For large teams with custom needs",
        "price_monthly": None,  # Custom pricing
        "price_yearly": None,
        "features": [
            "Everything in Agency",
            "Unlimited seats",
            "Dedicated infrastructure",
            "SSO/SAML",
            "On-premise option",
            "24/7 phone support",
        ],
    },
}

def create_product(key: str, data: dict) -> dict:
    """Create a Stripe product with monthly and yearly prices"""
    print(f"\n{'='*50}")
    print(f"Creating: {data['name']}")
    print(f"{'='*50}")
    
    # Create product
    product = stripe.Product.create(
        name=data["name"],
        description=data["description"],
        metadata={
            "tier": key,
            "features": ", ".join(data["features"]),
        },
    )
    print(f"✓ Product ID: {product.id}")
    
    result = {"product_id": product.id}
    
    # Create prices if not enterprise
    if data.get("price_monthly"):
        # Monthly price
        monthly_price = stripe.Price.create(
            product=product.id,
            unit_amount=data["price_monthly"],
            currency="usd",
            recurring={"interval": "month"},
            metadata={"billing": "monthly"},
        )
        print(f"✓ Monthly Price ID: {monthly_price.id}")
        print(f"  - ${data['price_monthly']/100:.2f}/month")
        result["monthly_price_id"] = monthly_price.id
        
        # Yearly price
        yearly_price = stripe.Price.create(
            product=product.id,
            unit_amount=data["price_yearly"],
            currency="usd",
            recurring={"interval": "year"},
            metadata={"billing": "yearly"},
        )
        print(f"✓ Yearly Price ID: {yearly_price.id}")
        print(f"  - ${data['price_yearly']/100:.2f}/year")
        result["yearly_price_id"] = yearly_price.id
    
    return result

def main():
    print("\n" + "="*60)
    print("  Papabase Stripe Products Setup")
    print("="*60)
    print(f"\nUsing Stripe key: {stripe.api_key[:10]}...")
    
    # Test connection
    try:
        stripe.Balance.retrieve()
        print("✓ Connected to Stripe API\n")
    except Exception as e:
        print(f"❌ Stripe connection failed: {e}")
        exit(1)
    
    # Create products
    results = {}
    for key, data in PRODUCTS.items():
        try:
            results[key] = create_product(key, data)
        except Exception as e:
            print(f"❌ Failed to create {key}: {e}")
    
    # Print summary
    print("\n" + "="*60)
    print("  SETUP COMPLETE - Copy these to your .env file")
    print("="*60)
    
    print("\n```bash")
    print("# Stripe Configuration for Papabase")
    for key, data in results.items():
        print(f"STRIPE_PRODUCT_{key.upper()}={data['product_id']}")
        if "monthly_price_id" in data:
            print(f"STRIPE_PRICE_{key.upper()}_MONTHLY={data['monthly_price_id']}")
            print(f"STRIPE_PRICE_{key.upper()}_YEARLY={data['yearly_price_id']}")
    print("```")
    
    print("\n" + "="*60)
    print("Next steps:")
    print("1. Copy the above IDs to services/billing/.env")
    print("2. Set up webhook: stripe listen --forward-to localhost:8086/webhook")
    print("3. Test subscription flow in the Papabase UI")
    print("="*60 + "\n")

if __name__ == "__main__":
    main()
