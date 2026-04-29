import type { Metadata } from 'next'
import { Poppins } from 'next/font/google'
import './globals.css'
import { Toaster } from 'react-hot-toast'

const poppins = Poppins({
  weight: ['300', '400', '500', '600', '700'],
  subsets: ['latin'],
  variable: '--font-poppins',
})

export const metadata: Metadata = {
  title: 'Papabase - Family Studio Suite',
  description: 'AI-powered business operations platform for families and young professionals',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className={poppins.variable}>
      <body className="font-sans antialiased">
        {children}
        <Toaster 
          position="top-right"
          toastOptions={{
            duration: 4000,
            style: {
              background: '#643277',
              color: '#fff',
              fontFamily: 'var(--font-poppins)',
            },
            success: {
              style: {
                background: '#316844',
              },
            },
            error: {
              style: {
                background: '#EC89A3',
              },
            },
          }}
        />
      </body>
    </html>
  )
}
