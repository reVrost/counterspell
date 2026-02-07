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

  type ToolItem = { tool: string; call: string; result: string };
  type TaskDisplayItem =
    | { type: 'message'; id: string; message: Message }
    | { type: 'assistant'; id: string; message: Message; items: ToolItem[] }
    | { type: 'thinking'; id: string; items: ToolItem[] };

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
        items.push({ type: 'message', id: msg.id || `msg-${i}`, message: msg });
        i++;
        continue;
      }

      // Assistant messages: group all parts (thinking + text + tools) together
      if (msg.role === 'assistant') {
        const blocks = parseParts(msg);
        const toolItems: ToolItem[] = [];

        // Extract thinking blocks
        const thinkingBlocks = blocks.filter((b) => b.type === 'thinking');
        for (const block of thinkingBlocks) {
          toolItems.push({ tool: 'thinking', call: block.text || '', result: '' });
        }

        // Extract tool uses and pair with results from subsequent tool messages
        const toolUses = blocks.filter((b) => b.type === 'tool_use');
        let resultIdx = i + 1;

        for (const tool of toolUses) {
          let result = '';
          // Look for matching tool result in subsequent messages
          while (resultIdx < taskMessages.length) {
            const resultMsg = taskMessages[resultIdx];
            if (resultMsg.role !== 'tool' && resultMsg.role !== 'tool_result') break;

            const resultBlocks = parseParts(resultMsg);
            const matchingResult = resultBlocks.find(
              (b) => b.type === 'tool_result' && b.tool_use_id === tool.id
            );

            if (matchingResult) {
              result = matchingResult.content || '';
              resultIdx++;
              break;
            } else if (resultMsg.content && !resultMsg.parts) {
              // Legacy format: tool message with content
              result = resultMsg.content;
              resultIdx++;
              break;
            }
            resultIdx++;
          }

          toolItems.push({
            tool: tool.name || 'tool',
            call: formatToolInput(tool.input ?? tool.content ?? tool.text ?? ''),
            result,
          });
        }

        // Consume tool messages that were matched
        const consumedToolMessages = resultIdx - i - 1;

        // Render as assistant item with all its parts
        items.push({
          type: 'assistant',
          id: msg.id || `assistant-${i}`,
          message: msg,
          items: toolItems,
        });

        i += 1 + consumedToolMessages;
        continue;
      }

      // Tool messages not consumed by assistant: render as standalone thinking items
      if (msg.role === 'tool' || msg.role === 'tool_result') {
        const toolItems: ToolItem[] = [];
        let j = i;

        while (j < taskMessages.length) {
          const toolMsg = taskMessages[j];
          if (toolMsg.role !== 'tool' && toolMsg.role !== 'tool_result') break;

          const { tool, call } = parseToolMessage(toolMsg);
          toolItems.push({
            tool,
            call,
            result: toolMsg.content || '',
          });
          j++;
        }

        if (toolItems.length > 0) {
          items.push({
            type: 'thinking',
            id: msg.id || `tools-${i}`,
            items: toolItems,
          });
        }

        i = j;
        continue;
      }

      // Other roles: render as message
      items.push({ type: 'message', id: msg.id || `msg-${i}`, message: msg });
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

  $effect(() => {
    if (scrollContainerId) {
      tick().then(() => {
        const container = document.getElementById(scrollContainerId);
        if (container) {
          container.scrollTo({ top: container.scrollHeight });
        }
      });
    }
  });

  function categorizeGroup(items: ToolItem[]): { label: string; color: string } {
    if (items.length === 0) return { label: 'Thinking', color: 'text-zinc-500' };

    const firstItem = items[0];

    // If they are all similar, type the whole group
    const isAllReading = items.every(
      (item) =>
        (item.tool || '').toLowerCase().includes('read') ||
        (item.call || '').toLowerCase().includes('cat ') ||
        (item.call || '').toLowerCase().includes('view_file') ||
        (item.call || '').toLowerCase().includes('sed -n')
    );
    if (isAllReading) return { label: 'Reading', color: 'text-emerald-500/80' };

    const isAllExploring = items.every(
      (item) =>
        (item.tool || '').toLowerCase().includes('ls') ||
        (item.tool || '').toLowerCase().includes('list') ||
        (item.call || '').toLowerCase().startsWith('ls ')
    );
    if (isAllExploring) return { label: 'Exploring', color: 'text-emerald-500/80' };

    const isAllSearching = items.every(
      (item) =>
        (item.tool || '').toLowerCase().includes('search') ||
        (item.tool || '').toLowerCase().includes('grep')
    );
    if (isAllSearching) return { label: 'Searching', color: 'text-emerald-500/80' };

    const isAllEditing = items.every(
      (item) =>
        (item.tool || '').toLowerCase().includes('write') ||
        (item.tool || '').toLowerCase().includes('edit') ||
        (item.tool || '').toLowerCase().includes('patch')
    );
    if (isAllEditing) return { label: 'Editing', color: 'text-emerald-500/80' };

    return { label: 'Thinking', color: 'text-zinc-500' };
  }
