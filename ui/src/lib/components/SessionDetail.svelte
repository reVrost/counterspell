<script lang="ts">
  import { goto } from '$app/navigation';
  import { sessionsAPI } from '$lib/api';
  import { appState } from '$lib/stores/app.svelte';
  import { cn } from '$lib/utils';
  import ChatInput from './ChatInput.svelte';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
  import type { Session, SessionMessage } from '$lib/types';

  interface Props {
    session: Session;
    messages: SessionMessage[];
    onRefresh?: () => Promise<void>;
  }

  let { session, messages, onRefresh }: Props = $props();

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

  function handleBack() {
    goto('/dashboard');
  }

  async function handlePromote() {
    try {
      const res = await sessionsAPI.promote(session.id);
      appState.showToast('Promoted to task', 'success');
      if (res.task_id) {
        goto(`/tasks/${res.task_id}`);
      }
    } catch (err) {
      appState.showToast(err instanceof Error ? err.message : 'Failed to promote session', 'error');
    }
  }

  async function handleChatSubmit(message: string, modelId: string) {
    try {
      await sessionsAPI.chat(session.id, message, modelId);
      await onRefresh?.();
    } catch (err) {
      appState.showToast(err instanceof Error ? err.message : 'Failed to send message', 'error');
    }
  }
</script>

<div class="flex flex-col h-screen">
  <div
    class="px-4 py-2 border-b border-white/5 flex items-center justify-between shrink-0 bg-popover"
  >
    <div class="flex items-center gap-3 min-w-0">
      <button
        onclick={handleBack}
        class="w-11 h-11 rounded-full hover:bg-white/5 flex items-center justify-center text-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500/50"
        aria-label="Go back"
      >
        <ArrowLeftIcon class="w-5 h-5" />
      </button>
      <div class="min-w-0">
        <div class="flex items-center gap-2">
          <span class="text-gray-500 text-[10px] uppercase tracking-wide">
            {session.agent_backend}
          </span>
          <span class="text-[10px] text-gray-600 font-mono">#{session.id}</span>
        </div>
        <h2 class="text-sm font-bold text-[#FFFFFF] line-clamp-1">
          {session.title || 'Untitled session'}
        </h2>
      </div>
    </div>

    <button
      type="button"
      onclick={handlePromote}
      class="shrink-0 px-3 py-2 rounded-full text-xs font-medium uppercase tracking-wide border border-white/10 text-gray-300 hover:text-white hover:border-white/20 transition"
    >
      Promote
    </button>
  </div>

  <div class="flex-1 overflow-hidden flex flex-col relative">
    <div class="px-4 pt-3 text-xs font-medium text-gray-500">
      {formatRelativeTimestamp(session.last_message_at)}
    </div>

    <div class="flex-1 overflow-y-auto relative px-4 pb-28 pt-3">
      {#if visibleMessages.length === 0}
        <div class="text-xs font-medium text-gray-500">No messages yet.</div>
      {:else}
        <div class="space-y-2">
          {#each visibleMessages as msg}
            <div
              class={cn(
                'rounded-lg border px-3 py-2 text-xs font-medium',
                msg.role === 'user'
                  ? 'border-violet-500/30 bg-violet-500/10 text-gray-200'
                  : 'border-white/10 bg-white/5 text-gray-300'
              )}
            >
              <div class="flex items-center justify-between text-[10px] uppercase text-gray-500 mb-1">
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
        </div>
      {/if}
    </div>

    <div class="absolute bottom-0 inset-x-0 z-10 pb-4 px-4">
      <div
        class="absolute inset-0 bg-gradient-to-t from-[#0D1117] via-[#0D1117]/95 to-transparent pointer-events-none"
      ></div>
      <div class="relative mx-auto max-w-3xl">
        <ChatInput
          mode="chat"
          placeholder="Message this session..."
          onSubmit={handleChatSubmit}
        />
      </div>
    </div>
  </div>
</div>
