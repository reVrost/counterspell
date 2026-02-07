<script lang="ts">
  import { page } from '$app/stores';
  import { onDestroy } from 'svelte';
  import { sessionsAPI } from '$lib/api';
  import { createSessionSSE } from '$lib/utils/sse';
  import SessionDetail from '$lib/components/SessionDetail.svelte';
  import SessionDetailSkeleton from '$lib/components/SessionDetailSkeleton.svelte';
  import type { AgentStreamEvent, ContentBlock, Session, SessionMessage } from '$lib/types';

  let session = $state<Session | null>(null);
  let messages = $state<SessionMessage[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let activeSessionId = $state<string | null>(null);
  let eventSource: EventSource | null = null;
  let sseSessionId: string | null = null;

  type StreamBlockState = {
    index: number;
    blockType: ContentBlock['type'];
    args: string;
    role: string;
    block?: ContentBlock;
  };

  const streamRoles = new Map<string, string>();
  const streamState = new Map<string, StreamBlockState>();
  const streamSeq = new Map<string, number>();

  function resetStreamState() {
    streamRoles.clear();
    streamState.clear();
    streamSeq.clear();
  }

  function nextStreamSeq(messageId: string): number {
    const next = (streamSeq.get(messageId) ?? 0) + 1;
    streamSeq.set(messageId, next);
    return next;
  }

  function normalizeRole(role?: string): string {
    return role || 'assistant';
  }

  function sessionFieldsFromBlock(
    role: string,
    block: ContentBlock,
    fallbackContent: string
  ): {
    role: string;
    kind: string;
    content: string;
    toolName: string | null;
    toolCallId: string | null;
  } {
    switch (block.type) {
      case 'tool_result':
        return {
          role: 'tool',
          kind: 'tool_result',
          content: block.content || fallbackContent || '',
          toolName: null,
          toolCallId: block.tool_use_id || null,
        };
      case 'tool_use':
        return {
          role: normalizeRole(role),
          kind: 'tool_use',
          content: JSON.stringify(block.input ?? {}),
          toolName: block.name || null,
          toolCallId: block.id || null,
        };
      case 'thinking':
        return {
          role: normalizeRole(role),
          kind: 'thinking',
          content: block.text || fallbackContent || '',
          toolName: null,
          toolCallId: null,
        };
      case 'text':
      default:
        return {
          role: normalizeRole(role),
          kind: 'text',
          content: block.text || fallbackContent || '',
          toolName: null,
          toolCallId: null,
        };
    }
  }

  function sessionRawFromBlock(block: ContentBlock, content: string): string {
    const raw: Record<string, unknown> = { type: block.type };
    switch (block.type) {
      case 'tool_use':
        raw.tool = block.name || '';
        raw.input = block.input ?? {};
        raw.id = block.id || '';
        break;
      case 'tool_result':
        raw.tool_use_id = block.tool_use_id || '';
        raw.content = content;
        break;
      default:
        raw.content = content;
        break;
    }
    return JSON.stringify(raw);
  }

  function startStreamBlock(
    messageId: string,
    role: string,
    blockType: ContentBlock['type'],
    block?: ContentBlock
  ) {
    const seq = nextStreamSeq(messageId);
    const now = Date.now();
    const kind = block?.type || blockType || 'text';
    const lastSequence = messages.at(-1)?.sequence ?? 0;
    const sessionId = session?.id || activeSessionId || '';
    const msg: SessionMessage = {
      id: `stream:${messageId}:${seq}`,
      session_id: sessionId,
      sequence: lastSequence + 1,
      role: normalizeRole(role),
      kind,
      content: '',
      tool_name: block?.name || null,
      tool_call_id: block?.id || block?.tool_use_id || null,
      raw_json: JSON.stringify({ type: kind }),
      created_at: now,
    };
    messages = [...messages, msg];
    streamState.set(messageId, {
      index: messages.length - 1,
      blockType: kind,
      args: '',
      role: normalizeRole(role),
      block,
    });
  }

  function applyStreamEvent(event: AgentStreamEvent) {
    if (!event.message_id || !event.type) return;
    const messageId = event.message_id;
    switch (event.type) {
      case 'message_start': {
        if (event.role) {
          streamRoles.set(messageId, event.role);
        }
        break;
      }
      case 'content_start': {
        const role = event.role || streamRoles.get(messageId) || 'assistant';
        const blockType = (event.block_type || event.block?.type || 'text') as ContentBlock['type'];
        const block = event.block || { type: blockType };
        startStreamBlock(messageId, role, blockType, block);
        break;
      }
      case 'content_delta': {
        const state = streamState.get(messageId);
        if (!state) return;
        const msg = messages[state.index];
        if (!msg) return;
        if (state.blockType === 'text' || state.blockType === 'thinking') {
          msg.content = (msg.content || '') + (event.delta || '');
        } else if (state.blockType === 'tool_use') {
          state.args = (state.args || '') + (event.delta || '');
          msg.content = state.args;
        } else {
          msg.content = (msg.content || '') + (event.delta || '');
        }
        msg.raw_json = JSON.stringify({ type: state.blockType, content: msg.content });
        messages = [...messages];
        break;
      }
      case 'content_end': {
        const state = streamState.get(messageId);
        if (!state) return;
        const msg = messages[state.index];
        if (!msg) return;
        let block = event.block || state.block || { type: state.blockType };
        if (block.type === 'tool_use' && !block.input && state.args) {
          try {
            block = { ...block, input: JSON.parse(state.args) };
          } catch {
            block = { ...block, input: { raw: state.args } };
          }
        }
        if (block.type === 'tool_result' && !block.content) {
          block = { ...block, content: msg.content || '' };
        }
        if ((block.type === 'text' || block.type === 'thinking') && !block.text) {
          block = { ...block, text: msg.content || '' };
        }
        const fields = sessionFieldsFromBlock(state.role, block, msg.content || '');
        msg.role = fields.role;
        msg.kind = fields.kind;
        msg.content = fields.content;
        msg.tool_name = fields.toolName;
        msg.tool_call_id = fields.toolCallId;
        msg.raw_json = sessionRawFromBlock(block, fields.content);
        messages = [...messages];
        streamState.delete(messageId);
        break;
      }
      case 'message_end': {
        streamState.delete(messageId);
        streamRoles.delete(messageId);
        break;
      }
      default:
        break;
    }
  }

  function setupSSE(sessionId: string) {
    if (eventSource && sseSessionId === sessionId) {
      return;
    }

    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }

    sseSessionId = sessionId;
    eventSource = createSessionSSE(sessionId, {
      onAgentUpdate: (data: string) => {
        try {
          const envelope = JSON.parse(data);
          if (!envelope?.data) return;
          const streamEvent = JSON.parse(envelope.data) as AgentStreamEvent;
          applyStreamEvent(streamEvent);
        } catch (e) {
          console.error('Failed to parse session agent update JSON:', e);
        }
      },
      onError: (err) => {
        console.error('Session SSE error:', err);
      },
    });
  }

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
      if (activeSessionId !== sessionId) {
        resetStreamState();
      }
      const data = await sessionsAPI.get(sessionId);
      session = data.session;
      messages = data.messages || [];
      activeSessionId = sessionId;
      setupSSE(sessionId);
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
    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }
    sseSessionId = null;
    resetStreamState();
  });
</script>

<svelte:head>
  <title>{session?.title || 'Session'} - Counterspell</title>
</svelte:head>

<div class="min-h-[100dvh] bg-background flex flex-col">
  <div class="flex-1 overflow-hidden">
    {#if loading}
      <SessionDetailSkeleton />
    {:else if error}
      <div class="flex items-center justify-center h-full">
        <div class="text-center">
          <p class="text-base text-red-400 mb-2">{error}</p>
          <button
            onclick={() => $page.params.id && loadSession($page.params.id)}
            class="px-4 py-2 bg-violet-500/20 border border-violet-500/30 rounded-lg text-sm text-violet-300 hover:bg-violet-500/30 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    {:else if session}
      <SessionDetail
        {session}
        {messages}
        onRefresh={() => loadSession(activeSessionId || '', { showLoading: false })}
      />
    {/if}
  </div>
</div>