</script>

<div>
  {#if mode === 'task'}
    {#if isEmpty}
      <div class={cn('text-xs text-gray-500', emptyClass)}>{emptyText}</div>
    {:else}
      <div class={cn('space-y-1', className)}>
        {#each taskItems as item}
          {#if item.type === 'message'}
            {#if item.message.role === 'user'}
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
                      class="w-8 h-8 rounded-full bg-gray-800 border border-gray-700 flex items-center justify-center text-[10px] font-bold text-gray-300"
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
            {:else if item.message.role === 'assistant'}
              <div class="px-12 py-2 pr-4">
                <MarkdownRenderer
                  content={item.message.content}
                  class="text-[13px] text-[#FFFFFF] font-medium leading-relaxed font-sans"
                />
              </div>
            {:else}
              <div class="px-12 py-2 pr-4 opacity-70">
                <p class="text-[13px] text-[#FFFFFF] font-medium leading-relaxed font-sans">
                  {item.message.content}
                </p>
              </div>
            {/if}
          {:else if item.type === 'assistant'}
            {@const blocks = parseParts(item.message)}
            {@const textBlocks = blocks.filter((b) => b.type === 'text')}
            {@const textContent = textBlocks.map((b) => b.text).join('')}
            {@const hasThinking = item.items.some((it) => it.tool === 'thinking')}
            {@const hasTools = item.items.some((it) => it.tool !== 'thinking')}

            <div class="px-12 py-2 pr-4 space-y-2">
              <!-- Thinking blocks (collapsible) -->
              {#if hasThinking}
                {@const cat = categorizeGroup(item.items)}
                <details class="my-2 group" open>
                  <summary
                    class="flex items-center gap-2.5 cursor-pointer text-zinc-500 hover:text-zinc-300 transition-colors list-none outline-none select-none py-1"
                  >
                    <div
                      class="w-3.5 h-3.5 flex items-center justify-center group-open:rotate-90 transition-transform opacity-60"
                    >
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="12"
                        height="12"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="3"
                        stroke-linecap="round"
                        stroke-linejoin="round"><path d="m9 18 6-6-6-6" /></svg
                      >
                    </div>
                    <div class="flex items-center gap-2">
                      <span
                        class={cn('text-[10px] font-black tracking-[0.15em] uppercase', cat.color)}
                        >{cat.label}</span
                      >
                      <span class="text-[10px] opacity-40 font-mono"
                        >({item.items.filter((it) => it.tool === 'thinking').length})</span
                      >
                    </div>
                  </summary>
                  <div
                    class="mt-2 space-y-1.5 border-l border-white/[0.06] ml-[6px] pl-4 transition-all"
                  >
                    {#each item.items.filter((it) => it.tool === 'thinking') as toolItem}
                      <ToolBlock
                        tool={toolItem.tool}
                        call={toolItem.call}
                        result={toolItem.result}
                      />
                    {/each}
                  </div>
                </details>
              {/if}

              <!-- Text content -->
              {#if textContent.trim()}
                <MarkdownRenderer
                  content={textContent}
                  class="text-[13px] text-[#FFFFFF] font-medium leading-relaxed font-sans"
                />
              {/if}

              <!-- Tool uses -->
              {#if hasTools}
                <div class="space-y-1.5">
                  {#each item.items.filter((it) => it.tool !== 'thinking') as toolItem}
                    <ToolBlock tool={toolItem.tool} call={toolItem.call} result={toolItem.result} />
                  {/each}
                </div>
              {/if}
            </div>
          {:else}
            {@const cat = categorizeGroup(item.items)}
            <details class="my-2 group" open>
              <summary
                class="flex items-center gap-2.5 cursor-pointer text-zinc-500 hover:text-zinc-300 transition-colors list-none outline-none select-none py-1"
              >
                <div
                  class="w-3.5 h-3.5 flex items-center justify-center group-open:rotate-90 transition-transform opacity-60"
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="12"
                    height="12"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="3"
                    stroke-linecap="round"
                    stroke-linejoin="round"><path d="m9 18 6-6-6-6" /></svg
                  >
                </div>
                <div class="flex items-center gap-2">
                  <span class={cn('text-[10px] font-black tracking-[0.15em] uppercase', cat.color)}
                    >{cat.label}</span
                  >
                  <span class="text-[10px] opacity-40 font-mono">({item.items.length})</span>
                </div>
              </summary>
              <div
                class="mt-2 space-y-1.5 border-l border-white/[0.06] ml-[6px] pl-4 transition-all"
              >
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
                <div class="text-[13px] font-medium whitespace-pre-wrap break-words">
                  {item.message.content || ''}
                </div>
              </div>
            {:else}
              <div class="px-3 py-2 text-xs font-medium">
                <div class="text-[13px] whitespace-pre-wrap break-words">
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
