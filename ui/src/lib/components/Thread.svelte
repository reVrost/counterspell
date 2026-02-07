<script lang="ts">
  import { cn, getInitial } from '$lib/utils';
  import type { ContentBlock, Message, SessionMessage } from '$lib/types';
  import MarkdownRenderer from './MarkdownRenderer.svelte';
  import ToolBlock from './ToolBlock.svelte';
  import { tick } from 'svelte';
  import { appState } from '$lib/stores/app.svelte';

  interface Props {
    mode: 'task' | 'session';
    messages: Message[] | SessionMessage[];
    emptyText?: string;
    emptyClass?: string;
    class?: string;
    scrollContainerId?: string;
  }

  let {
    mode,
    messages,
    emptyText = 'No messages yet.',
    emptyClass = 'text-xs text-gray-500',
    class: className = '',
    scrollContainerId,
  }: Props = $props();

  const userAvatarUrl = $derived(
    appState.githubLogin ? `https://github.com/${appState.githubLogin}.png` : null
  );
  const userInitial = $derived(getInitial(appState.githubLogin || appState.userEmail));

  let expandedThinking = $state<Set<string>>(new Set());

  function toggleThinking(id: string) {
    const newSet = new Set(expandedThinking);
    if (newSet.has(id)) {
      newSet.delete(id);
    } else {
      newSet.add(id);
    }
    expandedThinking = newSet;
  }

  type ToolItem = { tool: string; call: string; result: string };

  type TaskDisplayItem =
    | { type: 'user'; id: string; message: Message }
    | { type: 'assistant_text'; id: string; content: string }
    | { type: 'thinking'; id: string; content: string }
    | { type: 'tool'; id: string; tool: string; call: string; result: string };

  type SessionDisplayItem =
    | { type: 'message'; id: string; message: SessionMessage }
    | { type: 'tool'; id: string; tool: string; call: string; result: string };

  function formatToolInput(value: unknown): string {
    if (value == null) return '';
    if (typeof value === 'string') return value;
    try {
      return JSON.stringify(value, null, 2);
    } catch {
      return String(value);
    }
  }

  function extractToolFromUnknown(value: unknown): { tool: string; call: string } | null {
    if (!value) return null;
    if (Array.isArray(value)) {
      for (const entry of value) {
        const found = extractToolFromUnknown(entry);
        if (found) return found;
      }
      return null;
    }
    if (typeof value !== 'object') return null;

    const record = value as Record<string, unknown>;
    const tool = record.tool ?? record.tool_name ?? record.name ?? record.toolName;
    if (!tool) return null;

    const call = formatToolInput(
      record.input ??
        record.arguments ??
        record.args ??
        record.content ??
        record.command ??
        record.data
    );
    return { tool: String(tool), call };
  }

  function parseToolFromJson(text: string): { tool: string; call: string } | null {
    if (!text) return null;
    const trimmed = text.trim();
    if (!trimmed || (!trimmed.startsWith('{') && !trimmed.startsWith('['))) return null;
    try {
      const parsed = JSON.parse(trimmed);
      return extractToolFromUnknown(parsed);
    } catch {
      return null;
    }
  }

  function parseToolFromContent(text: string): { tool: string; call: string } {
    const trimmed = text.trim();
    if (!trimmed) return { tool: 'tool', call: '' };

    // 1. Try JSON first
    const fromJson = parseToolFromJson(trimmed);
    if (fromJson) return fromJson;

    const lines = trimmed.split('\n');
    const firstLine = lines[0].trim();

    // 2. Match "toolName: call content" or "toolName call content"
    // Improved regex to handle colons and whitespace more strictly
    const prefixed = firstLine.match(/^([a-zA-Z0-9_-]{2,})[:\s]+(.+)$/);

    if (prefixed) {
      const tool = prefixed[1];
      const firstLineContent = prefixed[2];
      const remainingLines = lines.slice(1);

      // Reconstruct the call: take the rest of the first line + all subsequent lines
      const call = [firstLineContent, ...remainingLines].join('\n').trim();
      return { tool, call };
    }

    // 3. Match "toolName\ncall content" (Tool name on its own line)
    if (lines.length > 1 && /^[a-zA-Z0-9_-]{2,}$/.test(firstLine)) {
      return {
        tool: firstLine,
        call: lines.slice(1).join('\n').trim(),
      };
    }

    // 4. Fallback: treat the whole thing as the call for a generic 'tool'
    return { tool: 'tool', call: trimmed };
  }

  function parseToolMessage(msg: Message): { tool: string; call: string } {
    const fromParts = parseToolFromJson(msg.parts || '');
    if (fromParts) return fromParts;
    return parseToolFromContent(msg.content || '');
  }

  function parseParts(msg: Message): ContentBlock[] {
    if (msg.parts) {
      try {
        const parsed = JSON.parse(msg.parts);
        if (Array.isArray(parsed)) return parsed;
      } catch {
        // ignore parse errors
      }
    }
    if (msg.content) {
      return [{ type: 'text', text: msg.content }];
    }
    return [];
  }

  function concatText(blocks: ContentBlock[]): string {
    return blocks
      .filter((b) => b.type === 'text' && b.text)
      .map((b) => b.text)
      .join('');
  }

  function isSystemLikeMessage(message: SessionMessage): boolean {
    const role = message.role?.toLowerCase?.() ?? '';
    if (role === 'system' || role === 'developer') return true;
    return message.kind?.toLowerCase?.() === 'system';
  }

  const taskItems = $derived.by(() => {
    if (mode !== 'task') return [] as TaskDisplayItem[];
    const taskMessages = messages as Message[];
    const items: TaskDisplayItem[] = [];
    let i = 0;

    while (i < taskMessages.length) {
      const msg = taskMessages[i];

      // User messages: render as-is
      if (msg.role === 'user') {
        items.push({ type: 'user', id: msg.id || `msg-${i}`, message: msg });
        i++;
        continue;
      }

      // Assistant messages: parse parts and render each separately (flat)
      if (msg.role === 'assistant') {
        const blocks = parseParts(msg);

        // Render each block as a separate item
        for (const block of blocks) {
          if (block.type === 'thinking') {
            items.push({
              type: 'thinking',
              id: `${msg.id}-thinking-${items.length}`,
              content: block.text || '',
            });
          } else if (block.type === 'text') {
            items.push({
              type: 'assistant_text',
              id: `${msg.id}-text-${items.length}`,
              content: block.text || '',
            });
          } else if (block.type === 'tool_use') {
            // Look for matching tool result in subsequent messages
            let result = '';
            let resultIdx = i + 1;

            while (resultIdx < taskMessages.length) {
              const resultMsg = taskMessages[resultIdx];
              if (resultMsg.role !== 'tool' && resultMsg.role !== 'tool_result') break;

              const resultBlocks = parseParts(resultMsg);
              const matchingResult = resultBlocks.find(
                (b) => b.type === 'tool_result' && b.tool_use_id === block.id
              );

              if (matchingResult) {
                result = matchingResult.content || '';
                break;
              } else if (resultMsg.content && !resultMsg.parts) {
                result = resultMsg.content;
                break;
              }
              resultIdx++;
            }

            items.push({
              type: 'tool',
              id: `${msg.id}-tool-${items.length}`,
              tool: block.name || 'tool',
              call: formatToolInput(block.input ?? block.content ?? block.text ?? ''),
              result,
            });
          }
        }

        // If no parts parsed but has content, render as text
        if (blocks.length === 0 && msg.content) {
          items.push({
            type: 'assistant_text',
            id: `${msg.id}-text-${items.length}`,
            content: msg.content,
          });
        }

        i++;
        continue;
      }

      // Tool messages not consumed: render as tool items
      if (msg.role === 'tool' || msg.role === 'tool_result') {
        const { tool, call } = parseToolMessage(msg);
        items.push({
          type: 'tool',
          id: msg.id || `tool-${i}`,
          tool,
          call,
          result: msg.content || '',
        });
        i++;
        continue;
      }

      // Other roles: skip
      i++;
    }

    return items;
  });

  const sessionItems = $derived.by(() => {
    if (mode !== 'session') return [] as SessionDisplayItem[];
    const sessionMessages = (messages as SessionMessage[]).filter(
      (msg) => !isSystemLikeMessage(msg)
    );
    const items: SessionDisplayItem[] = [];
    let i = 0;

    while (i < sessionMessages.length) {
      const msg = sessionMessages[i];
      if (msg.kind === 'tool_use' || msg.kind === 'tool_result') {
        if (msg.kind === 'tool_use') {
          let result = '';
          const next = sessionMessages[i + 1];
          if (
            next &&
            next.kind === 'tool_result' &&
            (!msg.tool_call_id || !next.tool_call_id || next.tool_call_id === msg.tool_call_id)
          ) {
            result = next.content || '';
            i++;
          }
          items.push({
            type: 'tool',
            id: msg.id,
            tool: msg.tool_name || 'tool',
            call: msg.content || '',
            result,
          });
        } else {
          items.push({
            type: 'tool',
            id: msg.id,
            tool: msg.tool_name || 'tool result',
            call: '',
            result: msg.content || '',
          });
        }
        i++;
        continue;
      }
      items.push({ type: 'message', id: msg.id, message: msg });
      i++;
    }

    return items;
  });

  const isEmpty = $derived.by(() =>
    mode === 'task' ? taskItems.length == 0 : sessionItems.length == 0
  );

  // Auto-scroll to bottom when messages change
  $effect(() => {
    const messageCount = messages.length;
    if (scrollContainerId && messageCount > 0) {
      tick().then(() => {
        const container = document.getElementById(scrollContainerId);
        if (container) {
          container.scrollTo({ top: container.scrollHeight, behavior: 'smooth' });
        }
      });
    }
  });
