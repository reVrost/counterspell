<script lang="ts">
  import { cn } from '$lib/utils';

  interface Props {
    tool: string;
    call?: string;
    result?: string;
    class?: string;
  }

  let { tool, call = '', result = '', class: className = '' }: Props = $props();

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
  const isShellTool = $derived.by(() =>
    /(bash|shell|zsh|sh|cmd|powershell)/.test(normalizedTool)
  );
  const isWriteTool = $derived.by(() => /(write|edit|patch|diff|apply_patch)/.test(normalizedTool));
  const callIsDiff = $derived.by(() => looksLikeDiff(call) || isWriteTool);
  const resultIsDiff = $derived.by(() => looksLikeDiff(result) && !callIsDiff);

  const callClass = $derived.by(() => {
    if (!call) return '';
    if (callIsDiff) return 'max-h-[496px] overflow-y-auto bg-[#0D1117]';
    if (isShellTool) return 'max-h-[228px] overflow-y-auto';
    return 'max-h-[360px] overflow-y-auto';
  });

  const resultClass = $derived.by(() => {
    if (!result) return '';
    if (resultIsDiff) return 'max-h-[496px] overflow-y-auto bg-[#0D1117]';
    if (isShellTool) return 'max-h-[228px] overflow-y-auto';
    return 'max-h-[360px] overflow-y-auto';
  });

  const toolLabel = $derived.by(() => (tool && tool.trim() ? tool.trim() : 'tool'));
</script>

<div
  class={cn(
    'rounded-xl border border-white/10 bg-[#0b0b0b] shadow-[0_8px_24px_rgba(0,0,0,0.35)] overflow-hidden',
    className
  )}
>
  <div
    class="flex items-center justify-between px-3 py-1.5 bg-gradient-to-r from-white/5 to-transparent border-b border-white/10"
  >
    <div class="flex items-center gap-2 min-w-0">
      <div class="h-2 w-2 rounded-full bg-emerald-400/70 shadow-[0_0_10px_rgba(52,211,153,0.4)]"></div>
      <span
        class="text-[10px] font-semibold uppercase tracking-[0.28em] text-gray-500 truncate"
        title={toolLabel}
      >
        {toolLabel}
      </span>
    </div>
    <span class="text-[9px] text-gray-600 uppercase tracking-[0.2em]">tool</span>
  </div>

  {#if call}
    <div
      class={cn(
        'px-3 py-2 text-xs text-gray-300 font-mono overflow-x-auto leading-snug',
        callIsDiff ? 'border-b border-white/5' : '',
        callClass
      )}
    >
      <div class="text-[9px] uppercase tracking-[0.2em] text-gray-500 mb-1">call</div>
      <pre class="whitespace-pre-wrap break-words leading-tight">{call}</pre>
    </div>
  {/if}

  {#if result}
    <div
      class={cn(
        'px-3 py-2 text-xs text-gray-300 font-mono overflow-x-auto leading-snug',
        !call ? '' : 'border-t border-white/5',
        resultClass
      )}
    >
      <div class="text-[9px] uppercase tracking-[0.2em] text-gray-500 mb-1">result</div>
      <pre class="whitespace-pre-wrap break-words leading-tight">{result}</pre>
    </div>
  {/if}
</div>
