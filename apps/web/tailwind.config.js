/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        papabase: {
          pink: '#EC89A3',
          green: '#316844',
          teal: '#3DC2B9',
          purple: '#643277',
          pinkLight: '#F9B8C8',
          greenLight: '#4A8B64',
          tealLight: '#6FD9D1',
          purpleLight: '#8B5A9E',
        },
      },
      fontFamily: {
        sans: ['var(--font-poppins)', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
