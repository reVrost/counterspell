<script lang="ts">
  import { cn } from '$lib/utils';
  import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';

  interface Props {
    tool: string;
    call?: string;
    result?: string;
    class?: string;
  }

  let { tool, call = '', result = '', class: className = '' }: Props = $props();
  let expanded = $state(false);

  const callCmd = $derived.by(() => {
    if (!call) return '';
    try {
      const parsed = JSON.parse(call);
      return parsed.cmd || parsed.command || parsed.path || call;
    } catch {
      return call;
    }
  });

  const summaryCmd = $derived.by(() => {
    if (!callCmd) return '';
    const firstLine = callCmd.split('\n')[0].trim();
    if (firstLine.length > 50) return firstLine.slice(0, 50) + '...';
    return firstLine + (callCmd.includes('\n') ? '...' : '');
  });

  function looksLikeDiff(text: string): boolean {
    if (!text) return false;
    const trimmed = text.trim();
    if (!trimmed) return false;
    return (
      trimmed.startsWith('*** Begin Patch') ||
      trimmed.startsWith('diff --git') ||
      trimmed.startsWith('@@') ||
      trimmed.includes('\n@@') ||
      /^\+\+\+|^---/m.test(trimmed)
    );
  }

  const normalizedTool = $derived.by(() => (tool || '').trim().toLowerCase());
  const isShellTool = $derived.by(() => /(bash|shell|zsh|sh|cmd|powershell)/.test(normalizedTool));
  const isWriteTool = $derived.by(() => /(write|edit|patch|diff|apply_patch)/.test(normalizedTool));
  const callIsDiff = $derived.by(() => looksLikeDiff(call) || isWriteTool);
  const resultIsDiff = $derived.by(() => looksLikeDiff(result) && !callIsDiff);

  const resultClass = $derived.by(() => {
    if (!result) return '';
    if (resultIsDiff) return 'max-h-[400px] overflow-y-auto bg-[#0D1117]';
    return 'max-h-[300px] overflow-y-auto';
  });

  const toolLabel = $derived.by(() =>
    tool && tool.trim() ? tool.trim().replace('_command', '') : 'tool'
  );

  const friendlyInfo = $derived.by(() => {
    const lowTool = (tool || '').toLowerCase();
    const lowCall = (callCmd || '').toLowerCase();
    let label = toolLabel;
    let icon = '🔧';
    let summary = summaryCmd;

    // Detect high-level action
    if (
      lowTool.includes('search') ||
      lowTool.includes('grep') ||
      lowCall.includes('grep ') ||
      lowCall.includes('find ')
    ) {
      label = 'Searching';
      icon = '🔍';
    } else if (
      lowTool.includes('read') ||
      lowCall.includes('cat ') ||
      lowCall.includes('sed -n') ||
      lowCall.includes('view_file')
    ) {
      label = 'Reading';
      icon = '📖';
    } else if (
      lowTool.includes('write') ||
      lowTool.includes('edit') ||
      lowTool.includes('patch') ||
      lowTool.includes('replace') ||
      lowCall.includes('sed -i') ||
      (lowCall.includes('python3') && (lowCall.includes('write') || lowCall.includes('save')))
    ) {
      label = 'Editing';
      icon = '📝';
    } else if (
      lowTool.includes('list') ||
      lowTool.includes('ls') ||
      lowCall.startsWith('ls ') ||
      lowCall === 'ls' ||
      lowCall.includes('dir ')
    ) {
      label = 'Exploring';
      icon = '📂';
    } else if (isShellTool) {
      label = 'Executing';
      icon = '⚡';
    }

    // Try to extract a path for the summary
    const pathMatch = callCmd.match(/(?:\/|(?:\.\.\/)+)[a-zA-Z0-9._\-\/]+\.[a-zA-Z0-9]+/);
    if (pathMatch) {
      const fullPath = pathMatch[0];
      const parts = fullPath.split('/');
      const filename = parts[parts.length - 1];
      summary = filename;
      // If it's a very common filename (like index.ts), show a bit of parent
      if (
        (filename === 'index.ts' || filename === 'index.js' || filename === 'main.go') &&
        parts.length > 1
      ) {
        summary = `${parts[parts.length - 2]}/${filename}`;
      }
    }

    return { label, icon, summary };
  });
</script>

<div
  class={cn(
    'rounded-xl border border-white/10 bg-[#0b0b0b] shadow-[0_4px_12px_rgba(0,0,0,0.2)] overflow-hidden transition-all duration-200',
    expanded ? 'my-3 ring-1 ring-white/5' : 'my-1',
    className
  )}
>
  <button
    onclick={() => (expanded = !expanded)}
    class="w-full flex items-center justify-between px-3 py-1.5 bg-gradient-to-r from-white/[0.03] to-transparent hover:bg-white/[0.05] active:bg-white/[0.07] transition-colors group text-left outline-none"
  >
    <div class="flex items-center gap-2.5 min-w-0">
      <div
        class={cn(
          'w-4 h-4 rounded flex items-center justify-center transition-transform duration-200',
          expanded ? 'rotate-90 text-violet-400' : 'text-zinc-500'
        )}
      >
        <ChevronRightIcon class="w-3.5 h-3.5" />
      </div>

      <div class="flex items-center gap-2 min-w-0">
        <div class="flex flex-col min-w-0">
          <div class="flex items-center gap-1.5 font-mono text-[12px]">
            <span
              class={cn(
                'font-bold lowercase opacity-70',
                friendlyInfo.label === 'Thinking' ? 'text-zinc-500' : 'text-violet-400'
              )}>{friendlyInfo.label}:</span
            >
            <span class="text-zinc-300 truncate font-medium">{friendlyInfo.summary}</span>
          </div>
        </div>
      </div>
    </div>
    <div class="flex items-center gap-2 shrink-0 ml-4">
      <div
        class={cn(
          'h-1 w-1 rounded-full',
          friendlyInfo.label === 'Thinking'
            ? 'bg-zinc-500/50'
            : 'bg-violet-500/50 shadow-[0_0_8px_rgba(16,185,129,0.3)]'
        )}
      ></div>
      <span class="text-[9px] text-zinc-600 font-bold uppercase tracking-widest hidden sm:block"
        >{toolLabel}</span
      >
    </div>
  </button>

  {#if expanded}
    <div class="border-t border-white/5 bg-black/20">
      {#if call && callCmd !== call}
        <div class="px-3 py-2 border-b border-white/5">
          <p class="text-[10px] uppercase tracking-wider text-zinc-600 font-bold mb-1">Full Call</p>
          <pre
            class="text-[11px] text-zinc-400 font-mono whitespace-pre-wrap leading-relaxed">{call}</pre>
        </div>
      {/if}

      {#if callCmd && callCmd.includes('\n')}
        <div class="px-3 py-2 border-b border-white/5">
          <p class="text-[10px] uppercase tracking-wider text-zinc-600 font-bold mb-1">Command</p>
          <pre
            class="text-[11px] text-violet-400/90 font-mono whitespace-pre-wrap leading-relaxed">{callCmd}</pre>
        </div>
      {/if}

      {#if result}
        <div class={cn('px-3 py-2 text-zinc-300 font-mono leading-relaxed', resultClass)}>
          <p class="text-[10px] uppercase tracking-wider text-zinc-600 font-bold mb-2">Result</p>
          <pre class="text-[11px] whitespace-pre-wrap break-words">{result}</pre>
        </div>
      {:else}
        <div class="px-3 py-2 text-[11px] text-zinc-600 italic">No output result.</div>
      {/if}
    </div>
  {/if}
</div>
