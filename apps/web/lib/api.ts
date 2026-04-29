const API_BASE = process.env.NEXT_PUBLIC_API_URL || '/api/v1'

async function fetchAPI(endpoint: string, options: RequestInit = {}) {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string> || {}),
  }

  // Add auth token if available
  const token = localStorage.getItem('papabase-token')
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers,
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }))
    throw new Error(error.message || `HTTP ${response.status}`)
  }

  return response.json()
}

// Lead API
export const leadsApi = {
  list: () => fetchAPI('/leads'),
  get: (id: string) => fetchAPI(`/leads/${id}`),
  create: (data: any) => fetchAPI('/leads', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: any) => fetchAPI(`/leads/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) => fetchAPI(`/leads/${id}`, { method: 'DELETE' }),
  score: (leadId: string) => fetchAPI('/ai/leads/score', { method: 'POST', body: JSON.stringify({ lead_id: leadId }) }),
  scoreBatch: () => fetchAPI('/ai/leads/score/batch', { method: 'POST' }),
  getInsights: (id: string) => fetchAPI(`/ai/leads/${id}/insights`),
}

// Task API
export const tasksApi = {
  list: () => fetchAPI('/tasks'),
  get: (id: string) => fetchAPI(`/tasks/${id}`),
  create: (data: any) => fetchAPI('/tasks', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: any) => fetchAPI(`/tasks/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) => fetchAPI(`/tasks/${id}`, { method: 'DELETE' }),
  generate: (note: string, context?: any) => fetchAPI('/ai/tasks/generate', { method: 'POST', body: JSON.stringify({ note, context }) }),
  breakdown: (title: string, description: string) => fetchAPI('/ai/tasks/breakdown', { method: 'POST', body: JSON.stringify({ title, description }) }),
}

// Dad AI Website API
export const websiteApi = {
  generate: (data: any) => fetchAPI('/ai/generate', { method: 'POST', body: JSON.stringify(data) }),
  templates: () => fetchAPI('/ai/templates'),
  projects: () => fetchAPI('/ai/projects'),
  getProject: (id: string) => fetchAPI(`/ai/projects/${id}`),
}

// Content API
export const contentApi = {
  businessDescription: (data: any) => fetchAPI('/ai/content/business-description', { method: 'POST', body: JSON.stringify(data) }),
  seoMeta: (data: any) => fetchAPI('/ai/content/seo-meta', { method: 'POST', body: JSON.stringify(data) }),
  blog: (data: any) => fetchAPI('/ai/content/blog', { method: 'POST', body: JSON.stringify(data) }),
  social: (data: any) => fetchAPI('/ai/content/social', { method: 'POST', body: JSON.stringify(data) }),
  faq: (data: any) => fetchAPI('/ai/content/faq', { method: 'POST', body: JSON.stringify(data) }),
}

// Proposal API
export const proposalApi = {
  generate: (data: any) => fetchAPI('/ai/proposals/generate', { method: 'POST', body: JSON.stringify(data) }),
  quote: (data: any) => fetchAPI('/ai/proposals/quote', { method: 'POST', body: JSON.stringify(data) }),
  scope: (data: any) => fetchAPI('/ai/proposals/scope', { method: 'POST', body: JSON.stringify(data) }),
  pricingTiers: (data: any) => fetchAPI('/ai/proposals/pricing-tiers', { method: 'POST', body: JSON.stringify(data) }),
}

// Pricing API
export const pricingApi = {
  plans: () => fetchAPI('/pricing/plans'),
  usage: () => fetchAPI('/billing/usage'),
  invoices: () => fetchAPI('/billing/invoices'),
}
