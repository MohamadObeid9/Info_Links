const CACHE_NAME = 'infolinks-v1';
const ASSETS = [
  '/',
  '/index.html',
  '/html/body.html',
  '/app.js',
  '/styles/app.css',
  '/styles/variables.css',
  '/styles/layout.css',
  '/styles/components.css',
  '/styles/responsive.css',
  '/styles/skeleton.css',
  '/styles/admin.css',
  '/js/state.js',
  '/js/supabase.js',
  '/js/cache.js',
  '/js/data.js',
  '/js/home.js',
  '/js/ui.js',
  '/js/views.js',
  '/js/admin.js',
  '/js/feedback.js',
  '/js/modals.js',
  '/js/export.js',
  '/js/skeleton.js'
];

self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE_NAME).then(cache => cache.addAll(ASSETS))
  );
});

self.addEventListener('fetch', event => {
  // Only intercept GET requests for the app shell, bypass Supabase API calls
  if (event.request.method !== 'GET' || event.request.url.includes('supabase.co')) {
    return;
  }
  
  event.respondWith(
    caches.match(event.request).then(response => {
      // Network-first strategy for app shell
      return fetch(event.request).then(fetchRes => {
        return caches.open(CACHE_NAME).then(cache => {
          cache.put(event.request, fetchRes.clone());
          return fetchRes;
        });
      }).catch(() => response); // Fallback to cache if offline
    })
  );
});
