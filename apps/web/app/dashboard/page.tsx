'use client'

import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { useAuthStore } from '@/lib/store'
import { leadsApi, tasksApi, websiteApi, pricingApi } from '@/lib/api'
import toast from 'react-hot-toast'
import {
  LayoutDashboard, Users, CheckSquare, Sparkles, FileText, CreditCard,
  Settings, LogOut, Menu, X, Bell, Search, Plus, TrendingUp, DollarSign,
  Clock, Target, Zap, BarChart3, Globe, Palette, Type, MessageSquare
} from 'lucide-react'

// Dashboard sub-components
import LeadsView from '@/components/dashboard/LeadsView'
import TasksView from '@/components/dashboard/TasksView'
import WebsiteGenerator from '@/components/dashboard/WebsiteGenerator'
import ContentStudio from '@/components/dashboard/ContentStudio'
import ProposalsView from '@/components/dashboard/ProposalsView'
import PricingView from '@/components/dashboard/PricingView'

interface DashboardStats {
  totalLeads: number
  hotLeads: number
  totalTasks: number
  pendingTasks: number
  websitesGenerated: number
  monthlyUsage: number
}

export default function DashboardPage() {
  const { user, logout } = useAuthStore()
  const [activeTab, setActiveTab] = useState('overview')
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [stats, setStats] = useState<DashboardStats>({
    totalLeads: 0,
    hotLeads: 0,
    totalTasks: 0,
    pendingTasks: 0,
    websitesGenerated: 0,
    monthlyUsage: 0,
  })
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    loadDashboardData()
  }, [])

  const loadDashboardData = async () => {
    try {
      const [leads, tasks, websites, pricing] = await Promise.all([
        leadsApi.list().catch(() => ({ leads: [] })),
        tasksApi.list().catch(() => ({ tasks: [] })),
        websiteApi.projects().catch(() => ({ projects: [] })),
        pricingApi.usage().catch(() => ({ tokens_used: 0 })),
      ])

      const leadsList = leads.leads || []
      const tasksList = tasks.tasks || []

      setStats({
        totalLeads: leadsList.length,
        hotLeads: leadsList.filter((l: any) => l.status === 'lead').length,
        totalTasks: tasksList.length,
        pendingTasks: tasksList.filter((t: any) => t.status !== 'done').length,
        websitesGenerated: websites.projects?.length || 0,
        monthlyUsage: pricing.tokens_used || 0,
      })
    } catch (error) {
      console.error('Failed to load dashboard data:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const navItems = [
    { id: 'overview', label: 'Overview', icon: LayoutDashboard, color: 'papabase-purple' },
    { id: 'leads', label: 'Leads', icon: Users, color: 'papabase-pink' },
    { id: 'tasks', label: 'Tasks', icon: CheckSquare, color: 'papabase-green' },
    { id: 'websites', label: 'Dad AI Websites', icon: Sparkles, color: 'papabase-teal' },
    { id: 'content', label: 'Content Studio', icon: Type, color: 'papabase-pink' },
    { id: 'proposals', label: 'Proposals', icon: FileText, color: 'papabase-purple' },
    { id: 'pricing', label: 'Pricing & Billing', icon: CreditCard, color: 'papabase-green' },
  ]

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Mobile Sidebar Overlay */}
      {sidebarOpen && (
        <div 
          className="fixed inset-0 bg-black/50 z-40 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside className={`fixed top-0 left-0 h-full w-64 bg-white shadow-xl z-50 transform transition-transform lg:translate-x-0 ${
        sidebarOpen ? 'translate-x-0' : '-translate-x-full'
      }`}>
        <div className="p-6">
          <div className="flex items-center justify-between mb-8">
            <div className="flex items-center gap-2">
              <Sparkles className="text-papabase-purple" size={28} />
              <span className="text-xl font-bold text-papabase-purple">Papabase</span>
            </div>
            <button 
              onClick={() => setSidebarOpen(false)}
              className="lg:hidden text-gray-400 hover:text-gray-600"
            >
              <X size={24} />
            </button>
          </div>

          <nav className="space-y-2">
            {navItems.map((item) => {
              const Icon = item.icon
              return (
                <button
                  key={item.id}
                  onClick={() => {
                    setActiveTab(item.id)
                    setSidebarOpen(false)
                  }}
                  className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${
                    activeTab === item.id
                      ? `bg-${item.color} text-white shadow-lg`
                      : 'text-gray-600 hover:bg-gray-100'
                  }`}
                >
                  <Icon size={20} />
                  <span className="font-medium">{item.label}</span>
                </button>
              )
            })}
          </nav>
        </div>

        <div className="absolute bottom-0 left-0 right-0 p-6 border-t">
          <div className="flex items-center gap-3 mb-4">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-papabase-purple to-papabase-pink flex items-center justify-center text-white font-bold">
              {user?.name?.[0]?.toUpperCase() || 'U'}
            </div>
            <div className="flex-1">
              <p className="font-medium text-gray-800">{user?.name || 'User'}</p>
              <p className="text-xs text-gray-500 capitalize">{user?.plan || 'starter'} plan</p>
            </div>
          </div>
          <button
            onClick={() => logout()}
            className="w-full flex items-center justify-center gap-2 py-2 text-gray-600 hover:text-papabase-purple transition-colors"
          >
            <LogOut size={18} />
            Sign Out
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="lg:ml-64">
        {/* Header */}
        <header className="bg-white shadow-sm sticky top-0 z-30">
          <div className="flex items-center justify-between px-6 py-4">
            <div className="flex items-center gap-4">
              <button
                onClick={() => setSidebarOpen(true)}
                className="lg:hidden text-gray-600 hover:text-gray-800"
              >
                <Menu size={24} />
              </button>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={20} />
                <input
                  type="text"
                  placeholder="Search..."
                  className="pl-10 pr-4 py-2 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-teal w-64"
                />
              </div>
            </div>
            <div className="flex items-center gap-4">
              <button className="relative p-2 text-gray-600 hover:text-papabase-purple">
                <Bell size={20} />
                <span className="absolute top-1 right-1 w-2 h-2 bg-papabase-pink rounded-full" />
              </button>
              <button className="p-2 text-gray-600 hover:text-papabase-purple">
                <Settings size={20} />
              </button>
            </div>
          </div>
        </header>

        {/* Page Content */}
        <div className="p-6">
          <AnimatePresence mode="wait">
            {activeTab === 'overview' && (
              <OverviewView 
                key="overview" 
                stats={stats} 
                onNavigate={setActiveTab}
              />
            )}
            {activeTab === 'leads' && <LeadsView key="leads" />}
            {activeTab === 'tasks' && <TasksView key="tasks" />}
            {activeTab === 'websites' && <WebsiteGenerator key="websites" />}
            {activeTab === 'content' && <ContentStudio key="content" />}
            {activeTab === 'proposals' && <ProposalsView key="proposals" />}
            {activeTab === 'pricing' && <PricingView key="pricing" />}
          </AnimatePresence>
        </div>
      </main>
    </div>
  )
}

