'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { contentApi } from '@/lib/api'
import toast from 'react-hot-toast'
import { Type, FileText, Hash, MessageSquare, HelpCircle, Sparkles, Copy } from 'lucide-react'

export default function ContentStudio() {
  const [activeTool, setActiveTool] = useState('business')
  const [isGenerating, setIsGenerating] = useState(false)
  const [result, setResult] = useState('')

  const tools = [
    { id: 'business', name: 'Business Description', icon: FileText },
    { id: 'seo', name: 'SEO Meta Tags', icon: Hash },
    { id: 'blog', name: 'Blog Post', icon: Type },
    { id: 'social', name: 'Social Media', icon: MessageSquare },
    { id: 'faq', name: 'FAQ Generator', icon: HelpCircle },
  ]

  const [formData, setFormData] = useState({
    business_name: '',
    industry: '',
    key_services: '',
    topic: '',
    outline: '',
    tone: 'professional',
    page_content: '',
    keywords: '',
    business_type: '',
    common_questions: '',
  })

  const handleGenerate = async () => {
    setIsGenerating(true)
    try {
      let data
      switch (activeTool) {
        case 'business':
          data = await contentApi.businessDescription(formData)
          setResult(data.content?.content || 'Generated content...')
          break
        case 'seo':
          data = await contentApi.seoMeta(formData)
          setResult(JSON.stringify(data.meta_tags || {}, null, 2))
          break
        case 'blog':
          data = await contentApi.blog(formData)
          setResult(data.blog_post?.content || 'Generated blog...')
          break
        case 'social':
          data = await contentApi.social(formData)
          setResult(JSON.stringify(data.posts || [], null, 2))
          break
        case 'faq':
          data = await contentApi.faq(formData)
          setResult(JSON.stringify(data.faqs || [], null, 2))
          break
      }
      toast.success('Content generated!')
    } catch (error) {
      toast.error('Failed to generate content')
    } finally {
      setIsGenerating(false)
    }
  }

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Content Studio</h1>
        <p className="text-gray-600">AI-powered content generation for your business</p>
      </div>

      <div className="grid lg:grid-cols-4 gap-6">
        {/* Tools Sidebar */}
        <div className="lg:col-span-1 space-y-2">
          {tools.map((tool) => {
            const Icon = tool.icon
            return (
              <button
                key={tool.id}
                onClick={() => { setActiveTool(tool.id); setResult('') }}
                className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${
                  activeTool === tool.id
                    ? 'bg-papabase-pink text-white shadow'
                    : 'bg-white text-gray-600 hover:bg-gray-50'
                }`}
              >
                <Icon size={20} />
                <span className="font-medium">{tool.name}</span>
              </button>
            )
          })}
        </div>

        {/* Main Content Area */}
        <div className="lg:col-span-3 space-y-6">
          {/* Input Form */}
          <div className="bg-white p-6 rounded-2xl shadow-sm">
            {activeTool === 'business' && (
              <>
                <h2 className="text-xl font-bold text-gray-800 mb-4">Business Description Generator</h2>
                <div className="space-y-4">
                  <input type="text" placeholder="Business Name" value={formData.business_name} onChange={(e) => setFormData({...formData, business_name: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink" />
                  <input type="text" placeholder="Industry" value={formData.industry} onChange={(e) => setFormData({...formData, industry: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink" />
                  <textarea placeholder="Key Services (comma separated)" value={formData.key_services} onChange={(e) => setFormData({...formData, key_services: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink" rows={3} />
                </div>
              </>
            )}

            {activeTool === 'seo' && (
              <>
                <h2 className="text-xl font-bold text-gray-800 mb-4">SEO Meta Tags Generator</h2>
                <div className="space-y-4">
                  <textarea placeholder="Page Content" value={formData.page_content} onChange={(e) => setFormData({...formData, page_content: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink" rows={4} />
                  <input type="text" placeholder="Target Keywords" value={formData.keywords} onChange={(e) => setFormData({...formData, keywords: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink" />
                </div>
              </>
            )}

            {activeTool === 'blog' && (
              <>
                <h2 className="text-xl font-bold text-gray-800 mb-4">Blog Post Generator</h2>
                <div className="space-y-4">
                  <input type="text" placeholder="Topic" value={formData.topic} onChange={(e) => setFormData({...formData, topic: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink" />
                  <textarea placeholder="Outline" value={formData.outline} onChange={(e) => setFormData({...formData, outline: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink" rows={3} />
                  <select value={formData.tone} onChange={(e) => setFormData({...formData, tone: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink">
                    <option value="professional">Professional</option>
                    <option value="casual">Casual</option>
                    <option value="friendly">Friendly</option>
                    <option value="authoritative">Authoritative</option>
                  </select>
                </div>
              </>
            )}

            {activeTool === 'social' && (
              <>
                <h2 className="text-xl font-bold text-gray-800 mb-4">Social Media Posts</h2>
                <div className="space-y-4">
                  <input type="text" placeholder="Topic" value={formData.topic} onChange={(e) => setFormData({...formData, topic: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink" />
                  <input type="text" placeholder="Platforms (e.g., Twitter, LinkedIn, Instagram)" value={formData.topic} onChange={(e) => setFormData({...formData, topic: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink" />
                </div>
              </>
            )}

            {activeTool === 'faq' && (
              <>
                <h2 className="text-xl font-bold text-gray-800 mb-4">FAQ Generator</h2>
                <div className="space-y-4">
                  <input type="text" placeholder="Business Type" value={formData.business_type} onChange={(e) => setFormData({...formData, business_type: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink" />
                  <textarea placeholder="Common Questions/Topics" value={formData.common_questions} onChange={(e) => setFormData({...formData, common_questions: e.target.value})} className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:border-papabase-pink" rows={3} />
                </div>
              </>
            )}

            <button onClick={handleGenerate} disabled={isGenerating} className="mt-6 w-full py-4 bg-gradient-to-r from-papabase-pink to-papabase-purple text-white rounded-xl font-semibold hover:shadow-lg disabled:opacity-50 flex items-center justify-center gap-2">
              <Sparkles size={20} />
              {isGenerating ? 'Generating...' : 'Generate Content'}
            </button>
          </div>

          {/* Output */}
          {result && (
            <div className="bg-white p-6 rounded-2xl shadow-sm">
              <div className="flex items-center justify-between mb-4">
                <h3 className="font-bold text-gray-800">Generated Content</h3>
                <button onClick={() => { navigator.clipboard.writeText(result); toast.success('Copied!') }} className="flex items-center gap-2 text-sm text-papabase-purple hover:underline">
                  <Copy size={16} /> Copy
                </button>
              </div>
              <pre className="bg-gray-50 p-4 rounded-xl overflow-x-auto text-sm whitespace-pre-wrap">{result}</pre>
            </div>
          )}
        </div>
      </div>
    </motion.div>
  )
}
