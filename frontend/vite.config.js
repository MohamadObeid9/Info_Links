import { defineConfig } from 'vite';
import { VitePWA } from 'vite-plugin-pwa';

// Cloudflare Rocket Loader rewrites type="module" to a fake type and then
// never runs the bundle (CSP also blocks its eval path). The page chrome
// still loads, but #coursesOutput stays empty. data-cfasync="false" opts
// scripts out; turning Rocket Loader off in the dashboard is the live fix.
function disableCloudflareRocketLoader() {
  return {
    name: 'disable-cloudflare-rocket-loader',
    enforce: 'post',
    transformIndexHtml: {
      order: 'post',
      handler(html) {
        return html.replace(
          /<script(?![^>]*\bdata-cfasync=)/gi,
          '<script data-cfasync="false"',
        );
      },
    },
  };
}

export default defineConfig({
  root: './',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    rollupOptions: {
      input: './index.html',
    },
  },
  test: {
    environment: 'happy-dom',
    include: ['js/**/*.test.js'],
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/openapi.json': 'http://localhost:8080',
      '/auth.md': 'http://localhost:8080',
      '/.well-known': 'http://localhost:8080',
      '/robots.txt': 'http://localhost:8080',
      '/sitemap.xml': 'http://localhost:8080',
      '/courses': 'http://localhost:8080',
      '/course': 'http://localhost:8080',
      '/program': 'http://localhost:8080',
    },
  },
  plugins: [
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: [
        'favicon.ico',
        'assets/favicon.ico',
        'assets/favicon-16x16.png',
        'assets/favicon-32x32.png',
        'assets/apple-touch-icon.png',
        'assets/android-chrome-192x192.png',
        'assets/android-chrome-512x512.png',
      ],
      manifest: {
        name: 'Info Links',
        short_name: 'InfoLinks',
        description: 'All courses, materials & links organized.',
        theme_color: '#0f172a',
        background_color: '#0f0f13',
        icons: [
          {
            src: '/assets/android-chrome-192x192.png',
            sizes: '192x192',
            type: 'image/png',
          },
          {
            src: '/assets/android-chrome-512x512.png',
            sizes: '512x512',
            type: 'image/png',
          },
        ],
      },
    }),
    disableCloudflareRocketLoader(),
  ],
});
