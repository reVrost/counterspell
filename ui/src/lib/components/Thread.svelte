<script lang="ts">
  import { cn } from '$lib/utils';
  import type { ContentBlock, Message, SessionMessage } from '$lib/types';
  import MarkdownRenderer from './MarkdownRenderer.svelte';
  import ToolBlock from './ToolBlock.svelte';

  interface Props {
    mode: 'task' | 'session';
    messages: Message[] | SessionMessage[];
    emptyText?: string;
    emptyClass?: string;
    class?: string;
  }

  let {
    mode,
    messages,
    emptyText = 'No messages yet.',
    emptyClass = 'text-xs text-gray-500',
    class: className = '',
  }: Props = $props();

  type ThinkingItem = { tool: string; call: string; result: string };
  type TaskDisplayItem =
    | { type: 'message'; id: string; message: Message }
    | { type: 'thinking'; id: string; items: ThinkingItem[] };

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
      const blocks = parseParts(msg);
      const text = concatText(blocks);
      const thinkingBlocks = blocks.filter((b) => b.type === 'thinking');
      const toolUses = blocks.filter((b) => b.type === 'tool_use');
      const toolResults = blocks.filter((b) => b.type === 'tool_result');

      const hasToolRole = msg.role === 'tool' || msg.role === 'tool_result';
      const hasThinking = thinkingBlocks.length > 0;
      const hasToolBlocks = toolUses.length > 0 || toolResults.length > 0;

      if (text) {
        items.push({ type: 'message', id: msg.id || `msg-${i}`, message: msg });
      }

      if (hasToolRole || hasThinking || hasToolBlocks) {
        const thinkingItems: ThinkingItem[] = [];
        const groupId = msg.id || `thinking-${i}`;

        for (const block of thinkingBlocks) {
          thinkingItems.push({ tool: 'thinking', call: block.text || '', result: '' });
        }

        const remainingResults = [...toolResults];
        for (const tool of toolUses) {
          let result = '';
          if (remainingResults.length > 0) {
            const match = remainingResults.shift();
            result = match?.content || '';
          } else if (i + 1 < taskMessages.length) {
            const next = taskMessages[i + 1];
            const nextBlocks = parseParts(next);
            const nextResult = nextBlocks.find((b) => b.type === 'tool_result');
            if (next.role === 'tool_result' || nextResult) {
              result = nextResult?.content || next.content || '';
              i++;
            }
          }
          const toolName = tool.name || 'tool';
          const call = formatToolInput(tool.input ?? tool.content ?? tool.text ?? '');
          thinkingItems.push({ tool: toolName, call, result });
        }

        for (const block of remainingResults) {
          thinkingItems.push({ tool: 'tool result', call: '', result: block.content || '' });
        }

        if (hasToolRole && toolUses.length === 0 && toolResults.length === 0 && msg.content) {
          if (msg.role === 'tool') {
            const { tool, call } = parseToolMessage(msg);
            thinkingItems.push({ tool, call, result: '' });
          } else {
            thinkingItems.push({ tool: 'tool result', call: '', result: msg.content });
          }
        }

        if (thinkingItems.length > 0) {
          items.push({ type: 'thinking', id: groupId, items: thinkingItems });
        }
      }

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
</script>

{#if mode === 'task'}
  {#if isEmpty}
    <div class={cn('text-xs text-gray-500', emptyClass)}>{emptyText}</div>
  {:else}
    <div class={cn('space-y-1', className)}>
      {#each taskItems as item}
        {#if item.type === 'message'}
          {#if item.message.role === 'user'}
            <div class="flex gap-4 px-4 py-2 items-start">
              <div
                class="flex-1 min-w-0 bg-[#1e1e1e]/60 border border-white/10 rounded-xl px-4 py-3 text-[#FFFFFF] shadow-lg"
              >
                <p class="text-lg font-medium leading-relaxed">{item.message.content}</p>
              </div>
            </div>
          {:else if item.message.role === 'assistant'}
            <div class="px-12 py-2 pr-4">
              <MarkdownRenderer
                content={item.message.content}
                class="text-base text-[#FFFFFF] font-medium leading-relaxed font-sans"
              />
            </div>
          {:else}
            <div class="px-12 py-2 pr-4">
              <p class="text-base text-[#FFFFFF] font-medium leading-relaxed font-sans">
                {item.message.content}
              </p>
            </div>
          {/if}
        {:else}
          <details class="mx-12 my-4 group" open>
            <summary
              class="flex items-center gap-2 cursor-pointer text-gray-500 hover:text-gray-300 transition-colors list-none outline-none"
            >
              <div
                class="w-4 h-4 flex items-center justify-center group-open:rotate-90 transition-transform"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="14"
                  height="14"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2.5"
                  stroke-linecap="round"
                  stroke-linejoin="round"><path d="m9 18 6-6-6-6" /></svg
                >
              </div>
              <span class="text-xs font-bold tracking-widest uppercase">Thinking</span>
            </summary>
            <div class="mt-3 space-y-3">
              {#each item.items as toolItem}
                <ToolBlock tool={toolItem.tool} call={toolItem.call} result={toolItem.result} />
              {/each}
            </div>
          </details>
        {/if}
      {/each}
    </div>
  {/if}
{:else if isEmpty}
  <div class={cn('text-xs font-medium text-gray-500', emptyClass)}>{emptyText}</div>
{:else}
  <div class={cn('space-y-2', className)}>
    {#each sessionItems as item}
      {#if item.type === 'tool'}
        <ToolBlock tool={item.tool} call={item.call} result={item.result} />
      {:else}
        <div
          class={cn(
            'rounded-lg border px-3 py-2 text-xs font-medium',
            item.message.role === 'user'
              ? 'border-violet-500/30 bg-violet-500/10 '
              : 'border-white/10 bg-white/5 '
          )}
        >
          <div class="flex items-center justify-between text-sm uppercase text-gray-500 mb-1">
            <span>{item.message.role}</span>
            <span>{item.message.kind}</span>
          </div>
          <div class="text-base whitespace-pre-wrap break-words">
            {item.message.content || ''}
          </div>
        </div>
      {/if}
    {/each}
  </div>
{/if}