</script>

<div>
  {#if mode === 'task'}
    {#if isEmpty}
      <div class={cn('text-xs text-gray-500', emptyClass)}>{emptyText}</div>
    {:else}
      <div class={cn('space-y-1', className)}>
        {#each taskItems as item}
          {#if item.type === 'user'}
            <div class="flex gap-3 px-4 py-2 items-start">
              <div class="shrink-0 mt-1">
                {#if userAvatarUrl}
                  <img
                    src={userAvatarUrl}
                    alt="User"
                    class="w-8 h-8 rounded-full border border-white/10 shadow-sm"
                  />
                {:else}
                  <div
                    class="w-8 h-8 rounded-full bg-gray-800 border border-gray-700 flex items-center justify-center text-xs font-bold text-gray-300"
                  >
                    {userInitial}
                  </div>
                {/if}
              </div>
              <div
                class="flex-1 min-w-0 bg-[#1e1e1e]/60 border border-white/10 rounded-2xl px-4 py-3 text-[#FFFFFF] shadow-lg"
              >
                <p class="text-[13px] font-medium leading-relaxed">{item.message.content}</p>
              </div>
            </div>
          {:else if item.type === 'assistant_text'}
            <div class="px-12 py-2 pr-4">
              <MarkdownRenderer
                content={item.content}
                class="text-base text-[#FFFFFF] font-medium leading-relaxed font-sans"
              />
            </div>
          {:else if item.type === 'thinking'}
            <div class="px-12 py-1 pr-4">
              <div class="flex items-center gap-2 text-zinc-500/60">
                <div class="w-1.5 h-1.5 rounded-full bg-zinc-500/40"></div>
                <span class="text-xs uppercase tracking-wider font-bold">Thinking</span>
              </div>
              <div class="mt-1.5 pl-3.5 border-l border-white/[0.06]">
                {#if item.content.split('\n').length > 3 || item.content.length > 200}
                  {@const isExpanded = expandedThinking.has(item.id)}
                  <p
                    class="text-sm text-zinc-400/70 leading-relaxed whitespace-pre-wrap transition-all duration-200"
                    class:line-clamp-3={!isExpanded}
                  >
                    {item.content}
                  </p>
                  <button
                    type="button"
                    class="mt-1.5 text-xs text-zinc-500/80 hover:text-zinc-400/80 font-medium transition-colors flex items-center gap-1"
                    onclick={() => toggleThinking(item.id)}
                  >
                    {#if isExpanded}
                      <svg
                        class="w-3 h-3"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"><path d="m18 15-6-6-6 6" /></svg
                      >
                      Show less
                    {:else}
                      <svg
                        class="w-3 h-3"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"><path d="m6 9 6 6 6-6" /></svg
                      >
                      Show more
                    {/if}
                  </button>
                {:else}
                  <p class="text-sm text-zinc-400/70 leading-relaxed whitespace-pre-wrap">
                    {item.content}
                  </p>
                {/if}
              </div>
            </div>
          {:else if item.type === 'tool'}
            <div class="px-12 py-1 pr-4">
              <ToolBlock tool={item.tool} call={item.call} result={item.result} />
            </div>
          {/if}
        {/each}
      </div>
    {/if}
  {:else if isEmpty}
    <div class={cn('text-xs font-medium text-gray-500', emptyClass)}>{emptyText}</div>
  {:else}
    <div class={cn('space-y-1', className)}>
      {#each sessionItems as item}
        {#if item.type === 'tool'}
          <ToolBlock tool={item.tool} call={item.call} result={item.result} />
        {:else}
          <div class="flex gap-3 px-2 py-6 items-start">
            {#if item.message.role === 'user'}
              <div class="shrink-0 mt-1">
                {#if userAvatarUrl}
                  <img
                    src={userAvatarUrl}
                    alt="User"
                    class="w-7 h-7 rounded-full border border-white/10 shadow-sm"
                  />
                {:else}
                  <div
                    class="w-7 h-7 rounded-full bg-gray-800 border border-gray-700 flex items-center justify-center text-[9px] font-bold text-gray-300"
                  >
                    {userInitial}
                  </div>
                {/if}
              </div>
              <div
                class="flex-1 min-w-0 bg-violet-500/10 border border-white/5 rounded-2xl px-4 py-3 text-[#FFFFFF] shadow-sm"
              >
                <div class="text-base font-medium whitespace-pre-wrap break-words">
                  {item.message.content || ''}
                </div>
              </div>
            {:else}
              <div class="px-3 py-2 text-xs font-medium">
                <div class="text-base whitespace-pre-wrap break-words">
                  {item.message.content || ''}
                </div>
              </div>
            {/if}
          </div>
        {/if}
      {/each}
    </div>
  {/if}
</div>
