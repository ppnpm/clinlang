import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { VitePWA } from 'vite-plugin-pwa';
import path from 'node:path';

// ClinLang web app — Vite + React + PWA.
//
// During `npm run dev`, the dev server proxies /api/* to the Go binary
// on http://localhost:8080. In production the Go binary embeds web/dist
// and serves the SPA from the same origin as the API, so no proxy is
// needed (VITE_API_BASE stays empty).
export default defineConfig({
  // appType 'spa' is the default but stating it explicitly guarantees
  // the dev server falls back to index.html for any unknown path —
  // refreshing on /notes/foo loads the SPA instead of returning 404.
  appType: 'spa',
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    allowedHosts:['8c27-182-77-76-127.ngrok-free.app'],
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  plugins: [
    react(),
    VitePWA({
      registerType: 'prompt',
      includeAssets: ['icon.svg', 'favicon.ico', 'robots.txt'],
      manifest: {
        name: 'ClinLang',
        short_name: 'ClinLang',
        description:
          'Personal shorthand and templating tool for clinicians\' own notes. Not a medical device.',
        theme_color: '#0f172a',
        background_color: '#f8fafc',
        display: 'standalone',
        orientation: 'any',
        start_url: '/',
        scope: '/',
        icons: [
          {
            src: 'icon.svg',
            sizes: 'any',
            type: 'image/svg+xml',
            purpose: 'any maskable',
          },
        ],
      },
      workbox: {
        // Cache the SPA shell and static assets, but NEVER cache /api/*.
        // Clinical data must always hit the server fresh.
        globPatterns: ['**/*.{js,css,html,ico,svg,png}'],
        navigateFallback: '/index.html',
        navigateFallbackDenylist: [/^\/api\//],
        runtimeCaching: [
          {
            urlPattern: ({ url }) => url.pathname.startsWith('/api/'),
            handler: 'NetworkOnly',
            options: { cacheName: 'api-never-cache' },
          },
        ],
      },
      devOptions: {
        enabled: false,
      },
    }),
  ],
});
