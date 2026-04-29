'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { useAuthStore } from '@/lib/store'
import { useRouter } from 'next/navigation'
import toast from 'react-hot-toast'
import { Users, MapPin, Briefcase, Heart, Clock, ArrowRight, ArrowLeft } from 'lucide-react'

interface FormData {
  role: string
  geography: string
  businessExperience: string
  familyDesires: string
  familySize: number
  hoursPerWeek: number
}

const steps = [
  { id: 'role', title: 'Who are you?', icon: Users },
  { id: 'geography', title: 'Where are you located?', icon: MapPin },
  { id: 'experience', title: 'Business Experience', icon: Briefcase },
  { id: 'desires', title: 'Family Goals', icon: Heart },
  { id: 'family', title: 'Family Size', icon: Users },
  { id: 'hours', title: 'Time Commitment', icon: Clock },
]

export default function OnboardingQuestionnaire() {
  const router = useRouter()
  const { completeOnboarding } = useAuthStore()
  const [currentStep, setCurrentStep] = useState(0)
  const [formData, setFormData] = useState<FormData>({
    role: '',
    geography: '',
    businessExperience: '',
    familyDesires: '',
    familySize: 1,
    hoursPerWeek: 10,
  })
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleNext = () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1)
    } else {
      handleSubmit()
    }
  }

  const handleBack = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1)
    }
  }

  const handleSubmit = async () => {
    setIsSubmitting(true)
    try {
      completeOnboarding({
        role: formData.role as any,
        geography: formData.geography,
        businessExperience: formData.businessExperience,
        familyDesires: formData.familyDesires,
        familySize: formData.familySize,
        hoursPerWeek: formData.hoursPerWeek,
      })
      toast.success('Welcome to Papabase!')
      router.push('/dashboard')
    } catch (error) {
      toast.error('Failed to complete onboarding')
    } finally {
      setIsSubmitting(false)
    }
  }

  const canProceed = () => {
    switch (currentStep) {
      case 0: return formData.role !== ''
      case 1: return formData.geography !== ''
      case 2: return formData.businessExperience !== ''
      case 3: return formData.familyDesires !== ''
      case 4: return formData.familySize >= 1
      case 5: return formData.hoursPerWeek >= 1
      default: return false
    }
  }

  const progress = ((currentStep + 1) / steps.length) * 100

  return (
    <div className="min-h-screen gradient-main flex items-center justify-center p-4">
      <motion.div 
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="w-full max-w-2xl bg-white rounded-3xl shadow-2xl overflow-hidden"
      >
        {/* Progress Bar */}
        <div className="bg-papabase-purple h-2">
          <motion.div 
            className="h-full bg-papabase-teal"
            initial={{ width: 0 }}
            animate={{ width: `${progress}%` }}
            transition={{ duration: 0.3 }}
          />
        </div>

        {/* Step Indicators */}
        <div className="flex justify-between px-8 py-6 bg-gray-50">
          {steps.map((step, index) => {
            const Icon = step.icon
            return (
              <div key={step.id} className="flex flex-col items-center">
                <div className={`w-10 h-10 rounded-full flex items-center justify-center transition-colors ${
                  index <= currentStep 
                    ? 'bg-papabase-purple text-white' 
                    : 'bg-gray-200 text-gray-400'
                }`}>
                  <Icon size={18} />
                </div>
                <span className={`text-xs mt-1 hidden sm:block ${
                  index <= currentStep ? 'text-papabase-purple font-medium' : 'text-gray-400'
                }`}>
                  {step.title.split(' ')[0]}
                </span>
              </div>
            )
          })}
        </div>

        {/* Content */}
        <div className="p-8">
          <AnimatePresence mode="wait">
            {/* Step 1: Role */}
            {currentStep === 0 && (
              <motion.div
                key="role"
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
              >
                <h2 className="text-2xl font-bold text-papabase-purple mb-2">Who are you?</h2>
                <p className="text-gray-600 mb-6">This helps us personalize your experience</p>
                
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  {[
                    { id: 'parent', label: 'Parent', desc: 'Building for my family', emoji: '👨‍👩‍👧‍👦' },
                    { id: 'child', label: 'Young Adult', desc: 'Starting my journey', emoji: '🚀' },
                    { id: 'other', label: 'Other', desc: 'Something else', emoji: '✨' },
                  ].map((option) => (
                    <button
                      key={option.id}
                      onClick={() => setFormData({ ...formData, role: option.id })}
                      className={`p-6 rounded-xl border-2 transition-all ${
                        formData.role === option.id
                          ? 'border-papabase-purple bg-papabase-pinkLight'
                          : 'border-gray-200 hover:border-papabase-teal'
                      }`}
                    >
                      <span className="text-4xl mb-2 block">{option.emoji}</span>
                      <span className="font-semibold text-gray-800">{option.label}</span>
                      <span className="text-sm text-gray-500 block mt-1">{option.desc}</span>
                    </button>
                  ))}
                </div>
              </motion.div>
            )}

            {/* Step 2: Geography */}
            {currentStep === 1 && (
              <motion.div
                key="geography"
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
              >
                <h2 className="text-2xl font-bold text-papabase-purple mb-2">Where are you located?</h2>
                <p className="text-gray-600 mb-6">This helps us connect you with local resources</p>
                
                <input
                  type="text"
                  value={formData.geography}
                  onChange={(e) => setFormData({ ...formData, geography: e.target.value })}
                  placeholder="City, State/Country"
                  className="w-full p-4 border-2 border-gray-200 rounded-xl focus:border-papabase-purple focus:outline-none text-lg"
                  autoFocus
                />
                
                <div className="mt-4 flex flex-wrap gap-2">
                  {['United States', 'United Kingdom', 'Canada', 'Australia', 'Other'].map((country) => (
                    <button
                      key={country}
                      onClick={() => setFormData({ ...formData, geography: country })}
                      className={`px-4 py-2 rounded-full text-sm transition-colors ${
                        formData.geography === country
                          ? 'bg-papabase-purple text-white'
                          : 'bg-gray-100 hover:bg-papabase-teal hover:text-white'
                      }`}
                    >
                      {country}
                    </button>
                  ))}
                </div>
              </motion.div>
            )}

            {/* Step 3: Experience */}
            {currentStep === 2 && (
              <motion.div
                key="experience"
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
              >
                <h2 className="text-2xl font-bold text-papabase-purple mb-2">What's your business experience?</h2>
                <p className="text-gray-600 mb-6">Be honest - we'll tailor everything to your level</p>
                
                <div className="space-y-3">
                  {[
                    { id: 'beginner', label: 'Just Starting', desc: 'No experience yet, eager to learn', icon: '🌱' },
                    { id: 'some', label: 'Some Experience', desc: 'Tried a few things before', icon: '🌿' },
                    { id: 'experienced', label: 'Experienced', desc: 'Run/have run businesses', icon: '🌳' },
                    { id: 'expert', label: 'Expert', desc: 'Multiple successful ventures', icon: '🏆' },
                  ].map((option) => (
                    <button
                      key={option.id}
                      onClick={() => setFormData({ ...formData, businessExperience: option.id })}
                      className={`w-full p-4 rounded-xl border-2 text-left transition-all ${
                        formData.businessExperience === option.id
                          ? 'border-papabase-purple bg-papabase-pinkLight'
                          : 'border-gray-200 hover:border-papabase-teal'
                      }`}
                    >
                      <span className="text-2xl mr-3">{option.icon}</span>
                      <span className="font-semibold text-gray-800">{option.label}</span>
                      <span className="text-sm text-gray-500 block mt-1">{option.desc}</span>
                    </button>
                  ))}
                </div>
              </motion.div>
            )}

            {/* Step 4: Family Desires */}
            {currentStep === 3 && (
              <motion.div
                key="desires"
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
              >
                <h2 className="text-2xl font-bold text-papabase-purple mb-2">What do you desire for your family?</h2>
                <p className="text-gray-600 mb-6">Your goals help us suggest the right path</p>
                
                <textarea
                  value={formData.familyDesires}
                  onChange={(e) => setFormData({ ...formData, familyDesires: e.target.value })}
                  placeholder="What kind of future are you building? What does success look like for your family?"
                  className="w-full p-4 border-2 border-gray-200 rounded-xl focus:border-papabase-purple focus:outline-none text-lg min-h-[200px]"
                  autoFocus
                />
                
                <div className="mt-4 flex flex-wrap gap-2">
                  {['Financial Freedom', 'More Family Time', 'Build Legacy', 'Help Others', 'Location Freedom'].map((goal) => (
                    <button
                      key={goal}
                      onClick={() => setFormData({ 
                        ...formData, 
                        familyDesires: formData.familyDesires 
                          ? `${formData.familyDesires}, ${goal}` 
                          : goal 
                      })}
                      className="px-4 py-2 rounded-full text-sm bg-gray-100 hover:bg-papabase-teal hover:text-white transition-colors"
                    >
                      + {goal}
                    </button>
                  ))}
                </div>
              </motion.div>
            )}

            {/* Step 5: Family Size */}
            {currentStep === 4 && (
              <motion.div
                key="family"
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
              >
                <h2 className="text-2xl font-bold text-papabase-purple mb-2">How big is your family?</h2>
                <p className="text-gray-600 mb-6">Including yourself</p>
                
                <div className="flex items-center justify-center gap-8 py-8">
                  <button
                    onClick={() => setFormData({ ...formData, familySize: Math.max(1, formData.familySize - 1) })}
                    className="w-16 h-16 rounded-full bg-papabase-pink text-white text-3xl font-bold hover:bg-papabase-purple transition-colors"
                  >
                    -
                  </button>
                  <div className="text-center">
                    <span className="text-6xl font-bold text-papabase-purple">{formData.familySize}</span>
                    <span className="block text-gray-500 mt-2">family member{formData.familySize !== 1 ? 's' : ''}</span>
                  </div>
                  <button
                    onClick={() => setFormData({ ...formData, familySize: formData.familySize + 1 })}
                    className="w-16 h-16 rounded-full bg-papabase-teal text-white text-3xl font-bold hover:bg-papabase-purple transition-colors"
                  >
                    +
                  </button>
                </div>
                
                <div className="flex justify-center gap-4 text-sm text-gray-500">
                  <span>👤 Just me</span>
                  <span>👥 Small (2-4)</span>
                  <span>👨‍👩‍👧‍👦 Large (5+)</span>
                </div>
              </motion.div>
            )}

            {/* Step 6: Hours */}
            {currentStep === 5 && (
              <motion.div
                key="hours"
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
              >
                <h2 className="text-2xl font-bold text-papabase-purple mb-2">How many hours per week?</h2>
                <p className="text-gray-600 mb-6">Be realistic - consistency beats intensity</p>
                
                <div className="py-8">
                  <input
                    type="range"
                    min="1"
                    max="80"
                    value={formData.hoursPerWeek}
                    onChange={(e) => setFormData({ ...formData, hoursPerWeek: parseInt(e.target.value) })}
                    className="w-full h-4 bg-gray-200 rounded-lg appearance-none cursor-pointer accent-papabase-purple"
                    autoFocus
                  />
                  <div className="flex justify-between mt-4 text-sm text-gray-500">
                    <span>1 hr</span>
                    <span>20 hrs</span>
                    <span>40 hrs</span>
                    <span>80 hrs</span>
                  </div>
                </div>
                
                <div className="text-center py-6">
                  <span className="text-5xl font-bold text-papabase-teal">{formData.hoursPerWeek}</span>
                  <span className="text-gray-500 ml-2">hours per week</span>
                </div>
                
                <div className="grid grid-cols-3 gap-4 mt-4">
                  <div className={`p-4 rounded-xl text-center ${formData.hoursPerWeek <= 10 ? 'bg-papabase-pinkLight' : 'bg-gray-100'}`}>
                    <span className="text-2xl">🌱</span>
                    <span className="block text-sm font-medium">Part-time</span>
                    <span className="text-xs text-gray-500">1-10 hrs</span>
                  </div>
                  <div className={`p-4 rounded-xl text-center ${formData.hoursPerWeek > 10 && formData.hoursPerWeek <= 30 ? 'bg-papabase-tealLight' : 'bg-gray-100'}`}>
                    <span className="text-2xl">🌿</span>
                    <span className="block text-sm font-medium">Serious</span>
                    <span className="text-xs text-gray-500">11-30 hrs</span>
                  </div>
                  <div className={`p-4 rounded-xl text-center ${formData.hoursPerWeek > 30 ? 'bg-papabase-purple text-white' : 'bg-gray-100'}`}>
                    <span className="text-2xl">🚀</span>
                    <span className="block text-sm font-medium">All-in</span>
                    <span className="text-xs opacity-70">31+ hrs</span>
                  </div>
                </div>
              </motion.div>
            )}
          </AnimatePresence>

          {/* Navigation */}
          <div className="flex justify-between mt-8 pt-6 border-t">
            <button
              onClick={handleBack}
              disabled={currentStep === 0}
              className={`flex items-center gap-2 px-6 py-3 rounded-xl transition-colors ${
                currentStep === 0
                  ? 'text-gray-300 cursor-not-allowed'
                  : 'text-papabase-purple hover:bg-gray-100'
              }`}
            >
              <ArrowLeft size={20} />
              Back
            </button>
            
            <button
              onClick={handleNext}
              disabled={!canProceed() || isSubmitting}
              className={`flex items-center gap-2 px-8 py-3 rounded-xl font-semibold transition-all ${
                canProceed() && !isSubmitting
                  ? 'bg-papabase-purple text-white hover:bg-papabase-green shadow-lg hover:shadow-xl'
                  : 'bg-gray-200 text-gray-400 cursor-not-allowed'
              }`}
            >
              {isSubmitting ? 'Saving...' : currentStep === steps.length - 1 ? "Let's Go!" : 'Continue'}
              <ArrowRight size={20} />
            </button>
          </div>
        </div>
      </motion.div>
    </div>
  )
}
