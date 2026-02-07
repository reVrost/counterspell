<script lang="ts">
  import { authAPI } from '$lib/api';
  import GithubIcon from '@lucide/svelte/icons/github';
  import KeyIcon from '@lucide/svelte/icons/key';
  import XIcon from '@lucide/svelte/icons/x';
  import { browser } from '$app/environment';
  import Logo from '$lib/components/Logo.svelte';

  let loading = $state(false);
  let errorMsg = $state('');
  let showError = $state(false);
  let checkingAuth = $state(true);

  async function handleLogin() {
    loading = true;
    showError = false;
    try {
      await authAPI.loginWithInvoker();
    } catch (err) {
      console.error('Login failed:', err);
      loading = false;
      errorMsg = 'Failed to initiate login';
      showError = true;
    }
  }

  async function clearAllCookies() {
    if (!browser) return;

    // Clear all cookies
    document.cookie.split(';').forEach((c) => {
      const domain = window.location.hostname;
      const domains = [domain, `.${domain}`, 'localhost'];
      domains.forEach((d) => {
        document.cookie = c
          .replace(/^ +/, '')
          .replace(/=.*/, `=;expires=${new Date(0).toUTCString()};path=/;domain=${d}`);
        document.cookie = c
          .replace(/^ +/, '')
          .replace(/=.*/, `=;expires=${new Date(0).toUTCString()};path=/;`);
      });
    });

    console.log('✅ All cookies cleared');
  }

  function dismissError() {
    showError = false;
    errorMsg = '';
  }

  $effect(() => {
    if (!browser) return;

    (async () => {
      console.log('📍 Landing page mounted, checking auth...');

      // Check for OAuth errors in URL
      const urlParams = new URLSearchParams(window.location.search);
      const error = urlParams.get('error');
      const errorDesc = urlParams.get('error_description');

      if (error) {
        errorMsg = errorDesc || `Login error: ${error}`;
        showError = true;
        // Clear error from URL
        window.history.replaceState({}, '', '/');
        checkingAuth = false;
        return;
      }

      //Check if already authenticated
      try {
        const session = await authAPI.checkSession();
        console.log('✅ Auth check result:', session);

        if (session.authenticated) {
          console.log('🚀 Redirecting to dashboard...');
          window.location.href = '/dashboard';
        } else {
          console.log('❓ Not authenticated, staying on landing page');
          if (session.authErrorCode === 'OWNER_MISMATCH') {
            errorMsg =
              session.authErrorMessage ||
              'Wrong account for this machine. Please sign in again with the original owner account.';
            showError = true;
          }
        }
      } catch (e) {
        console.log('❌ Auth check failed:', e);
        // Check if it's a 401 error (token expired)
        if (e instanceof Error && e.message.includes('401')) {
          console.log('⚠️ Token expired, clearing cookies...');
          await clearAllCookies();
        }
        // Not authenticated, stay on landing page
      } finally {
        checkingAuth = false;
      }
    })();
  });
</script>

<svelte:head>
  {#if browser}
    <script>
    </script>
  {/if}
</svelte:head>

<div class="h-[100dvh] flex flex-col overflow-hidden bg-[#0C0E12]">
  <!-- Background Effects -->
  <div class="absolute inset-0 overflow-hidden pointer-events-none">
    <div
      class="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-[100px] animate-pulse"
    ></div>
    <div
      class="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-[100px] animate-pulse"
      style="animation-delay: 2s;"
    ></div>
  </div>

  <!-- Error Toast -->
  {#if showError}
    <div
      class="fixed top-4 left-1/2 -translate-x-1/2 z-[200] bg-red-500/90 backdrop-blur text-white px-4 py-3 rounded-lg shadow-lg flex items-center gap-3 animate-in fade-in slide-in-from-top-4 duration-300"
    >
      <XIcon class="w-4 h-4 cursor-pointer hover:opacity-80" onclick={dismissError} />
      <span class="text-sm font-medium">{errorMsg}</span>
    </div>
  {/if}

  <!-- Landing Content -->
  <div
    class="fixed inset-0 z-[100] bg-[#0C0E12] flex flex-col items-center justify-center text-center px-6"
  >
    <!-- Content -->
    <div class="relative z-10 max-w-md w-full space-y-8">
      {#if checkingAuth}
        <div class="flex flex-col items-center gap-6">
          <div
            class="relative w-20 h-20 rounded-3xl bg-violet-500/10 border border-violet-500/20 flex items-center justify-center animate-pulse-glow"
          >
            <i class="fas fa-ghost text-3xl text-violet-400"></i>
            <div
              class="absolute inset-0 border-2 border-violet-500/30 rounded-3xl animate-ping opacity-20"
            ></div>
          </div>
          <div class="space-y-2">
            <h3 class="text-sm font-semibold text-gray-200">Aligning Realities</h3>
            <p class="text-[11px] text-gray-500 font-medium uppercase tracking-widest">
              Checking Authentication
            </p>
          </div>
        </div>
      {:else}
        <div class="space-y-4">
          <div class="mb-6 flex justify-center">
            <Logo class="w-16 h-16" />
          </div>
          <h1 class="text-3xl font-bold text-white tracking-tight">Welcome to Counterspell</h1>
          <p class="text-gray-400 text-sm leading-relaxed">
            Mobile-first, hosted AI agent Kanban.
            <br />
            Orchestrate from your pocket.
          </p>
        </div>

        <!-- Login Button -->
        {#if !loading}
          <div>
            <button
              onclick={handleLogin}
              class="w-full bg-white text-black font-bold h-12 rounded-lg hover:bg-gray-200 transition active:scale-95 flex items-center justify-center gap-2"
            >
              <KeyIcon class="w-5 h-5" />
              Continue with SSO
            </button>
          </div>
        {:else}
          <!-- Loading State -->
          <div class="space-y-4">
            <div
              class="bg-gray-900/50 rounded-xl p-4 border border-gray-800 text-left space-y-3 font-mono text-sm"
            >
              <div class="flex items-center gap-3">
                <div
                  class="w-4 h-4 rounded-full flex items-center justify-center bg-purple-500/20 text-purple-400"
                >
                  <i class="fas fa-circle-notch fa-spin"></i>
                </div>
                <span class="text-gray-200">Redirecting to Counterspell...</span>
              </div>
            </div>
          </div>
        {/if}
      {/if}
    </div>

    <!-- Footer -->
    <div class="absolute bottom-8 text-center space-y-2">
      <p class="text-sm text-gray-600">
        <a
          href="https://github.com/revrost/counterspell"
          target="_blank"
          class="hover:text-gray-400 transition"
        >
          <GithubIcon class="w-3 h-3 inline mr-1" />
          Open Source
        </a>
      </p>
    </div>
  </div>
</div>

<style>
  :global(body) {
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }
  :global(button, a) {
    -webkit-tap-highlight-color: transparent;
    touch-action: manipulation;
  }
</style>
