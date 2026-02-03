<script lang="ts">
  import { page } from '$app/stores';
  import { onDestroy } from 'svelte';
  import { sessionsAPI } from '$lib/api';
  import SessionDetail from '$lib/components/SessionDetail.svelte';
  import type { Session, SessionMessage } from '$lib/types';

  let session = $state<Session | null>(null);
  let messages = $state<SessionMessage[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let activeSessionId = $state<string | null>(null);

  async function loadSession(
    sessionId: string,
    options: { showLoading?: boolean; showError?: boolean } = {}
  ) {
    const showLoading = options.showLoading ?? true;
    const showError = options.showError ?? showLoading;

    if (showLoading) {
      loading = true;
      error = null;
    } else if (showError) {
      error = null;
    }

    try {
      const data = await sessionsAPI.get(sessionId);
      session = data.session;
      messages = data.messages || [];
      activeSessionId = sessionId;
    } catch (err) {
      if (showError) {
        error = err instanceof Error ? err.message : 'Failed to load session';
      }
    } finally {
      if (showLoading) {
        loading = false;
      }
    }
  }

  $effect(() => {
    const sessionId = $page.params.id;
    if (!sessionId) {
      loading = false;
      error = 'Missing session id';
      return;
    }

    if (activeSessionId !== sessionId) {
      loadSession(sessionId);
    }
  });

  onDestroy(() => {
    activeSessionId = null;
  });
</script>

<svelte:head>
  <title>{session?.title || 'Session'} - Counterspell</title>
</svelte:head>

<div class="min-h-screen bg-background flex flex-col">
  <div class="flex-1 overflow-hidden">
    {#if loading}
      <div class="p-4 text-sm text-gray-500">Loading session...</div>
    {:else if error}
      <div class="flex items-center justify-center h-full">
        <div class="text-center">
          <p class="text-base text-red-400 mb-2">{error}</p>
          <button
            onclick={() => loadSession($page.params.id)}
            class="px-4 py-2 bg-violet-500/20 border border-violet-500/30 rounded-lg text-sm text-violet-300 hover:bg-violet-500/30 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    {:else if session}
      <SessionDetail {session} {messages} onRefresh={() => loadSession(session.id, { showLoading: false })} />
    {/if}
  </div>
</div>
