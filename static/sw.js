const CACHE_NAME = 'counterspell-v1';
const ASSETS_TO_CACHE = [
  '/',
  '/static/icon-192.svg',
  '/static/icon-512.svg',
  'https://unpkg.com/htmx.org@2.0.4',
  'https://unpkg.com/htmx-ext-sse@2.2.2/sse.js',
  'https://cdn.jsdelivr.net/npm/alpinejs@3.14.8/dist/cdn.min.js',
  'https://cdn.tailwindcss.com',
  'https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.7.2/css/all.min.css'
];

// Install event - cache assets
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(ASSETS_TO_CACHE);
    })
  );
  self.skipWaiting();
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== CACHE_NAME) {
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
  self.clients.claim();
});

// Fetch event - serve from cache, fallback to network
self.addEventListener('fetch', (event) => {
  // Skip SSE and API requests
  if (event.request.url.includes('/events') ||
      event.request.url.includes('/stream') ||
      event.request.url.includes('/api/')) {
    return;
  }

  event.respondWith(
    caches.match(event.request).then((cachedResponse) => {
      if (cachedResponse) {
        // For HTML requests, try network first to get fresh content
        if (event.request.headers.get('accept')?.includes('text/html')) {
          return fetch(event.request).then((networkResponse) => {
            // Update cache
            caches.open(CACHE_NAME).then((cache) => {
              cache.put(event.request, networkResponse.clone());
            });
            return networkResponse;
          }).catch(() => cachedResponse);
        }
        return cachedResponse;
      }

      return fetch(event.request).then((networkResponse) => {
        // Only cache successful responses
        if (networkResponse.ok && !event.request.url.includes('/feed')) {
          caches.open(CACHE_NAME).then((cache) => {
            cache.put(event.request, networkResponse.clone());
          });
        }
        return networkResponse;
      });
    })
  );
});
