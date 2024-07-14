/** @type {import('tailwindcss').Config} */
export default {
  content: ["./app/**/*.{js,jsx,ts,tsx}"],
  theme: {
    extend: {
      colors: {
        primary: {"50":"#eff6ff","100":"#dbeafe","200":"#bfdbfe","300":"#93c5fd","400":"#60a5fa","500":"#3b82f6","600":"#2563eb","700":"#1d4ed8","800":"#1e40af","900":"#1e3a8a","950":"#172554"},
        'dark-gunmetal': '#182235', // top nav bg
        'yankees-blue': '#1E293B', // sidebar bg,
        'ebony': '#0F172A', // main dark bg
        'charcoal': '#334155', // top nav/other border color
        'hepatica': '#6366F1', // bright purple, cta btn
        'majorelle': '#4F46E5', // dark purple, cta btn, white text (#fff)\
        'anti-flash-white': '#F1F5F9', // heading tags
        'bright-gray': '#E5E7EB', // white alternative heading / paragraph tag
        'az-white': '#E2E8F0', // sidebar text white, hover
        'mystic-white':'#DBE2EB', // sidebar text white, disabled
        'cadet-gray': '#94A3B8', // form field label, subheading text, breadcrumb

      }
    },
  },
  plugins: [],
}

