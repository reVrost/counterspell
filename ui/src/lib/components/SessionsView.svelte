<script lang="ts">
  import { onMount } from 'svelte';
  import { sessionsAPI } from '$lib/api';
  import { appState } from '$lib/stores/app.svelte';
  import ChatInput from '$lib/components/ChatInput.svelte';
  import { cn } from '$lib/utils';
  import type { Session, SessionMessage } from '$lib/types';

  let sessions = $state<Session[]>([]);
  let selectedSession = $state<Session | null>(null);
  let messages = $state<SessionMessage[]>([]);
  let loading = $state(true);
  let detailLoading = $state(false);
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

  function isSystemLikeMessage(message: SessionMessage): boolean {
    const role = message.role?.toLowerCase?.() ?? '';
    if (role === 'system' || role === 'developer') return true;
    return message.kind?.toLowerCase?.() === 'system';
  }

  const visibleMessages = $derived.by(() => messages.filter((msg) => !isSystemLikeMessage(msg)));

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
    try {
      detailLoading = true;
      const data = await sessionsAPI.get(sessionId);
      selectedSession = data.session;
      messages = data.messages || [];
    } catch (err) {
      appState.showToast(err instanceof Error ? err.message : 'Failed to load session', 'error');
    } finally {
      detailLoading = false;
    }
  }

  async function promoteSession() {
    if (!selectedSession) return;
    try {
      const res = await sessionsAPI.promote(selectedSession.id);
      appState.showToast('Promoted to task', 'success');
      if (res.task_id) {
        appState.openModal(res.task_id);
      }
    } catch (err) {
      appState.showToast(err instanceof Error ? err.message : 'Failed to promote session', 'error');
    }
  }

  async function createSession() {
    try {
      const backend = appState.settings?.agentBackend || 'native';
      const session = await sessionsAPI.create(backend);
      await loadSessions();
      await openSession(session.id);
    } catch (err) {
      appState.showToast(err instanceof Error ? err.message : 'Failed to create session', 'error');
    }
  }

  async function sendMessage(message: string, modelId: string) {
    if (!selectedSession) return;
    try {
      await sessionsAPI.chat(selectedSession.id, message, modelId);
      await openSession(selectedSession.id);
      await loadSessions();
    } catch (err) {
      appState.showToast(err instanceof Error ? err.message : 'Failed to send message', 'error');
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
    <div class="grid min-w-0 gap-4 md:grid-cols-2">
      <div class={cn('space-y-2 min-w-0', selectedSession ? 'hidden md:block' : '')}>
        {#if filteredSessions.length === 0}
          <div class="text-sm text-gray-500">No sessions yet.</div>
        {:else}
          {#each filteredSessions as session}
            <button
              type="button"
              onclick={() => openSession(session.id)}
              class={cn(
                'w-full text-left rounded-xl border p-4 md:p-3 transition',
                selectedSession?.id === session.id
                  ? 'border-white/30 bg-white/5'
                  : 'border-white/10 hover:border-white/20'
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

      <div
        class={cn(
          'relative min-w-0 rounded-xl border border-white/10 bg-white/[0.02] min-h-[50vh] md:min-h-[420px] max-h-[calc(100vh-240px)] flex flex-col overflow-hidden',
          selectedSession ? 'flex' : 'hidden md:flex'
        )}
      >
        {#if !selectedSession}
          <div class="p-4 text-sm text-gray-500">Select a session to view.</div>
        {:else if detailLoading}
          <div class="p-4 text-sm text-gray-500">Loading session...</div>
        {:else}
          <div
            class="flex items-center justify-between gap-3 border-b border-white/10 bg-white/[0.02] px-3 py-3"
          >
            <div class="flex min-w-0 items-center gap-2">
              <button
                type="button"
                onclick={() => (selectedSession = null)}
                class="px-3 py-1.5 rounded-full text-[11px] font-medium uppercase tracking-wide border border-white/10 text-gray-300 hover:text-white hover:border-white/20 transition md:hidden"
              >
                Back
              </button>
              <div class="min-w-0">
                <div class="text-sm font-semibold text-gray-100 truncate">
                  {selectedSession.title || 'Untitled session'}
                </div>
                <div class="text-xs font-medium text-gray-500 truncate">
                  {selectedSession.agent_backend} • {formatRelativeTimestamp(selectedSession.last_message_at)}
                </div>
              </div>
            </div>
            <button
              type="button"
              onclick={promoteSession}
              class="shrink-0 px-3 py-2 rounded-full text-xs font-medium uppercase tracking-wide border border-white/10 text-gray-300 hover:text-white hover:border-white/20 transition"
            >
              Promote
            </button>
          </div>

          <div class="flex-1 overflow-y-auto relative">
            <div class="space-y-2 px-3 py-3 pb-28">
              {#if visibleMessages.length === 0}
                <div class="text-xs font-medium text-gray-500">No messages yet.</div>
              {:else}
                {#each visibleMessages as msg}
                  <div
                    class={cn(
                      'rounded-lg border px-3 py-2 text-xs font-medium',
                      msg.role === 'user'
                        ? 'border-violet-500/30 bg-violet-500/10 text-gray-200'
                        : 'border-white/10 bg-white/5 text-gray-300'
                    )}
                  >
                    <div
                      class="flex items-center justify-between text-[10px] uppercase text-gray-500 mb-1"
                    >
                      <span>{msg.role}</span>
                      <span>{msg.kind}</span>
                    </div>
                    {#if msg.kind === 'tool_use'}
                      <div class="text-[10px] uppercase text-blue-300 mb-1">
                        {msg.tool_name || 'tool'}
                      </div>
                      <pre class="whitespace-pre-wrap break-words text-[11px] text-gray-400">{msg.content ||
                          ''}</pre>
                    {:else if msg.kind === 'tool_result'}
                      <pre class="whitespace-pre-wrap break-words text-[11px] text-gray-400">{msg.content ||
                          ''}</pre>
                    {:else}
                      <div class="text-[12px] whitespace-pre-wrap break-words">
                        {msg.content || ''}
                      </div>
                    {/if}
                  </div>
                {/each}
              {/if}
            </div>
          </div>

          <div class="absolute bottom-0 inset-x-0 z-10 pb-4 px-3">
            <div
              class="absolute inset-0 bg-gradient-to-t from-[#0D1117] via-[#0D1117]/95 to-transparent pointer-events-none"
            ></div>
            <div class="relative mx-auto max-w-3xl">
              <ChatInput
                mode="chat"
                placeholder="Message this session..."
                onSubmit={sendMessage}
              />
            </div>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>
