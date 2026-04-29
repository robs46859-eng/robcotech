'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { proposalApi, leadsApi } from '@/lib/api'
import toast from 'react-hot-toast'
import { FileText, Sparkles, DollarSign, ListChecks, Copy, Download } from 'lucide-react'

export default function ProposalsView() {
  const [activeTab, setActiveTab] = useState('generate')
  const [isGenerating, setIsGenerating] = useState(false)
  const [result, setResult] = useState<any>(null)
  const [formData, setFormData] = useState({
    lead_id: '',
    requirements: '',
    project_type: '',
    base_scope: '',
    services: [] as string[],
    client_info: { name: '', company: '' },
  })

  const handleGenerateProposal = async () => {
    setIsGenerating(true)
    try {
      const data = await proposalApi.generate(formData)
      setResult(data.proposal)
      toast.success('Proposal generated!')
    } catch (error) {
      toast.error('Failed to generate proposal')
    } finally {
      setIsGenerating(false)
    }
  }

  const handleGenerateQuote = async () => {
    setIsGenerating(true)
    try {
      const data = await proposalApi.quote(formData)
      setResult(data.quote)
      toast.success('Quote generated!')
    } catch (error) {
      toast.error('Failed to generate quote')
    } finally {
      setIsGenerating(false)
    }
  }

  const handlePricingTiers = async () => {
    setIsGenerating(true)
    try {
      const data = await proposalApi.pricingTiers(formData)
      setResult(data.pricing_tiers)
      toast.success('Pricing tiers generated!')
    } catch (error) {
      toast.error('Failed to generate pricing')
    } finally {
      setIsGenerating(false)
    }
  }

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Proposal Builder</h1>
        <p className="text-gray-600">Create professional proposals and quotes in seconds</p>
      </div>

      {/* Tabs */}
      <div className="flex gap-2">
        {[
          { id: 'generate', label: 'Full Proposal', icon: FileText },
          { id: 'quote', label: 'Quick Quote', icon: DollarSign },
          { id: 'tiers', label: 'Pricing Tiers', icon: ListChecks },
        ].map((tab) => {
          const Icon = tab.icon
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-all ${
                activeTab === tab.id
                  ? 'bg-papabase-purple text-white'
                  : 'bg-white text-gray-600 hover:bg-gray-50'
              }`}
            >
              <Icon size={18} />
              {tab.label}
            </button>
          )
        })}
      </div>

      {/* Full Proposal */}
      {activeTab === 'generate' && (
        <div className="bg-white p-6 rounded-2xl shadow-sm space-y-4">
          <h2 className="text-xl font-bold text-gray-800">Generate Full Proposal</h2>
          <select value={formData.lead_id} onChange={(e) => setFormData({...formData, lead_id: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-purple">
            <option value="">Select a Lead</option>
            <option value="demo-1">Demo Lead - John Doe (Acme Corp)</option>
          </select>
          <textarea placeholder="Project Requirements" value={formData.requirements} onChange={(e) => setFormData({...formData, requirements: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-purple" rows={6} />
          <button onClick={handleGenerateProposal} disabled={isGenerating} className="w-full py-4 bg-gradient-to-r from-papabase-purple to-papabase-pink text-white rounded-xl font-semibold hover:shadow-lg disabled:opacity-50 flex items-center justify-center gap-2">
            <Sparkles size={20} />
            {isGenerating ? 'Generating...' : 'Generate Proposal'}
          </button>
        </div>
      )}

      {/* Quick Quote */}
      {activeTab === 'quote' && (
        <div className="bg-white p-6 rounded-2xl shadow-sm space-y-4">
          <h2 className="text-xl font-bold text-gray-800">Generate Quick Quote</h2>
          <input type="text" placeholder="Client Name" value={formData.client_info.name} onChange={(e) => setFormData({...formData, client_info: {...formData.client_info, name: e.target.value}})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-teal" />
          <input type="text" placeholder="Company" value={formData.client_info.company} onChange={(e) => setFormData({...formData, client_info: {...formData.client_info, company: e.target.value}})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-teal" />
          <div className="space-y-2">
            <p className="text-sm font-medium text-gray-700">Services</p>
            {['Web Design', 'Development', 'SEO', 'Content Writing', 'Consulting'].map((service) => (
              <label key={service} className="flex items-center gap-3">
                <input type="checkbox" checked={formData.services.includes(service)} onChange={(e) => setFormData({...formData, services: e.target.checked ? [...formData.services, service] : formData.services.filter(s => s !== service)})} className="w-4 h-4 text-papabase-teal" />
                <span className="text-gray-700">{service}</span>
              </label>
            ))}
          </div>
          <button onClick={handleGenerateQuote} disabled={isGenerating} className="w-full py-4 bg-gradient-to-r from-papabase-teal to-papabase-green text-white rounded-xl font-semibold hover:shadow-lg disabled:opacity-50 flex items-center justify-center gap-2">
            <DollarSign size={20} />
            {isGenerating ? 'Generating...' : 'Generate Quote'}
          </button>
        </div>
      )}

      {/* Pricing Tiers */}
      {activeTab === 'tiers' && (
        <div className="bg-white p-6 rounded-2xl shadow-sm space-y-4">
          <h2 className="text-xl font-bold text-gray-800">Generate Pricing Tiers</h2>
          <textarea placeholder="Base Scope of Work" value={formData.base_scope} onChange={(e) => setFormData({...formData, base_scope: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-green" rows={4} />
          <button onClick={handlePricingTiers} disabled={isGenerating} className="w-full py-4 bg-gradient-to-r from-papabase-green to-papabase-teal text-white rounded-xl font-semibold hover:shadow-lg disabled:opacity-50 flex items-center justify-center gap-2">
            <ListChecks size={20} />
            {isGenerating ? 'Generating...' : 'Generate Pricing Tiers'}
          </button>
        </div>
      )}

      {/* Result */}
      {result && (
        <div className="bg-white p-6 rounded-2xl shadow-sm">
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-bold text-gray-800">{result.title || 'Generated Proposal'}</h3>
            <div className="flex gap-2">
              <button onClick={() => { navigator.clipboard.writeText(JSON.stringify(result, null, 2)); toast.success('Copied!') }} className="flex items-center gap-2 text-sm text-papabase-purple hover:underline"><Copy size={16} /> Copy</button>
              <button className="flex items-center gap-2 text-sm text-papabase-purple hover:underline"><Download size={16} /> PDF</button>
            </div>
          </div>
          
          {result.executive_summary && (
            <div className="mb-4">
              <h4 className="font-semibold text-gray-700 mb-2">Executive Summary</h4>
              <p className="text-gray-600">{result.executive_summary}</p>
            </div>
          )}

          {result.scope_of_work && (
            <div className="mb-4">
              <h4 className="font-semibold text-gray-700 mb-2">Scope of Work</h4>
              <div className="space-y-2">
                {result.scope_of_work.map((item: any, i: number) => (
                  <div key={i} className="p-3 bg-gray-50 rounded-lg">
                    <p className="font-medium">{item.title}</p>
                    <p className="text-sm text-gray-500">${item.price?.toLocaleString()}</p>
                  </div>
                ))}
              </div>
            </div>
          )}

          {result.pricing && (
            <div className="p-4 bg-papabase-purple/10 rounded-xl">
              <div className="flex justify-between items-center">
                <span className="font-semibold">Total:</span>
                <span className="text-2xl font-bold text-papabase-purple">${result.pricing.total?.toLocaleString()}</span>
              </div>
            </div>
          )}
        </div>
      )}
    </motion.div>
  )
}
