<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { FileDiff, parsePatchFiles } from '@pierre/diffs';
  import type { FileDiffMetadata, ParsedPatch } from '@pierre/diffs';
  import { cn } from '$lib/utils';

  interface Props {
    diff: string;
    class?: string;
  }

  interface FileStat {
    filename: string;
    additions: number;
    deletions: number;
  }

  let { diff, class: className = '' }: Props = $props();
  let container: HTMLElement;
  let fileDiffs: FileDiff[] = [];
  
  // Parse the diff to get file statistics
  export function getFileStats(): FileStat[] {
    if (!diff) return [];
    
    const stats: FileStat[] = [];
    const cleanDiff = stripAnsiCodes(diff);
    const lines = cleanDiff.split('\n');
    let currentFile = '';
    let additions = 0;
    let deletions = 0;
    
    for (const line of lines) {
      // Handle "added: right/lotr.md" format from backend
      const addedMatch = line.match(/^added:\s*(.+)$/);
      const modifiedMatch = line.match(/^modified:\s*(.+)$/);
      const deletedMatch = line.match(/^deleted:\s*(.+)$/);
      
      if (addedMatch || modifiedMatch || deletedMatch) {
        if (currentFile) {
          stats.push({ filename: currentFile, additions, deletions });
        }
        currentFile = addedMatch?.[1] || modifiedMatch?.[1] || deletedMatch?.[1] || '';
        additions = 0;
        deletions = 0;
      }
      // Handle standard git diff format
      else if (line.startsWith('diff --git')) {
        if (currentFile) {
          stats.push({ filename: currentFile, additions, deletions });
        }
        currentFile = '';
        additions = 0;
        deletions = 0;
      } else if (line.startsWith('+++ b/') || line.startsWith('--- a/')) {
        const match = line.match(/^[+-]{3} [ab]\/(.+)$/);
        if (match && !currentFile) {
          currentFile = match[1];
        }
      } else if (line.startsWith('+') && !line.startsWith('+++')) {
        additions++;
      } else if (line.startsWith('-') && !line.startsWith('---')) {
        deletions++;
      }
    }
    
    if (currentFile) {
      stats.push({ filename: currentFile, additions, deletions });
    }
    
    return stats;
  }

  let parsedPatches: ParsedPatch[] = $derived(diff ? parsePatchFiles(diff) : []);

  // Re-render when diff changes
  $effect(() => {
    if (container) {
      renderDiffs();
    }
  });

  function stripAnsiCodes(str: string): string {
    // eslint-disable-next-line no-control-regex
    return str.replace(/\u001b\[[0-9;]*[a-zA-Z]/g, '');
  }

  function isStandardGitDiff(str: string): boolean {
    return str.includes('diff --git') || str.includes('@@');
  }

  function renderSimpleDiff(diffText: string): string {
    const lines = diffText.split('\n');
    let html = '';
    let currentFile: string | null = null;
    let inHunk = false;
    let lineNumber = 0;
    
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      const cleanLine = stripAnsiCodes(line);
      const trimmedLine = cleanLine.trim();
      
      if (!trimmedLine && !inHunk) {
        continue;
      }
      
      // File header: diff --git a/file b/file
      if (cleanLine.startsWith('diff --git')) {
        if (currentFile) {
          html += '</div></div>'; // Close previous file
        }
        currentFile = cleanLine;
        inHunk = false;
        continue;
      }
      
      // New file mode
      if (cleanLine.startsWith('new file mode')) {
        html += `<div class="mb-4 border border-gray-700 rounded-lg overflow-hidden">`;
        html += `<div class="px-4 py-2 bg-gray-800/50 border-b border-gray-700 flex items-center gap-2">`;
        html += `<span class="text-xs font-medium text-green-400">+</span>`;
        html += `<span class="text-sm text-gray-300 font-mono">${escapeHtml(currentFile?.replace('diff --git a/', '').split(' b/')[0] || '')}</span>`;
        html += `<span class="text-xs text-gray-500 ml-auto">${escapeHtml(trimmedLine)}</span>`;
        html += `</div>`;
        html += `<div class="bg-[#0d1117]">`;
        continue;
      }
      
      // Deleted file mode
      if (cleanLine.startsWith('deleted file mode')) {
        html += `<div class="mb-4 border border-gray-700 rounded-lg overflow-hidden">`;
        html += `<div class="px-4 py-2 bg-gray-800/50 border-b border-gray-700 flex items-center gap-2">`;
        html += `<span class="text-xs font-medium text-red-400">-</span>`;
        html += `<span class="text-sm text-gray-300 font-mono">${escapeHtml(currentFile?.replace('diff --git a/', '').split(' b/')[0] || '')}</span>`;
        html += `<span class="text-xs text-gray-500 ml-auto">${escapeHtml(trimmedLine)}</span>`;
        html += `</div>`;
        html += `<div class="bg-[#0d1117]">`;
        continue;
      }
      
      // Index line - skip
      if (cleanLine.startsWith('index ')) {
        continue;
      }
      
      // --- and +++ lines
      if (cleanLine.startsWith('--- ') || cleanLine.startsWith('+++ ')) {
        continue;
      }
      
      // Hunk header: @@ -1,2 +3,4 @@
      if (cleanLine.startsWith('@@')) {
        inHunk = true;
        const match = cleanLine.match(/@@\s+(\-\d+(?:,\d+)?)\s+(\+\d+(?:,\d+)?)\s+@@/);
        if (match) {
          html += `<div class="px-4 py-1 bg-gray-800/30 text-gray-500 text-xs font-mono border-y border-gray-800">`;
          html += `<span class="text-gray-600">@@</span> ${escapeHtml(match[1])} <span class="text-gray-600">${match[2]}</span> <span class="text-gray-600">@@</span>`;
          html += `</div>`;
        }
        continue;
      }
      
      // Addition line
      if (cleanLine.startsWith('+') && !cleanLine.startsWith('+++')) {
        lineNumber++;
        const content = cleanLine.substring(1);
        html += `<div class="flex hover:bg-green-500/5 transition-colors">`;
        html += `<div class="w-12 px-2 py-0.5 text-right text-gray-600 text-xs select-none border-r border-gray-800 bg-gray-900/30">${lineNumber}</div>`;
        html += `<div class="flex-1 px-3 py-0.5 text-green-400 font-mono text-sm">`;
        html += `<span class="text-green-600 mr-2">+</span>${escapeHtml(content) || '&nbsp;'}`;
        html += `</div></div>`;
        continue;
      }
      
      // Deletion line
      if (cleanLine.startsWith('-') && !cleanLine.startsWith('---')) {
        const content = cleanLine.substring(1);
        html += `<div class="flex hover:bg-red-500/5 transition-colors">`;
        html += `<div class="w-12 px-2 py-0.5 text-right text-gray-600 text-xs select-none border-r border-gray-800 bg-gray-900/30">-</div>`;
        html += `<div class="flex-1 px-3 py-0.5 text-red-400 font-mono text-sm">`;
        html += `<span class="text-red-600 mr-2">-</span>${escapeHtml(content) || '&nbsp;'}`;
        html += `</div></div>`;
        continue;
      }
      
      // Context line (no prefix)
      if (cleanLine.startsWith(' ')) {
        lineNumber++;
        const content = cleanLine.substring(1);
        html += `<div class="flex hover:bg-gray-800/30 transition-colors">`;
        html += `<div class="w-12 px-2 py-0.5 text-right text-gray-600 text-xs select-none border-r border-gray-800 bg-gray-900/30">${lineNumber}</div>`;
        html += `<div class="flex-1 px-3 py-0.5 text-gray-300 font-mono text-sm">`;
        html += `<span class="text-gray-700 mr-2"> </span>${escapeHtml(content) || '&nbsp;'}`;
        html += `</div></div>`;
        continue;
      }
      
      // Regular line (shouldn't happen in git diff but handle it)
      if (trimmedLine) {
        html += `<div class="flex">`;
        html += `<div class="w-12 px-2 py-0.5 border-r border-gray-800 bg-gray-900/30"></div>`;
        html += `<div class="flex-1 px-3 py-0.5 text-gray-400 font-mono text-sm">${escapeHtml(cleanLine)}</div>`;
        html += `</div>`;
      }
    }
    
    // Close last file container
    if (currentFile) {
      html += '</div></div>';
    }
    
    return html || '<div class="p-4 text-gray-500 italic">No changes</div>';
  }

  function escapeHtml(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  function renderDiffs() {
    console.log('[DiffRenderer] renderDiffs called');
    console.log('[DiffRenderer] container exists:', !!container);
    console.log('[DiffRenderer] diff length:', diff?.length || 0);
    
    if (!container) {
      console.log('[DiffRenderer] No container, returning');
      return;
    }
    
    // Clean up existing
    fileDiffs.forEach(fd => fd.cleanUp());
    fileDiffs = [];
    
    // Clear container
    container.innerHTML = '';
    
    if (!diff || diff.trim().length === 0) {
      console.log('[DiffRenderer] Empty diff, showing "No changes"');
      container.innerHTML = '<div class="p-4 text-gray-500 italic">No changes</div>';
      return;
    }
    
    // Check if this is a standard git diff or terminal-formatted output
    const cleanDiff = stripAnsiCodes(diff);
    console.log('[DiffRenderer] Clean diff preview:', cleanDiff.substring(0, 100));
    console.log('[DiffRenderer] isStandardGitDiff:', isStandardGitDiff(cleanDiff));
    console.log('[DiffRenderer] parsedPatches length:', parsedPatches?.length || 0);
    
    // Always use simple renderer for now - Pierre has issues
    console.log('[DiffRenderer] Using simple renderer');
    container.innerHTML = renderSimpleDiff(diff);
    console.log('[DiffRenderer] Simple renderer completed');
  }

  onDestroy(() => {
    fileDiffs.forEach(fd => fd.cleanUp());
    fileDiffs = [];
  });
</script>

<div bind:this={container} class={cn('diff-renderer min-h-[100px]', className)}></div>

<style>
  :global(.diff-renderer) {
    --diffs-bg: #0d1117;
    --diffs-fg: #e6edf3;
    --diffs-line-number-color: #6e7681;
    --diffs-addition-bg: rgba(46, 160, 67, 0.15);
    --diffs-addition-fg: #3fb950;
    --diffs-deletion-bg: rgba(248, 81, 73, 0.15);
    --diffs-deletion-fg: #f85149;
  }
  
  :global(.diff-renderer [class*="file-header"]) {
    background: rgba(255, 255, 255, 0.05) !important;
    border: 1px solid rgba(255, 255, 255, 0.1) !important;
    border-radius: 8px 8px 0 0;
    padding: 12px 16px !important;
  }
  
  :global(.diff-renderer pre) {
    background: #0b0f17 !important;
    border: 1px solid rgba(255, 255, 255, 0.1) !important;
    border-top: none !important;
    border-radius: 0 0 8px 8px !important;
  }
</style>
