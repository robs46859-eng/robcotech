'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { pricingApi } from '@/lib/api'
import { CreditCard, Check, Sparkles, TrendingUp, ExternalLink } from 'lucide-react'
import toast from 'react-hot-toast'

// Stripe Price IDs (Recreated with rotated key)
const STRIPE_PRICE_IDS: Record<string, { monthly: string; yearly: string }> = {
  starter: {
    monthly: 'price_1TR6gt6X8IBUtLKfsGUeNliQ',
    yearly: 'price_1TR6gt6X8IBUtLKf2qc2Yxci',
  },
  studio: {
    monthly: 'price_1TR6gu6X8IBUtLKffqgXoEWO',
    yearly: 'price_1TR6gu6X8IBUtLKffUhXngpr',
  },
  agency: {
    monthly: 'price_1TR6gu6X8IBUtLKfLx7JOVKL',
    yearly: 'price_1TR6gv6X8IBUtLKfOtPyJMh7',
  },
}

export default function PricingView() {
  const [plans, setPlans] = useState<any[]>([])
  const [usage, setUsage] = useState<any>(null)
  const [billingPeriod, setBillingPeriod] = useState<'monthly' | 'yearly'>('monthly')
  const [isCheckingOut, setIsCheckingOut] = useState<string | null>(null)

  const handleUpgrade = async (planId: string) => {
    if (!STRIPE_PRICE_IDS[planId]) {
      toast.error('Price ID not configured for this plan')
      return
    }

    setIsCheckingOut(planId)
    try {
      // In production, this would call your backend to create a checkout session
      // For now, we'll use Stripe's client-side checkout
      const priceId = STRIPE_PRICE_IDS[planId][billingPeriod]
      
      // Redirect to Stripe Checkout
      // Note: You'll need to set up a backend endpoint to create the session securely
      toast.success(`Redirecting to Stripe for ${planId} ${billingPeriod} plan...`)
      
      // Example: window.location.href = `/api/create-checkout-session?price_id=${priceId}`
      
      // For demo purposes, show what would happen
      setTimeout(() => {
        toast.error('Checkout endpoint not configured yet. See STRIPE_SETUP.md')
        setIsCheckingOut(null)
      }, 2000)
      
    } catch (error) {
      toast.error('Failed to start checkout')
      setIsCheckingOut(null)
    }
  }

  useEffect(() => {
    loadPricing()
  }, [])

  const loadPricing = async () => {
    try {
      const [plansData, usageData] = await Promise.all([
        pricingApi.plans(),
        pricingApi.usage(),
      ])
      setPlans(plansData.plans || [])
      setUsage(usageData)
    } catch (error) {
      // Use mock data
      setPlans([
        { id: 'starter', name: 'Starter', price_monthly: 29, price_yearly: 290, features: ['Single-page HTML', 'Basic CRM', '3 AI generations/mo'] },
        { id: 'studio', name: 'Studio', price_monthly: 99, price_yearly: 990, features: ['Multi-page React', 'Full CRM', '15 AI generations/mo', 'Custom domain'] },
        { id: 'agency', name: 'Agency', price_monthly: 299, price_yearly: 2990, features: ['Full web apps', 'Unlimited AI', 'White-label', 'API access'] },
        { id: 'enterprise', name: 'Enterprise', price_monthly: 0, price_yearly: 0, features: ['Everything in Agency', 'Dedicated infra', '24/7 support'] },
      ])
    }
  }

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Pricing & Billing</h1>
        <p className="text-gray-600">Manage your subscription and usage</p>
      </div>

      {/* Usage Stats */}
      <div className="bg-gradient-to-r from-papabase-purple to-papabase-pink p-6 rounded-2xl text-white">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-white/80 mb-1">Current Plan</p>
            <p className="text-3xl font-bold">Starter</p>
          </div>
          <div className="text-right">
            <p className="text-white/80 mb-1">AI Usage This Month</p>
            <p className="text-3xl font-bold">{usage ? (usage.tokens_used / 1000).toFixed(1) + 'K' : '0'} / 3K</p>
          </div>
        </div>
        <div className="mt-4 bg-white/20 rounded-full h-3">
          <div className="bg-white rounded-full h-3 w-[33%]" />
        </div>
      </div>

      {/* Billing Period Toggle */}
      <div className="flex items-center justify-center gap-4">
        <span className={`font-medium ${billingPeriod === 'monthly' ? 'text-gray-800' : 'text-gray-400'}`}>Monthly</span>
        <button
          onClick={() => setBillingPeriod(billingPeriod === 'monthly' ? 'yearly' : 'monthly')}
          className="w-16 h-8 bg-papabase-purple rounded-full relative transition-colors"
        >
          <div className={`w-6 h-6 bg-white rounded-full absolute top-1 transition-transform ${billingPeriod === 'yearly' ? 'translate-x-9' : 'translate-x-1'}`} />
        </button>
        <span className={`font-medium ${billingPeriod === 'yearly' ? 'text-gray-800' : 'text-gray-400'}`}>
          Yearly <span className="text-papabase-green text-sm">(Save 2 months)</span>
        </span>
      </div>

      {/* Pricing Cards */}
      <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
        {plans.map((plan) => (
          <div
            key={plan.id}
            className={`bg-white p-6 rounded-2xl shadow-sm border-2 ${
              plan.id === 'studio' ? 'border-papabase-teal' : 'border-transparent'
            }`}
          >
            {plan.id === 'studio' && (
              <div className="flex items-center gap-2 text-papabase-teal text-sm font-medium mb-2">
                <Sparkles size={16} />
                Most Popular
              </div>
            )}
            <h3 className="text-xl font-bold text-gray-800">{plan.name}</h3>
            <div className="mt-4">
              <span className="text-4xl font-bold text-gray-800">
                ${billingPeriod === 'monthly' ? plan.price_monthly : plan.price_yearly}
              </span>
              <span className="text-gray-500">/{billingPeriod === 'monthly' ? 'mo' : 'yr'}</span>
            </div>
            <ul className="mt-6 space-y-3">
              {plan.features?.map((feature: string, i: number) => (
                <li key={i} className="flex items-center gap-2 text-gray-600">
                  <Check className="text-papabase-green flex-shrink-0" size={18} />
                  <span className="text-sm">{feature}</span>
                </li>
              ))}
            </ul>
            <button
              onClick={() => handleUpgrade(plan.id)}
              disabled={isCheckingOut === plan.id}
              className={`w-full mt-6 py-3 rounded-xl font-semibold transition-colors flex items-center justify-center gap-2 ${
                plan.id === 'studio'
                  ? 'bg-papabase-teal text-white hover:bg-papabase-green'
                  : 'bg-gray-100 text-gray-800 hover:bg-gray-200'
              } disabled:opacity-50`}
            >
              {isCheckingOut === plan.id ? (
                <>Processing...</>
              ) : plan.price_monthly === 0 ? (
                <>Contact Sales <ExternalLink size={16} /></>
              ) : (
                <>Upgrade <CreditCard size={16} /></>
              )}
            </button>
          </div>
        ))}
      </div>

      {/* Feature Comparison */}
      <div className="bg-white p-6 rounded-2xl shadow-sm">
        <h2 className="text-xl font-bold text-gray-800 mb-6">Feature Comparison</h2>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b">
                <th className="text-left py-4 px-4 text-gray-600 font-medium">Feature</th>
                <th className="text-center py-4 px-4 text-gray-800 font-semibold">Starter</th>
                <th className="text-center py-4 px-4 text-papabase-teal font-semibold">Studio</th>
                <th className="text-center py-4 px-4 text-gray-800 font-semibold">Agency</th>
                <th className="text-center py-4 px-4 text-gray-800 font-semibold">Enterprise</th>
              </tr>
            </thead>
            <tbody>
              {[
                { feature: 'Seats', starter: '1', studio: '5', agency: '20', enterprise: 'Unlimited' },
                { feature: 'AI Generations', starter: '3/mo', studio: '15/mo', agency: 'Unlimited', enterprise: 'Unlimited' },
                { feature: 'Output Types', starter: 'Single-page', studio: 'Multi-page', agency: 'Dashboards', enterprise: 'Custom' },
                { feature: 'Custom Domain', starter: '—', studio: '✓', agency: '✓', enterprise: '✓' },
                { feature: 'White Label', starter: '—', studio: '—', agency: '✓', enterprise: '✓' },
                { feature: 'API Access', starter: '—', studio: '—', agency: '✓', enterprise: '✓' },
                { feature: 'Support', starter: 'Community', studio: 'Priority', agency: 'Dedicated', enterprise: '24/7' },
              ].map((row, i) => (
                <tr key={i} className="border-b last:border-0">
                  <td className="py-4 px-4 font-medium text-gray-800">{row.feature}</td>
                  <td className="text-center py-4 px-4 text-gray-600">{row.starter}</td>
                  <td className="text-center py-4 px-4 text-gray-800 bg-papabase-teal/5">{row.studio}</td>
                  <td className="text-center py-4 px-4 text-gray-600">{row.agency}</td>
                  <td className="text-center py-4 px-4 text-gray-600">{row.enterprise}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Invoices */}
      <div className="bg-white p-6 rounded-2xl shadow-sm">
        <h2 className="text-xl font-bold text-gray-800 mb-4">Billing History</h2>
        <div className="space-y-3">
          {[
            { date: 'Apr 1, 2026', amount: '$29.00', status: 'Paid', invoice: 'INV-001' },
            { date: 'Mar 1, 2026', amount: '$29.00', status: 'Paid', invoice: 'INV-000' },
          ].map((invoice, i) => (
            <div key={i} className="flex items-center justify-between p-4 bg-gray-50 rounded-xl">
              <div className="flex items-center gap-4">
                <CreditCard className="text-gray-400" size={24} />
                <div>
                  <p className="font-medium text-gray-800">{invoice.invoice}</p>
                  <p className="text-sm text-gray-500">{invoice.date}</p>
                </div>
              </div>
              <div className="text-right">
                <p className="font-semibold text-gray-800">{invoice.amount}</p>
                <p className="text-sm text-papabase-green">{invoice.status}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </motion.div>
  )
}
