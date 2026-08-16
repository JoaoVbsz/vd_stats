/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        vd: {
          900: '#0a0a0a',
          800: '#141414',
          700: '#1e1e1e',
          600: '#2d2d2d',
          yellow: '#fbbf24',
          green: '#22c55e',
          text: '#e5e5e5',
          textMuted: '#a3a3a3',
        }
      }
    },
  },
  plugins: [],
}
