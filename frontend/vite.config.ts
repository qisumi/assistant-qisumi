import react from '@vitejs/plugin-react'
import {defineConfig} from 'vite'
import {VitePWA} from 'vite-plugin-pwa'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  plugins : [
    react(), VitePWA({
      registerType : 'autoUpdate',
      includeAssets :
          [ 'favicon.ico', 'apple-touch-icon.png', 'masked-icon.svg' ],
      manifest : {
        name : 'Assistant Qisumi',
        short_name : 'Qisumi',
        description : 'AI-based Task Planning & Memo System',
        theme_color : '#ffffff',
        icons : [
          {src : 'pwa-192x192.png', sizes : '192x192', type : 'image/png'},
          {src : 'pwa-512x512.png', sizes : '512x512', type : 'image/png'}, {
            src : 'pwa-512x512.png',
            sizes : '512x512',
            type : 'image/png',
            purpose : 'any maskable'
          }
        ]
      }
    })
  ],
  server : {
    port : 3000,
    proxy : {'/api' : {target : 'http://localhost:4569', changeOrigin : true}}
  }
})
