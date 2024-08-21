import { transform } from 'typescript';

/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,jsx,ts,tsx}"
  ],
  theme: {
    extend: {
      animation: {
        'fade-in': 'fadeIn 0.8 ease-out forwards',
        'fade-out': 'fadeOut 0.8s ease-out forwards',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0', transform: 'translateY(90px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        fadeOut: {
          '0%': { opacity: '1', transform: 'translateY(0)' },
          '100%': { opacity: '0', transform: 'translateY(90px)' },
        },
      },
      colors: {
        primary: {"50":"#eff6ff","100":"#dbeafe","200":"#bfdbfe","300":"#93c5fd","400":"#60a5fa","500":"#3b82f6","600":"#2563eb","700":"#1d4ed8","800":"#1e40af","900":"#1e3a8a","950":"#172554"},
        'dark-gunmetal': '#182235', // top nav bg
        'yankees-blue': '#1E293B', // sidebar bg,
        'ebony': '#0F172A', // main dark bg
        'night': '#111827', // potential dark bg
        'charcoal': '#334155', // top nav/other border color
        'hepatica': '#6366F1', // bright purple, cta btn
        'hepatica-lt': '#8470ff', // light purple
        'majorelle': '#4F46E5', // dark purple, cta btn, white text (#fff)
        'margorelle-d1': '#423ABF', // 1 shade darker than margorelle
        'margorelle-d2': '#352F99', // 2 shades darker than margorelle
        'margorelle-d3': '#282373', // 3 shades darker than margorelle
        'margorelle-d4': '#282373', // 4 shades darker than margorelle
        'margorelle-d5': '#1A174C', // 5 shades darker than margorelle
        'margorelle-d6': '#0D0C26', // 6 shades darker than margorelle
        'margorelle-comp1-r': '#bf3a9c', // red complementary to margorelle
        'margorelle-comp1-g': '#67bf3a', // green complementary to margorelle
        'anti-flash-white': '#F1F5F9', // heading tags
        'bright-gray': '#E5E7EB', // white alternative heading / paragraph tag
        'az-white': '#E2E8F0', // sidebar text white, hover
        'mystic-white':'#DBE2EB', // sidebar text white, disabled
        'cadet-gray': '#94A3B8', // form field label, subheading text, breadcrumb
        'maastricht': '#0d1e2f', // search bar background
        'ceil': '#93accd',
        'ebony-clay': '#253038',
        'polo-blue': '#93accd',
        'nevada-gray': '#606d79',
        'dark-ebony': '#0c0c0c',
        'dark-clay': '#1f2937',
        'maya-blue': '#67bfff33',
        'midnight-navy': '#374151',
      }
    },
    screens: {
      'lgMobile': '600px',
      'mdTablet': '940px',
      'mdDesktop': '1280px',
    }
  },
  plugins: [],
}

