'use client'

import { useState } from 'react'
import { Upload, FileText, AlertTriangle, CheckCircle, Loader2 } from 'lucide-react'

export default function Home() {
  const [uploading, setUploading] = useState(false)
  const [uploadProgress, setUploadProgress] = useState(0)
  const [project, setProject] = useState<{
    id: string
    name: string
    status: string
    elements: number
    issues: number
  } | null>(null)

  const handleUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    setUploading(true)
    setUploadProgress(0)

    // Simulate upload progress
    const progressInterval = setInterval(() => {
      setUploadProgress(prev => {
        if (prev >= 90) {
          clearInterval(progressInterval)
          return 90
        }
        return prev + 10
      })
    }, 200)

    try {
      // In production, would upload to BIM ingestion service
      const formData = new FormData()
      formData.append('files', file)

      // Mock API call
      await new Promise(resolve => setTimeout(resolve, 2000))

      // Simulate successful upload
      setProject({
        id: 'proj_' + Date.now(),
        name: file.name,
        status: 'analyzing',
        elements: 0,
        issues: 0,
      })

      clearInterval(progressInterval)
      setUploadProgress(100)

      // Start analysis workflow
      await startAnalysisWorkflow('proj_' + Date.now())

    } catch (error) {
      console.error('Upload failed:', error)
    } finally {
      setUploading(false)
    }
  }

  const startAnalysisWorkflow = async (projectId: string) => {
    // In production, would call orchestration service
    console.log('Starting analysis workflow for', projectId)
    
    // Simulate workflow progress
    setTimeout(() => {
      setProject(prev => prev ? {
        ...prev,
        status: 'completed',
        elements: Math.floor(Math.random() * 500) + 100,
        issues: Math.floor(Math.random() * 20),
      } : null)
    }, 3000)
  }

  return (
    <main className="min-h-screen bg-gradient-to-b from-background to-muted">
      {/* Header */}
      <header className="border-b bg-card">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center">
                <FileText className="w-6 h-6 text-primary-foreground" />
              </div>
              <div>
                <h1 className="text-xl font-bold">FullStackArkham</h1>
                <p className="text-sm text-muted-foreground">BIM Analysis Platform</p>
              </div>
            </div>
            <nav className="flex items-center gap-4">
              <a href="/projects" className="text-sm text-muted-foreground hover:text-foreground">
                Projects
              </a>
              <a href="/workflows" className="text-sm text-muted-foreground hover:text-foreground">
                Workflows
              </a>
              <a href="/settings" className="text-sm text-muted-foreground hover:text-foreground">
                Settings
              </a>
            </nav>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="container mx-auto px-4 py-8">
        {/* Upload Section */}
        <section className="mb-8">
          <div className="bg-card rounded-lg border p-6">
            <h2 className="text-lg font-semibold mb-4">Upload BIM Project</h2>
            
            <div className="border-2 border-dashed rounded-lg p-8 text-center hover:border-primary/50 transition-colors">
              <input
                type="file"
                accept=".ifc,.ifczip,.rvt,.dwg"
                onChange={handleUpload}
                disabled={uploading}
                className="hidden"
                id="file-upload"
              />
              <label htmlFor="file-upload" className="cursor-pointer">
                <Upload className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
                <p className="text-sm text-muted-foreground mb-2">
                  Drop your IFC file here or click to browse
                </p>
                <p className="text-xs text-muted-foreground">
                  Supported formats: IFC, RVT, DWG (max 100MB)
                </p>
              </label>

              {uploading && (
                <div className="mt-4">
                  <div className="flex items-center justify-center gap-2 mb-2">
                    <Loader2 className="w-4 h-4 animate-spin" />
                    <span className="text-sm text-muted-foreground">Uploading...</span>
                  </div>
                  <div className="w-full max-w-md mx-auto bg-muted rounded-full h-2">
                    <div 
                      className="bg-primary h-2 rounded-full transition-all"
                      style={{ width: `${uploadProgress}%` }}
                    />
                  </div>
                </div>
              )}
            </div>
          </div>
        </section>

        {/* Project Status */}
        {project && (
          <section className="mb-8">
            <div className="bg-card rounded-lg border p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold">Project Status</h2>
                <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                  project.status === 'completed' 
                    ? 'bg-green-100 text-green-800' 
                    : 'bg-blue-100 text-blue-800'
                }`}>
                  {project.status}
                </span>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="bg-muted rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <FileText className="w-5 h-5 text-muted-foreground" />
                    <span className="text-sm text-muted-foreground">Project Name</span>
                  </div>
                  <p className="text-lg font-medium truncate">{project.name}</p>
                </div>

                <div className="bg-muted rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <CheckCircle className="w-5 h-5 text-muted-foreground" />
                    <span className="text-sm text-muted-foreground">Elements</span>
                  </div>
                  <p className="text-lg font-medium">{project.elements}</p>
                </div>

                <div className="bg-muted rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <AlertTriangle className="w-5 h-5 text-muted-foreground" />
                    <span className="text-sm text-muted-foreground">Issues</span>
                  </div>
                  <p className="text-lg font-medium">{project.issues}</p>
                </div>
              </div>

              {project.status === 'completed' && (
                <div className="mt-4 flex gap-3">
                  <button className="px-4 py-2 bg-primary text-primary-foreground rounded-md text-sm font-medium hover:bg-primary/90">
                    View Report
                  </button>
                  <button className="px-4 py-2 bg-muted text-foreground rounded-md text-sm font-medium hover:bg-muted/80">
                    Export Results
                  </button>
                </div>
              )}
            </div>
          </section>
        )}

        {/* Recent Projects */}
        <section>
          <div className="bg-card rounded-lg border p-6">
            <h2 className="text-lg font-semibold mb-4">Recent Projects</h2>
            <div className="text-sm text-muted-foreground text-center py-8">
              No projects yet. Upload your first BIM file to get started.
            </div>
          </div>
        </section>
      </div>
    </main>
  )
}
