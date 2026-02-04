<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { goto } from '$app/navigation';
  import { sessionsAPI } from '$lib/api';
  import { appState } from '$lib/stores/app.svelte';
  import SessionSkeleton from './SessionSkeleton.svelte';
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
      class="flex w-full max-w-[520px] flex-col gap-4 sm:flex-row sm:items-center sm:justify-between px-2"
    >
      <div class="flex items-center gap-1.5 overflow-x-auto pb-1 scrollbar-hide">
        {#each ['all', 'native', 'claude-code', 'codex'] as option}
          <button
            type="button"
            onclick={() => (filter = option as any)}
            class={cn(
              'shrink-0 whitespace-nowrap px-3 py-1.5 rounded-full text-[11px] font-semibold uppercase tracking-wider transition-all duration-200',
              filter === option
                ? 'bg-white text-black'
                : 'text-gray-500 hover:text-gray-300 hover:bg-white/5'
            )}
          >
            {option === 'all' ? 'All' : option}
          </button>
        {/each}
      </div>
      <button
        type="button"
        onclick={createSession}
        class="group flex items-center justify-center gap-2 px-4 py-2 rounded-full text-[11px] font-semibold uppercase tracking-wider bg-white/5 border border-white/10 text-gray-300 hover:text-white hover:border-white/20 hover:bg-white/10 transition-all duration-200"
      >
        <span class="opacity-70 group-hover:opacity-100">+</span>
        New Session
      </button>
    </div>
  </div>

  {#if loading}
    <div class="flex flex-col gap-4">
      <div class="flex justify-center py-4">
        <div class="flex flex-col items-center gap-3">
          <div
            class="w-10 h-10 rounded-xl bg-violet-500/10 border border-violet-500/20 flex items-center justify-center animate-pulse-glow"
          >
            <i class="fas fa-comments text-lg text-violet-400"></i>
          </div>
          <p class="text-[11px] text-gray-500 font-medium uppercase tracking-wider">
            Loading Sessions
          </p>
        </div>
      </div>
      <SessionSkeleton />
    </div>
  {:else if error}
    <div class="px-4 py-3 rounded-lg bg-red-500/10 border border-red-500/20 text-sm text-red-400">
      {error}
    </div>
  {:else}
    <div class="flex flex-col min-w-0">
      {#if filteredSessions.length === 0}
        <div class="text-center py-12">
          <div class="text-sm text-gray-500">No sessions yet.</div>
        </div>
      {:else}
        <div class="divide-y divide-white/[0.04] border-t border-white/[0.04]">
          {#each filteredSessions as session}
            <button
              type="button"
              onclick={() => openSession(session.id)}
              class="w-full text-left py-5 px-4 sm:px-6 transition-all duration-200 hover:bg-white/[0.02] group"
            >
              <div class="flex items-start justify-between gap-4">
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2.5 mb-1.5">
                    <span
                      class="text-[15px] font-semibold text-gray-100 group-hover:text-white transition-colors truncate"
                    >
                      {session.title || 'Untitled session'}
                    </span>
                    <div
                      class={cn(
                        'flex items-center gap-1 px-2 py-0.5 rounded-full text-[9px] font-bold uppercase tracking-wider border transition-colors',
                        session.agent_backend === 'codex'
                          ? 'bg-purple-500/10 border-purpe-500/20 text-purple-400 group-hover:bg-purple-500/20'
                          : session.agent_backend === 'claude-code'
                            ? 'bg-purple-500/10 border-purple-500/20 text-purple-400 group-hover:bg-purple-500/20'
                            : 'bg-purple-500/10 border-purple-500/20 text-purple-400 group-hover:bg-purple-500/20'
                      )}
                    >
                      {#if session.agent_backend === 'codex'}
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          width="10"
                          height="10"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2.5"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          ><path d="m18 16 4-4-4-4" /><path d="m6 8-4 4 4 4" /><path
                            d="m14.5 4-5 16"
                          /></svg
                        >
                      {:else if session.agent_backend === 'claude-code'}
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          width="10"
                          height="10"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2.5"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          ><path d="M12 2v8" /><path d="m4.93 4.93 5.66 5.66" /><path
                            d="M2 12h8"
                          /><path d="m4.93 19.07 5.66-5.66" /><path d="M12 22v-8" /><path
                            d="m19.07 19.07-5.66-5.66"
                          /><path d="M22 12h-8" /><path d="m19.07 4.93-5.66 5.66" /></svg
                        >
                      {:else}
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          width="10"
                          height="10"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2.5"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          ><path d="M12 8V4H8" /><rect
                            width="16"
                            height="12"
                            x="4"
                            y="8"
                            rx="2"
                          /><path d="M2 14h2" /><path d="M20 14h2" /><path d="M15 13v2" /><path
                            d="M9 13v2"
                          /></svg
                        >
                      {/if}
                      {session.agent_backend === 'native'
                        ? 'Native'
                        : session.agent_backend === 'claude-code'
                          ? 'Claude'
                          : session.agent_backend === 'codex'
                            ? 'Codex'
                            : session.agent_backend}
                    </div>
                  </div>
                  <div
                    class="flex items-center gap-2 text-sm text-gray-500 group-hover:text-gray-400 transition-colors"
                  >
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      width="12"
                      height="12"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      class="opacity-50"
                      ><path
                        d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"
                      /></svg
                    >
                    <span
                      >{session.message_count}
                      {session.message_count === 1 ? 'message' : 'messages'}</span
                    >
                  </div>
                </div>
                <div class="flex flex-col items-end shrink-0 pt-1">
                  <span
                    class="text-[11px] font-medium text-gray-500 group-hover:text-gray-400 tabular-nums"
                  >
                    {formatRelativeTimestamp(session.last_message_at)}
                  </span>
                </div>
              </div>
            </button>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>