// Overview Component
function OverviewView({ stats, onNavigate }: { stats: DashboardStats, onNavigate: (tab: string) => void }) {
  const statCards = [
    { 
      title: 'Total Leads', 
      value: stats.totalLeads, 
      subtext: `${stats.hotLeads} hot`,
      icon: Users, 
      color: 'bg-papabase-pink',
      onClick: () => onNavigate('leads')
    },
    { 
      title: 'Pending Tasks', 
      value: stats.pendingTasks, 
      subtext: `${stats.totalTasks} total`,
      icon: CheckSquare, 
      color: 'bg-papabase-green',
      onClick: () => onNavigate('tasks')
    },
    { 
      title: 'Websites Generated', 
      value: stats.websitesGenerated, 
      subtext: 'AI-powered',
      icon: Sparkles, 
      color: 'bg-papabase-teal',
      onClick: () => onNavigate('websites')
    },
    { 
      title: 'AI Usage', 
      value: (stats.monthlyUsage / 1000).toFixed(1) + 'K', 
      subtext: 'tokens this month',
      icon: Zap, 
      color: 'bg-papabase-purple',
      onClick: () => onNavigate('pricing')
    },
  ]

  const quickActions = [
    { label: 'Add Lead', icon: Plus, tab: 'leads', color: 'papabase-pink' },
    { label: 'New Task', icon: CheckSquare, tab: 'tasks', color: 'papabase-green' },
    { label: 'Generate Website', icon: Sparkles, tab: 'websites', color: 'papabase-teal' },
    { label: 'Create Proposal', icon: FileText, tab: 'proposals', color: 'papabase-purple' },
  ]

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -20 }}
      className="space-y-8"
    >
      <div>
        <h1 className="text-3xl font-bold text-gray-800">Welcome back! 👋</h1>
        <p className="text-gray-600 mt-1">Here's what's happening with your family studio</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((card) => {
          const Icon = card.icon
          return (
            <button
              key={card.title}
              onClick={card.onClick}
              className="bg-white p-6 rounded-2xl shadow-sm hover:shadow-lg transition-shadow text-left"
            >
              <div className="flex items-center justify-between mb-4">
                <div className={`${card.color} p-3 rounded-xl text-white`}>
                  <Icon size={24} />
                </div>
                <TrendingUp size={20} className="text-green-500" />
              </div>
              <p className="text-3xl font-bold text-gray-800">{card.value}</p>
              <p className="text-sm text-gray-500 mt-1">{card.title}</p>
              <p className="text-xs text-papabase-purple font-medium mt-1">{card.subtext}</p>
            </button>
          )
        })}
      </div>

      {/* Quick Actions */}
      <div>
        <h2 className="text-xl font-bold text-gray-800 mb-4">Quick Actions</h2>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {quickActions.map((action) => {
            const Icon = action.icon
            return (
              <button
                key={action.label}
                onClick={() => onNavigate(action.tab)}
                className={`p-4 rounded-xl bg-white border-2 border-${action.color} text-${action.color} hover:bg-${action.color} hover:text-white transition-all flex items-center justify-center gap-2 font-medium`}
              >
                <Icon size={20} />
                {action.label}
              </button>
            )
          })}
        </div>
      </div>

      {/* Recent Activity Placeholder */}
      <div className="bg-white rounded-2xl p-6 shadow-sm">
        <h2 className="text-xl font-bold text-gray-800 mb-4">Getting Started Guide</h2>
        <div className="space-y-4">
          {[
            { step: 1, title: 'Add your first lead', desc: 'Start building your pipeline', tab: 'leads' },
            { step: 2, title: 'Create a task', desc: 'Stay organized and on track', tab: 'tasks' },
            { step: 3, title: 'Generate a website', desc: 'Let Dad AI build your site', tab: 'websites' },
            { step: 4, title: 'Create a proposal', desc: 'Win more clients faster', tab: 'proposals' },
          ].map((item) => (
            <button
              key={item.step}
              onClick={() => onNavigate(item.tab)}
              className="w-full flex items-center gap-4 p-4 rounded-xl hover:bg-gray-50 transition-colors text-left"
            >
              <div className="w-8 h-8 rounded-full bg-papabase-purple text-white flex items-center justify-center font-bold">
                {item.step}
              </div>
              <div>
                <p className="font-medium text-gray-800">{item.title}</p>
                <p className="text-sm text-gray-500">{item.desc}</p>
              </div>
            </button>
          ))}
        </div>
      </div>
    </motion.div>
  )
}
