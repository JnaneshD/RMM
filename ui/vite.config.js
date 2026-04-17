import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { Agent } from 'node:https'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/api": {
        target: "https://localhost:8081",
        changeOrigin: true,
        secure: false, // skips cert validation — fine for dev
        rewrite: (path) => path.replace(/^\/api/, ""),
        agent: new Agent({ rejectUnauthorized: false }), // Remove this for security
      },
    },
  },
})