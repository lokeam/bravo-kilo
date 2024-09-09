/// <reference types="vitest" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true, // Enable Jest-like global functions
    environment: 'jsdom', // Simulates a DOM environment
    setupFiles: './vitest.setup.ts', // Similar to Jest's setupFiles
  },
});
