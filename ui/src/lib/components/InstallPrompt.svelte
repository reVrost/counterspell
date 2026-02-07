<script lang="ts">
  import { browser } from '$app/environment';
  import DownloadIcon from '@lucide/svelte/icons/download';
  import XIcon from '@lucide/svelte/icons/x';
  import { fade } from '$lib/utils/transitions';

  let deferredPrompt: Event | null = $state(null);
  let showPrompt = $state(false);
  let isInstalled = $state(false);

  const DISMISS_KEY = 'pwa-install-dismissed';
  const INSTALLED_KEY = 'pwa-installed';

  function handleBeforeInstallPrompt(e: Event) {
    e.preventDefault();
    deferredPrompt = e;
    showPrompt = true;
  }

  async function install() {
    if (!deferredPrompt) return;

    try {
      await (deferredPrompt as any).prompt();
      const { outcome } = await (deferredPrompt as any).userChoice;

      if (outcome === 'accepted') {
        if (browser) {
          localStorage.setItem(INSTALLED_KEY, 'true');
        }
        isInstalled = true;
      }
    } catch (err) {
      console.error('Install error:', err);
    } finally {
      deferredPrompt = null;
      showPrompt = false;
    }
  }

  function dismiss() {
    if (browser) {
      localStorage.setItem(DISMISS_KEY, 'true');
    }
    showPrompt = false;
  }

  if (browser) {
    const installed = localStorage.getItem(INSTALLED_KEY);
    const dismissed = localStorage.getItem(DISMISS_KEY);

    if (!installed) {
      isInstalled = false;
      if (!dismissed) {
        window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt);
      }
    } else {
      isInstalled = true;
    }

    window.addEventListener('appinstalled', () => {
      if (browser) {
        localStorage.setItem(INSTALLED_KEY, 'true');
      }
      isInstalled = true;
      showPrompt = false;
    });
  }
</script>

{#if showPrompt}
  <div class="fixed bottom-0 left-0 right-0 z-50 p-4 sm:p-6 pointer-events-none">
    <div class="relative max-w-md mx-auto pointer-events-auto">
      <div
        transition:fade|global={{ duration: 200 }}
        class="absolute inset-0 -top-2 -left-2 -right-2 -bottom-2 bg-black/60 backdrop-blur-sm rounded-3xl"
      ></div>
      <div
        class="relative bg-gray-950 border border-violet-500/20 rounded-2xl p-4 shadow-2xl shadow-violet-500/10"
      >
        <div class="flex items-start gap-3">
          <div
            class="w-10 h-10 rounded-xl bg-violet-500/10 border border-violet-500/20 flex items-center justify-center flex-shrink-0"
          >
            <DownloadIcon class="w-5 h-5 text-violet-400" />
          </div>

          <div class="flex-1 min-w-0">
            <h3 class="text-sm font-semibold text-white mb-1">Install Counterspell</h3>
            <p class="text-xs text-gray-400">Add to your home screen for the best experience</p>
          </div>

          <button
            onclick={dismiss}
            class="text-gray-500 hover:text-gray-300 transition-colors"
            aria-label="Dismiss"
          >
            <XIcon class="w-4 h-4" />
          </button>
        </div>

        <div class="flex gap-2 mt-4">
          <button
            onclick={dismiss}
            class="flex-1 px-4 py-2 rounded-lg bg-gray-900 text-xs font-medium text-gray-300 border border-gray-800 hover:bg-gray-800 transition-colors"
          >
            Later
          </button>
          <button
            onclick={install}
            class="flex-1 px-4 py-2 rounded-lg bg-violet-500 text-xs font-medium text-white border border-violet-500 hover:bg-violet-400 transition-colors"
          >
            Install
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
