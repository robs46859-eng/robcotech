import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface UserProfile {
  id: string
  email: string
  name: string
  role: 'parent' | 'child' | 'other'
  geography: string
  businessExperience: string
  familyDesires: string
  familySize: number
  hoursPerWeek: number
  tenantId: string
  plan: 'starter' | 'studio' | 'agency' | 'enterprise'
  onboardingComplete: boolean
}

interface AuthState {
  user: UserProfile | null
  token: string | null
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<void>
  signup: (data: Partial<UserProfile> & { password: string }) => Promise<void>
  completeOnboarding: (data: Partial<UserProfile>) => void
  logout: () => void
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL || '/api/v1'

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,

      login: async (email: string, password: string) => {
        try {
          const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password }),
          })
          
          if (!response.ok) {
            const error = await response.json()
            throw new Error(error.message || 'Login failed')
          }
          
          const data = await response.json()
          set({
            user: data.user,
            token: data.token,
            isAuthenticated: true,
          })
        } catch (error) {
          // For demo, create a mock user
          const mockUser: UserProfile = {
            id: 'user-' + Date.now(),
            email,
            name: email.split('@')[0],
            role: 'parent',
            geography: 'United States',
            businessExperience: 'beginner',
            familyDesires: 'Build a family business',
            familySize: 4,
            hoursPerWeek: 20,
            tenantId: 'tenant-' + Date.now(),
            plan: 'starter',
            onboardingComplete: false,
          }
          set({
            user: mockUser,
            token: 'demo-token',
            isAuthenticated: true,
          })
        }
      },

      signup: async (data: Partial<UserProfile> & { password: string }) => {
        try {
          const response = await fetch(`${API_BASE}/auth/signup`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
          })
          
          if (!response.ok) {
            const error = await response.json()
            throw new Error(error.message || 'Signup failed')
          }
          
          const result = await response.json()
          set({
            user: result.user,
            token: result.token,
            isAuthenticated: true,
          })
        } catch (error) {
          // For demo, create user locally
          const mockUser: UserProfile = {
            id: 'user-' + Date.now(),
            email: data.email || '',
            name: data.name || '',
            role: data.role || 'parent',
            geography: data.geography || '',
            businessExperience: data.businessExperience || 'beginner',
            familyDesires: data.familyDesires || '',
            familySize: data.familySize || 1,
            hoursPerWeek: data.hoursPerWeek || 10,
            tenantId: 'tenant-' + Date.now(),
            plan: 'starter',
            onboardingComplete: false,
          }
          set({
            user: mockUser,
            token: 'demo-token',
            isAuthenticated: true,
          })
        }
      },

      completeOnboarding: (data: Partial<UserProfile>) => {
        const currentUser = get().user
        if (currentUser) {
          set({
            user: {
              ...currentUser,
              ...data,
              onboardingComplete: true,
            },
          })
        }
      },

      logout: () => {
        set({
          user: null,
          token: null,
          isAuthenticated: false,
        })
      },
    }),
    {
      name: 'papabase-auth',
      partialize: (state) => ({ 
        user: state.user, 
        token: state.token, 
        isAuthenticated: state.isAuthenticated 
      }),
    }
  )
)
