'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { leadsApi } from '@/lib/api'
import toast from 'react-hot-toast'
import { Users, Plus, Search, Filter, MoreVertical, Phone, Mail, Building2, Flame, TrendingUp, BarChart3 } from 'lucide-react'

interface Lead {
  id: string
  name: string
  email: string
  phone: string
  company: string
  status: string
  source: string
  notes: string
  score?: number
  tier?: string
}

export default function LeadsView() {
  const [leads, setLeads] = useState<Lead[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [showAddModal, setShowAddModal] = useState(false)
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedLead, setSelectedLead] = useState<Lead | null>(null)
  const [leadScore, setLeadScore] = useState<any>(null)

  useEffect(() => {
    loadLeads()
  }, [])

  const loadLeads = async () => {
    try {
      const data = await leadsApi.list()
      setLeads(data.leads || [])
    } catch (error) {
      toast.error('Failed to load leads')
    } finally {
      setIsLoading(false)
    }
  }

  const handleScoreLead = async (lead: Lead) => {
    try {
      const score = await leadsApi.score(lead.id)
      setLeadScore(score)
      toast.success(`Lead scored: ${score.score}/100 (${score.tier})`)
    } catch (error) {
      toast.error('Failed to score lead')
    }
  }

  const handleGetInsights = async (lead: Lead) => {
    try {
      const insights = await leadsApi.getInsights(lead.id)
      setSelectedLead(lead)
      toast.success('Insights generated')
    } catch (error) {
      toast.error('Failed to get insights')
    }
  }

  const filteredLeads = leads.filter(lead =>
    lead.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    lead.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
    lead.company.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const stats = {
    total: leads.length,
    hot: leads.filter(l => l.status === 'lead').length,
    converting: leads.filter(l => ['quote', 'scheduled'].includes(l.status)).length,
    closed: leads.filter(l => ['invoiced', 'done'].includes(l.status)).length,
  }

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className="space-y-6"
    >
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-800">Leads</h1>
          <p className="text-gray-600">Manage and qualify your sales pipeline</p>
        </div>
        <button
          onClick={() => setShowAddModal(true)}
          className="flex items-center gap-2 px-6 py-3 bg-papabase-pink text-white rounded-xl font-medium hover:bg-papabase-purple transition-colors"
        >
          <Plus size={20} />
          Add Lead
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {[
          { label: 'Total Leads', value: stats.total, icon: Users, color: 'bg-papabase-purple' },
          { label: 'Hot Leads', value: stats.hot, icon: Flame, color: 'bg-red-500' },
          { label: 'Converting', value: stats.converting, icon: TrendingUp, color: 'bg-papabase-teal' },
          { label: 'Closed', value: stats.closed, icon: BarChart3, color: 'bg-papabase-green' },
        ].map((stat) => {
          const Icon = stat.icon
          return (
            <div key={stat.label} className="bg-white p-4 rounded-xl shadow-sm">
              <div className="flex items-center gap-3">
                <div className={`${stat.color} p-2 rounded-lg text-white`}>
                  <Icon size={20} />
                </div>
                <div>
                  <p className="text-2xl font-bold text-gray-800">{stat.value}</p>
                  <p className="text-sm text-gray-500">{stat.label}</p>
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Search & Filter */}
      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={20} />
          <input
            type="text"
            placeholder="Search leads..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-teal"
          />
        </div>
        <button className="flex items-center gap-2 px-4 py-3 border border-gray-200 rounded-xl hover:bg-gray-50">
          <Filter size={20} />
          Filter
        </button>
      </div>

      {/* Leads List */}
      <div className="bg-white rounded-xl shadow-sm overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-4 text-left text-sm font-semibold text-gray-600">Lead</th>
              <th className="px-6 py-4 text-left text-sm font-semibold text-gray-600">Company</th>
              <th className="px-6 py-4 text-left text-sm font-semibold text-gray-600">Status</th>
              <th className="px-6 py-4 text-left text-sm font-semibold text-gray-600">Source</th>
              <th className="px-6 py-4 text-right text-sm font-semibold text-gray-600">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {filteredLeads.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-6 py-12 text-center text-gray-500">
                  <Users size={48} className="mx-auto mb-4 text-gray-300" />
                  <p>No leads yet. Add your first lead to get started!</p>
                </td>
              </tr>
            ) : (
              filteredLeads.map((lead) => (
                <tr key={lead.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4">
                    <div>
                      <p className="font-medium text-gray-800">{lead.name}</p>
                      <div className="flex items-center gap-3 text-sm text-gray-500 mt-1">
                        {lead.email && (
                          <span className="flex items-center gap-1">
                            <Mail size={14} />
                            {lead.email}
                          </span>
                        )}
                        {lead.phone && (
                          <span className="flex items-center gap-1">
                            <Phone size={14} />
                            {lead.phone}
                          </span>
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    {lead.company ? (
                      <span className="flex items-center gap-2 text-gray-700">
                        <Building2 size={16} />
                        {lead.company}
                      </span>
                    ) : (
                      <span className="text-gray-400">—</span>
                    )}
                  </td>
                  <td className="px-6 py-4">
                    <span className={`px-3 py-1 rounded-full text-sm font-medium ${
                      lead.status === 'lead' ? 'bg-red-100 text-red-700' :
                      lead.status === 'quote' ? 'bg-yellow-100 text-yellow-700' :
                      lead.status === 'scheduled' ? 'bg-blue-100 text-blue-700' :
                      lead.status === 'invoiced' ? 'bg-purple-100 text-purple-700' :
                      'bg-green-100 text-green-700'
                    }`}>
                      {lead.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-gray-600">{lead.source || '—'}</td>
                  <td className="px-6 py-4">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        onClick={() => handleScoreLead(lead)}
                        className="p-2 text-papabase-purple hover:bg-papabase-purple hover:text-white rounded-lg transition-colors"
                        title="Score Lead with AI"
                      >
                        <Flame size={18} />
                      </button>
                      <button
                        onClick={() => handleGetInsights(lead)}
                        className="p-2 text-papabase-teal hover:bg-papabase-teal hover:text-white rounded-lg transition-colors"
                        title="Get AI Insights"
                      >
                        <BarChart3 size={18} />
                      </button>
                      <button className="p-2 text-gray-400 hover:bg-gray-100 rounded-lg transition-colors">
                        <MoreVertical size={18} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Add Lead Modal */}
      {showAddModal && (
        <AddLeadModal
          onClose={() => setShowAddModal(false)}
          onAdd={async (data) => {
            try {
              await leadsApi.create(data)
              toast.success('Lead added!')
              loadLeads()
              setShowAddModal(false)
            } catch (error) {
              toast.error('Failed to add lead')
            }
          }}
        />
      )}

      {/* Lead Insights Modal */}
      {selectedLead && leadScore && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl p-6 max-w-md w-full">
            <h3 className="text-xl font-bold text-gray-800 mb-4">
              AI Lead Score: {selectedLead.name}
            </h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-gray-600">Score</span>
                <span className={`text-2xl font-bold ${
                  leadScore.score >= 80 ? 'text-red-500' :
                  leadScore.score >= 60 ? 'text-yellow-500' : 'text-gray-500'
                }`}>
                  {leadScore.score}/100
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-gray-600">Tier</span>
                <span className="px-3 py-1 rounded-full bg-papabase-purple text-white text-sm font-medium">
                  {leadScore.tier}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-gray-600">Action</span>
                <span className="text-papabase-teal font-medium">{leadScore.recommended_action}</span>
              </div>
              <div>
                <p className="text-gray-600 text-sm mb-2">Reasoning</p>
                <p className="text-gray-800 text-sm bg-gray-50 p-3 rounded-lg">{leadScore.reasoning}</p>
              </div>
            </div>
            <button
              onClick={() => { setSelectedLead(null); setLeadScore(null); }}
              className="mt-6 w-full py-3 bg-papabase-purple text-white rounded-xl font-medium"
            >
              Close
            </button>
          </div>
        </div>
      )}
    </motion.div>
  )
}

// Add Lead Modal Component
function AddLeadModal({ onClose, onAdd }: { onClose: () => void, onAdd: (data: any) => Promise<void> }) {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    phone: '',
    company: '',
    source: 'website',
    notes: '',
    tenant_id: 'default',
  })

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-2xl p-6 max-w-md w-full">
        <h3 className="text-xl font-bold text-gray-800 mb-4">Add New Lead</h3>
        <form onSubmit={(e) => { e.preventDefault(); onAdd(formData); }} className="space-y-4">
          <input
            type="text"
            placeholder="Name *"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink"
            required
          />
          <input
            type="email"
            placeholder="Email"
            value={formData.email}
            onChange={(e) => setFormData({ ...formData, email: e.target.value })}
            className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink"
          />
          <input
            type="tel"
            placeholder="Phone"
            value={formData.phone}
            onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
            className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink"
          />
          <input
            type="text"
            placeholder="Company"
            value={formData.company}
            onChange={(e) => setFormData({ ...formData, company: e.target.value })}
            className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink"
          />
          <select
            value={formData.source}
            onChange={(e) => setFormData({ ...formData, source: e.target.value })}
            className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink"
          >
            <option value="website">Website</option>
            <option value="referral">Referral</option>
            <option value="social">Social Media</option>
            <option value="cold">Cold Outreach</option>
            <option value="other">Other</option>
          </select>
          <textarea
            placeholder="Notes"
            value={formData.notes}
            onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
            className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink"
            rows={3}
          />
          <div className="flex gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 py-3 border border-gray-200 rounded-xl font-medium hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="flex-1 py-3 bg-papabase-pink text-white rounded-xl font-medium hover:bg-papabase-purple"
            >
              Add Lead
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
