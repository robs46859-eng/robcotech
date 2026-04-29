'use client'

import { useEffect, useState } from 'react'
import { useAuthStore } from '@/lib/store'
import AuthPage from '@/components/AuthPage'
import OnboardingQuestionnaire from '@/components/OnboardingQuestionnaire'
import Dashboard from '@/app/dashboard/page'

export default function Home() {
  const { isAuthenticated, user } = useAuthStore()
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) {
    return (
      <div className="min-h-screen gradient-main flex items-center justify-center">
        <div className="text-white text-xl">Loading...</div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return <AuthPage />
  }

  if (user && !user.onboardingComplete) {
    return <OnboardingQuestionnaire />
  }

  return <Dashboard />
}
