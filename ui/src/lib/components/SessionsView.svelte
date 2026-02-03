<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { goto } from '$app/navigation';
  import { sessionsAPI } from '$lib/api';
  import { appState } from '$lib/stores/app.svelte';
  import { cn } from '$lib/utils';
  import type { Session } from '$lib/types';

  let sessions = $state<Session[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let filter = $state<'all' | 'native' | 'claude-code' | 'codex'>('all');

  const filteredSessions = $derived.by(() => {
    if (filter === 'all') return sessions;
    return sessions.filter((s) => s.agent_backend === filter);
  });

  function formatRelativeTimestamp(value?: number | null): string {
    if (!value) return 'No messages yet';
    const diffMs = Math.max(0, Date.now() - value);
    const seconds = Math.floor(diffMs / 1000);
    if (seconds < 45) return 'now';
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}m`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h`;
    const days = Math.floor(hours / 24);
    if (days < 7) return `${days}d`;
    const weeks = Math.floor(days / 7);
    if (weeks < 4) return `${weeks}w`;
    const months = Math.floor(days / 30);
    if (months < 12) return `${months}mo`;
    const years = Math.floor(days / 365);
    return `${years}y`;
  }

  async function loadSessions() {
    try {
      loading = true;
      error = null;
      sessions = await sessionsAPI.list();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load sessions';
    } finally {
      loading = false;
    }
  }

  async function openSession(sessionId: string) {
    goto(`/sessions/${sessionId}`);
  }

  async function createSession() {
    try {
      const backend = appState.settings?.agentBackend || 'native';
      const session = await sessionsAPI.create(backend);
      await loadSessions();
      goto(`/sessions/${session.id}`);
    } catch (err) {
      appState.showToast(err instanceof Error ? err.message : 'Failed to create session', 'error');
    }
  }

  onMount(() => {
    loadSessions();
  });
</script>

<div class="flex flex-col gap-6">
  <div class="flex justify-center">
    <div
      class="flex w-full max-w-[520px] flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
    >
      <div class="flex min-w-0 items-center gap-2 overflow-x-auto pb-1 px-1">
        {#each ['all', 'native', 'claude-code', 'codex'] as option}
          <button
            type="button"
            onclick={() => (filter = option as any)}
            class={cn(
              'shrink-0 whitespace-nowrap px-3 py-1.5 rounded-full text-xs font-medium uppercase tracking-wide border transition',
              filter === option
                ? 'bg-white/10 border-white/20 text-white'
                : 'bg-transparent border-white/10 text-gray-500 hover:text-gray-300'
            )}
          >
            {option === 'all' ? 'All' : option}
          </button>
        {/each}
      </div>
      <button
        type="button"
        onclick={createSession}
        class="w-full sm:w-auto px-3 py-2 rounded-full text-xs font-medium uppercase tracking-wide border border-white/10 text-gray-300 hover:text-white hover:border-white/20 transition"
      >
        New Session
      </button>
    </div>
  </div>

  {#if loading}
    <div class="text-sm text-gray-500">Loading sessions...</div>
  {:else if error}
    <div class="text-sm text-red-400">{error}</div>
  {:else}
    <div class="space-y-2 min-w-0">
      {#if filteredSessions.length === 0}
        <div class="text-sm text-gray-500">No sessions yet.</div>
      {:else}
        {#each filteredSessions as session}
          <button
            type="button"
            onclick={() => openSession(session.id)}
            class={cn(
              'w-full text-left rounded-xl border p-4 md:p-3 transition',
              'border-white/10 hover:border-white/20'
            )}
          >
            <div class="flex items-center justify-between gap-2">
              <div class="text-sm font-medium text-gray-200 truncate">
                {session.title || 'Untitled session'}
              </div>
              <span class="text-[10px] uppercase text-gray-500">
                {session.agent_backend}
              </span>
            </div>
            <div class="text-xs font-medium text-gray-500 mt-1">
              {formatRelativeTimestamp(session.last_message_at)}
            </div>
          </button>
        {/each}
      {/if}
    </div>
  {/if}
</div>
