'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { websiteApi } from '@/lib/api'
import toast from 'react-hot-toast'
import { Sparkles, Globe, Layout, Monitor, Palette, Type, Rocket, Check } from 'lucide-react'

export default function WebsiteGenerator() {
  const [step, setStep] = useState(1)
  const [isGenerating, setIsGenerating] = useState(false)
  const [formData, setFormData] = useState({
    prompt: '',
    business_type: '',
    tier: 'starter',
    output_type: 'single_page',
    color_scheme: '',
    features: [] as string[],
  })
  const [result, setResult] = useState<any>(null)

  const templates = [
    { id: 'landing-page', name: 'Landing Page', tier: 'starter', icon: Layout },
    { id: 'business-site', name: 'Business Site', tier: 'studio', icon: Globe },
    { id: 'portfolio', name: 'Portfolio', tier: 'studio', icon: Palette },
    { id: 'saas-dashboard', name: 'SaaS Dashboard', tier: 'agency', icon: Monitor },
    { id: 'ecommerce', name: 'E-commerce', tier: 'agency', icon: Layout },
  ]

  const handleGenerate = async () => {
    setIsGenerating(true)
    try {
      const data = await websiteApi.generate(formData)
      setResult(data)
      toast.success('Website generated!')
      setStep(3)
    } catch (error) {
      toast.error('Failed to generate website')
    } finally {
      setIsGenerating(false)
    }
  }

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className="space-y-6"
    >
      {/* Header */}
      <div>
        <div className="flex items-center gap-3 mb-2">
          <div className="p-3 bg-gradient-to-br from-papabase-teal to-papabase-green rounded-xl text-white">
            <Sparkles size={28} />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-gray-800">Dad AI Website Builder</h1>
            <p className="text-gray-600">Develop Another Day - Generate professional websites in seconds</p>
          </div>
        </div>
      </div>

      {/* Progress Steps */}
      <div className="flex items-center justify-center gap-4">
        {[
          { num: 1, label: 'Describe' },
          { num: 2, label: 'Customize' },
          { num: 3, label: 'Generate' },
        ].map((s) => (
          <div key={s.num} className="flex items-center">
            <div className={`w-10 h-10 rounded-full flex items-center justify-center font-bold ${
              step >= s.num ? 'bg-papabase-teal text-white' : 'bg-gray-200 text-gray-500'
            }`}>
              {s.num}
            </div>
            <span className={`ml-2 ${step >= s.num ? 'text-gray-800' : 'text-gray-400'}`}>{s.label}</span>
            {s.num < 3 && <div className={`w-12 h-1 mx-4 ${step > s.num ? 'bg-papabase-teal' : 'bg-gray-200'}`} />}
          </div>
        ))}
      </div>

      {/* Step 1: Describe */}
      {step === 1 && (
        <div className="bg-white p-6 rounded-2xl shadow-sm space-y-6">
          <h2 className="text-xl font-bold text-gray-800">Describe Your Website</h2>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              What kind of website do you need? *
            </label>
            <textarea
              value={formData.prompt}
              onChange={(e) => setFormData({ ...formData, prompt: e.target.value })}
              placeholder="e.g., 'I need a professional website for my plumbing business. We offer emergency services, installations, and repairs. Want customers to be able to book appointments online.'"
              className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-teal min-h-[150px]"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">Business Type</label>
            <input
              type="text"
              value={formData.business_type}
              onChange={(e) => setFormData({ ...formData, business_type: e.target.value })}
              placeholder="e.g., Plumbing, Web Design, Restaurant..."
              className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-teal"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-3">Select Template</label>
            <div className="grid grid-cols-2 lg:grid-cols-5 gap-3">
              {templates.map((template) => {
                const Icon = template.icon
                return (
                  <button
                    key={template.id}
                    onClick={() => setFormData({ ...formData, output_type: template.tier === 'agency' ? 'dashboard' : template.tier === 'studio' ? 'multi_page' : 'single_page' })}
                    className={`p-4 rounded-xl border-2 text-center transition-all ${
                      formData.output_type === (template.tier === 'agency' ? 'dashboard' : template.tier === 'studio' ? 'multi_page' : 'single_page')
                        ? 'border-papabase-teal bg-papabase-teal/10'
                        : 'border-gray-200 hover:border-papabase-teal'
                    }`}
                  >
                    <Icon className="mx-auto mb-2 text-papabase-purple" size={24} />
                    <p className="font-medium text-sm">{template.name}</p>
                    <p className="text-xs text-gray-500 capitalize">{template.tier}</p>
                  </button>
                )
              })}
            </div>
          </div>

          <button
            onClick={() => setStep(2)}
            disabled={!formData.prompt.trim()}
            className="w-full py-4 bg-gradient-to-r from-papabase-teal to-papabase-green text-white rounded-xl font-semibold hover:shadow-lg disabled:opacity-50"
          >
            Continue
          </button>
        </div>
      )}

      {/* Step 2: Customize */}
      {step === 2 && (
        <div className="bg-white p-6 rounded-2xl shadow-sm space-y-6">
          <h2 className="text-xl font-bold text-gray-800">Customize Your Site</h2>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-3">Color Scheme</label>
            <div className="flex gap-3">
              {[
                { name: 'Ocean', colors: ['#3DC2B9', '#316844'] },
                { name: 'Sunset', colors: ['#EC89A3', '#643277'] },
                { name: 'Forest', colors: ['#316844', '#3DC2B9'] },
                { name: 'Royal', colors: ['#643277', '#EC89A3'] },
              ].map((scheme) => (
                <button
                  key={scheme.name}
                  onClick={() => setFormData({ ...formData, color_scheme: scheme.name })}
                  className={`p-4 rounded-xl border-2 transition-all ${
                    formData.color_scheme === scheme.name
                      ? 'border-papabase-purple'
                      : 'border-gray-200'
                  }`}
                >
                  <div className="flex gap-1 mb-2">
                    {scheme.colors.map((c) => (
                      <div key={c} className="w-6 h-6 rounded-full" style={{ backgroundColor: c }} />
                    ))}
                  </div>
                  <p className="text-sm font-medium">{scheme.name}</p>
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-3">Features</label>
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
              {['Contact Form', 'Gallery', 'Testimonials', 'Services', 'About Us', 'Blog', 'Booking', 'FAQ'].map((feature) => (
                <button
                  key={feature}
                  onClick={() => setFormData({ 
                    ...formData, 
                    features: formData.features.includes(feature)
                      ? formData.features.filter(f => f !== feature)
                      : [...formData.features, feature]
                  })}
                  className={`p-3 rounded-xl border-2 flex items-center gap-2 transition-all ${
                    formData.features.includes(feature)
                      ? 'border-papabase-teal bg-papabase-teal/10 text-papabase-teal'
                      : 'border-gray-200 hover:border-papabase-teal'
                  }`}
                >
                  {formData.features.includes(feature) && <Check size={16} />}
                  <span className="text-sm font-medium">{feature}</span>
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-3">Pricing Tier</label>
            <div className="grid grid-cols-3 gap-4">
              {[
                { id: 'starter', name: 'Starter', price: '$29/mo', output: 'Single Page' },
                { id: 'studio', name: 'Studio', price: '$99/mo', output: 'Multi-Page' },
                { id: 'agency', name: 'Agency', price: '$299/mo', output: 'Dashboard' },
              ].map((tier) => (
                <button
                  key={tier.id}
                  onClick={() => setFormData({ ...formData, tier: tier.id as any })}
                  className={`p-4 rounded-xl border-2 text-center transition-all ${
                    formData.tier === tier.id
                      ? 'border-papabase-purple bg-papabase-purple/10'
                      : 'border-gray-200 hover:border-papabase-purple'
                  }`}
                >
                  <p className="font-bold text-gray-800">{tier.name}</p>
                  <p className="text-papabase-purple font-semibold">{tier.price}</p>
                  <p className="text-xs text-gray-500">{tier.output}</p>
                </button>
              ))}
            </div>
          </div>

          <div className="flex gap-3">
            <button
              onClick={() => setStep(1)}
              className="flex-1 py-4 border border-gray-200 rounded-xl font-semibold hover:bg-gray-50"
            >
              Back
            </button>
            <button
              onClick={handleGenerate}
              disabled={isGenerating}
              className="flex-1 py-4 bg-gradient-to-r from-papabase-teal to-papabase-green text-white rounded-xl font-semibold hover:shadow-lg disabled:opacity-50 flex items-center justify-center gap-2"
            >
              {isGenerating ? (
                <>Generating...</>
              ) : (
                <>
                  <Rocket size={20} />
                  Generate Website
                </>
              )}
            </button>
          </div>
        </div>
      )}

      {/* Step 3: Result */}
      {step === 3 && result && (
        <div className="bg-white p-6 rounded-2xl shadow-sm">
          <h2 className="text-xl font-bold text-gray-800 mb-4">Website Generated!</h2>
          
          <div className="aspect-video bg-gray-100 rounded-xl mb-6 flex items-center justify-center">
            <Globe className="text-gray-400" size={64} />
          </div>

          <div className="grid md:grid-cols-2 gap-4 mb-6">
            <div className="p-4 bg-gray-50 rounded-xl">
              <p className="text-sm text-gray-500 mb-1">Project ID</p>
              <p className="font-mono text-sm">{result.project_id}</p>
            </div>
            <div className="p-4 bg-gray-50 rounded-xl">
              <p className="text-sm text-gray-500 mb-1">Status</p>
              <p className="text-papabase-green font-medium">{result.status}</p>
            </div>
          </div>

          <div className="flex gap-3">
            <button
              onClick={() => { setStep(1); setResult(null); }}
              className="flex-1 py-4 border border-gray-200 rounded-xl font-semibold hover:bg-gray-50"
            >
              Generate Another
            </button>
            <button className="flex-1 py-4 bg-papabase-purple text-white rounded-xl font-semibold hover:bg-papabase-purple/90">
              View Full Site
            </button>
          </div>
        </div>
      )}
    </motion.div>
  )
}
