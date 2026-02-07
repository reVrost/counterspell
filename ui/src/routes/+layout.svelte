<script lang="ts">
  import '../app.css';
  import { QueryClient, QueryClientProvider } from '@tanstack/svelte-query';
  import { appState } from '$lib/stores/app.svelte';
  import { page } from '$app/stores';
  import { browser } from '$app/environment';
  import { authAPI } from '$lib/api';
  import { initGlobalErrorHandlers } from '$lib/utils/logger';
  import InstallPrompt from '$lib/components/InstallPrompt.svelte';

  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 1000 * 60 * 5,
        refetchOnWindowFocus: false,
        retry: false,
      },
    },
  });

  let { children } = $props();
  let isInitialized = $state(false);

  if (browser) {
    (window as any).forceLogout = async () => {
      console.log('🔄 Force logout triggered');
      appState.isAuthenticated = false;
      appState.userEmail = '';
      localStorage.clear();
      sessionStorage.clear();

      document.cookie.split(';').forEach((c) => {
        document.cookie = c
          .replace(/^ +/, '')
          .replace(/=.*/, '=;expires=' + new Date().toUTCString() + ';path=/');
      });

      try {
        await fetch('/auth/logout', {
          method: 'POST',
          credentials: 'include',
        });
      } catch (e) {
        console.error('Logout error:', e);
      }

      window.location.href = '/';
    };
  }

  // Initialize app state
  $effect(() => {
    const init = async () => {
      // Initialize global error handlers first
      initGlobalErrorHandlers();

      console.log('🚀 App layout mounting...');
      await appState.init();
      isInitialized = true;
      console.log('✅ App state initialized, isAuth:', appState.isAuthenticated);
    };
    init();
  });

  // Auth guard - handle authentication
  // With control-plane auth, isAuthenticated = githubConnected
  let hasRedirected = false;
  $effect(() => {
    if (!isInitialized || !browser || hasRedirected) return;

    const path = $page.url.pathname;

    console.log(
      '📍 Navigation:',
      path,
      'auth:',
      appState.isAuthenticated,
      'github:',
      appState.githubConnected
    );

    // Dashboard requires authentication (GitHub connected)
    if (path.startsWith('/app')) {
      if (!appState.isAuthenticated) {
        console.log('🔒 Not authenticated, redirecting to home...');
        hasRedirected = true;
        window.location.href = '/';
        return;
      }
    }

    // Landing page - if authenticated, redirect to dashboard
    if (path === '/') {
      if (appState.isAuthenticated) {
        console.log('✅ Authenticated, redirecting to dashboard...');
        hasRedirected = true;
        window.location.href = '/app';
      }
    }
  });
</script>

<QueryClientProvider client={queryClient}>
  {@render children()}
  {#if isInitialized}
    <InstallPrompt />
  {/if}
</QueryClientProvider>
