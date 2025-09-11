import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      // Proxy API requests to the Go backend
      '/v1': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          // Separate React Flow into its own chunk for better caching
          'react-flow': ['@xyflow/react'],
          // Group vendor libraries
          vendor: ['react', 'react-dom'],
          // Group state management
          store: ['zustand'],
        },
      },
    },
  },
  preview: {
    port: 3000,
  },
})
