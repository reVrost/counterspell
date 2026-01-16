const CACHE_NAME = 'counterspell-v2';

// App Shell: The minimal UI frame that loads instantly
const APP_SHELL = [
  '/',
  '/static/app.js',
  '/static/icon-192.png',
  '/static/icon-512.png'
];

// Static assets that rarely change
const STATIC_ASSETS = [
  'https://unpkg.com/htmx.org@2.0.4',
  'https://unpkg.com/htmx-ext-sse@2.2.2/sse.js',
  'https://unpkg.com/idiomorph@0.7.4/dist/idiomorph-ext.min.js',
  'https://cdn.jsdelivr.net/npm/@alpinejs/collapse@3.14.8/dist/cdn.min.js',
  'https://cdn.jsdelivr.net/npm/alpinejs@3.14.8/dist/cdn.min.js',
  'https://cdn.tailwindcss.com',
  'https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.7.2/css/all.min.css',
  'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap'
];

const ASSETS_TO_CACHE = [...APP_SHELL, ...STATIC_ASSETS];

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

// Fetch event - App Shell architecture with stale-while-revalidate for dynamic content
self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);
  
  // Skip SSE, API, and POST requests entirely
  if (event.request.method !== 'GET' ||
      url.pathname.includes('/events') ||
      url.pathname.includes('/stream') ||
      url.pathname.startsWith('/api/')) {
    return;
  }

  // HTML navigation requests: Network-first with App Shell fallback
  if (event.request.mode === 'navigate' || 
      event.request.headers.get('accept')?.includes('text/html')) {
    event.respondWith(
      fetch(event.request)
        .then((networkResponse) => {
          // Cache the fresh HTML shell
          if (networkResponse.ok) {
            const clone = networkResponse.clone();
            caches.open(CACHE_NAME).then((cache) => cache.put('/', clone));
          }
          return networkResponse;
        })
        .catch(() => {
          // Offline: serve cached App Shell, HTMX will fill content when back online
          return caches.match('/') || new Response(
            '<html><body style="background:#0C0E12;color:#fff;display:flex;align-items:center;justify-content:center;height:100vh;font-family:sans-serif"><div>Offline - Please reconnect</div></body></html>',
            { headers: { 'Content-Type': 'text/html' } }
          );
        })
    );
    return;
  }

  // Static assets: Cache-first with background revalidation
  event.respondWith(
    caches.match(event.request).then((cachedResponse) => {
      const fetchPromise = fetch(event.request).then((networkResponse) => {
        if (networkResponse.ok) {
          caches.open(CACHE_NAME).then((cache) => {
            cache.put(event.request, networkResponse.clone());
          });
        }
        return networkResponse;
      }).catch(() => cachedResponse);

      // Return cached immediately, update in background
      return cachedResponse || fetchPromise;
    })
  );
});
