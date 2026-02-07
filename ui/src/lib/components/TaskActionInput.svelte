<script lang="ts">
  import { appState } from '$lib/stores/app.svelte';
  import { cn } from '$lib/utils';
  import { tasksAPI, filesAPI } from '$lib/api';
  import { dropdownPop, DURATIONS } from '$lib/utils/transitions';
  import ModelSelector from './ModelSelector.svelte';
  import GitMergeIcon from '@lucide/svelte/icons/git-merge';
  import GithubIcon from '@lucide/svelte/icons/github';
  import TrashIcon from '@lucide/svelte/icons/trash';
  import PaperclipIcon from '@lucide/svelte/icons/paperclip';
  import ArrowUpIcon from '@lucide/svelte/icons/arrow-up';
  import FileIcon from '@lucide/svelte/icons/file';
  import XIcon from '@lucide/svelte/icons/x';
  import SearchIcon from '@lucide/svelte/icons/search';

  interface Props {
    taskId: string;
    taskStatus: string;
    placeholder?: string;
    onSubmit?: (message: string, modelId: string) => void;
    onMerge?: () => void;
    onCreatePR?: () => void;
    onDelete?: () => void;
  }

  let {
    taskId,
    taskStatus,
    placeholder = 'Ask for changes or approve...',
    onSubmit,
    onMerge,
    onCreatePR,
    onDelete,
  }: Props = $props();

  let text = $state('');
  let inputRef = $state<HTMLTextAreaElement | null>(null);
  let showFileMenu = $state(false);
  let files = $state<string[]>([]);
  let selectedIndex = $state(0);

  const isDone = $derived(taskStatus === 'done');
  const isInProgress = $derived(taskStatus === 'in_progress');
  const canMerge = $derived(!isDone && !isInProgress && text.trim().length === 0);
  const hasText = $derived(text.trim().length > 0);

  function resize() {
    if (!inputRef) return;
    inputRef.style.height = 'auto';
    let newHeight = inputRef.scrollHeight;
    const maxHeight = window.innerHeight * 0.25;
    if (newHeight > maxHeight) newHeight = maxHeight;
    inputRef.style.height = `${newHeight}px`;
  }

  async function checkMention(e: KeyboardEvent) {
    const match = text.match(/@([^ ]*)$/);
    if (match) {
      showFileMenu = true;
      const query = match[1] || '';
      await searchFiles(query);
    } else {
      showFileMenu = false;
      files = [];
    }
    if (e.key === 'Escape') {
      showFileMenu = false;
      files = [];
    }
  }

  async function searchFiles(query: string) {
    const workspaceId = appState.activeWorkspaceId;
    if (!workspaceId) return;
    try {
      files = await filesAPI.search(workspaceId, query);
      selectedIndex = 0;
    } catch (e) {
      console.error('File search failed:', e);
      files = [];
    }
  }

  function handleFileNav(e: KeyboardEvent) {
    if (!showFileMenu || !files || files.length === 0) return;
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      selectedIndex = Math.min(selectedIndex + 1, files.length - 1);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      selectedIndex = Math.max(selectedIndex - 1, 0);
    } else if (e.key === 'Enter' || e.key === 'Tab') {
      e.preventDefault();
      if (files[selectedIndex]) {
        insertFile(files[selectedIndex]);
      }
    }
  }

  function insertFile(f: string) {
    text = text.replace(/@[^ ]*$/, '') + f + ' ';
    showFileMenu = false;
    files = [];
    selectedIndex = 0;
    inputRef?.focus();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault();
      showFileMenu = false;
      files = [];
      return;
    }

    if (
      showFileMenu &&
      files &&
      files.length > 0 &&
      ['ArrowUp', 'ArrowDown', 'Tab'].includes(e.key)
    ) {
      handleFileNav(e);
    } else if (showFileMenu && files && files.length > 0 && e.key === 'Enter' && !e.shiftKey) {
      handleFileNav(e);
    } else if (!showFileMenu && e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      if (hasText) {
        submit();
      } else if (canMerge && onMerge) {
        onMerge();
      }
    }
  }

  async function submit() {
    if (!text.trim()) return;
    const msg = text.trim();
    text = '';
    if (inputRef) inputRef.style.height = 'auto';
    if (onSubmit) {
      onSubmit(msg, appState.activeModelId);
    }
    navigator.vibrate?.(30);
  }

  function handleMergeClick() {
    if (canMerge && onMerge) {
      navigator.vibrate?.(50);
      onMerge();
    }
  }

  function handleSubmitClick() {
    if (hasText) {
      submit();
    } else if (canMerge) {
      handleMergeClick();
    }
  }
</script>

<div class="relative w-full">
  <!-- Main Container -->
  <div
    class="bg-[#0C0E12] border border-white/10 rounded-[24px] shadow-2xl relative transition-all duration-200 ring-1 ring-white/5 flex flex-col group focus-within:border-white/20 focus-within:ring-white/10 overflow-hidden"
  >
    <!-- Input Area -->
    <div class="relative px-4 pt-4 pb-2">
      <!-- File Menu Popover -->
      {#if showFileMenu && files && files.length > 0}
        <div
          transition:dropdownPop|local
          class="absolute bottom-full left-0 mb-2 w-80 bg-popover border border-gray-700 rounded-xl shadow-2xl overflow-hidden max-h-58 overflow-y-auto z-50"
        >
          <div
            class="px-3 py-2 text-[10px] text-gray-500 font-bold uppercase tracking-wider border-b border-gray-800 flex items-center justify-between"
          >
            <span>Files</span>
            <span class="text-gray-600 font-normal normal-case"
              >↑↓ to navigate, Enter to select</span
            >
          </div>
          {#each files as file, idx}
            <button
              type="button"
              onclick={() => insertFile(file)}
              onmouseenter={() => (selectedIndex = idx)}
              class={cn(
                'w-full px-3 py-2 text-sm font-mono cursor-pointer transition flex items-center gap-2 text-left',
                idx === selectedIndex
                  ? 'bg-primary/20 text-primary'
                  : 'text-gray-300 hover:bg-white/5'
              )}
            >
              <FileIcon class="w-2.5 h-2.5 opacity-40" />
              <span class="truncate">{file}</span>
            </button>
          {/each}
        </div>
      {/if}

      {#if showFileMenu && files && files.length === 0}
        <div
          transition:dropdownPop|local
          class="absolute bottom-full left-0 mb-2 w-64 bg-popover border border-gray-700 rounded-xl shadow-2xl overflow-hidden z-50"
        >
          <div class="px-3 py-4 text-sm text-gray-500 text-center">
            <SearchIcon class="w-5 h-5 mx-auto mb-2 opacity-50" />
            <div>No files found</div>
          </div>
        </div>
      {/if}

      <textarea
        bind:this={inputRef}
        bind:value={text}
        oninput={resize}
        onkeyup={checkMention}
        onkeydown={handleKeydown}
        rows="1"
        {placeholder}
        aria-label={placeholder}
        disabled={isDone}
        class="bg-transparent border-none focus:ring-0 focus:outline-none text-white text-base placeholder-gray-500 w-full resize-none font-medium p-0 leading-relaxed max-h-[25vh] min-h-[24px] disabled:opacity-50 disabled:cursor-not-allowed"
      ></textarea>
    </div>

    <!-- Actions Bar -->
    <div class="flex items-center justify-between px-2 pb-2 mt-1">
      <!-- Left Side: Secondary Actions -->
      <div class="flex items-center gap-1">
        {#if !isDone && !isInProgress}
          <!-- Create PR Button -->
          <button
            type="button"
            onclick={onCreatePR}
            aria-label="Create Pull Request"
            class="h-9 px-3 flex items-center gap-2 rounded-xl text-gray-400 hover:text-white hover:bg-white/5 transition-all duration-150 text-xs font-medium"
          >
            <GithubIcon class="w-4 h-4" />
            <span class="hidden sm:inline">PR</span>
          </button>

          <!-- Delete Button -->
          <button
            type="button"
            onclick={onDelete}
            aria-label="Delete task"
            class="w-9 h-9 flex items-center justify-center rounded-xl text-gray-400 hover:text-red-400 hover:bg-red-500/10 transition-all duration-150"
          >
            <TrashIcon class="w-4 h-4" />
          </button>
        {/if}

        <!-- Model Selector -->
        {#if !isDone}
          <ModelSelector position="left" />
        {/if}
      </div>

      <!-- Right Side: Primary Action -->
      <div class="flex items-center gap-2">
        <!-- Merge Button (Primary CTA) -->
        {#if !isDone && !isInProgress}
          <button
            type="button"
            onclick={handleMergeClick}
            disabled={!canMerge}
            class={cn(
              'h-10 px-4 rounded-xl flex items-center gap-2 text-sm font-semibold transition-all duration-200',
              canMerge
                ? 'bg-green-600 hover:bg-green-500 text-white shadow-lg shadow-green-500/20 scale-100'
                : 'bg-gray-800 text-gray-500 cursor-not-allowed'
            )}
          >
            <GitMergeIcon class="w-4 h-4" />
            <span>Merge</span>
          </button>
        {/if}

        <!-- Send Button (when typing) -->
        {#if hasText && !isDone}
          <button
            type="button"
            onclick={handleSubmitClick}
            aria-label="Send message"
            class="w-10 h-10 rounded-xl flex items-center justify-center bg-violet-600 hover:bg-violet-500 text-white transition-all duration-150 shadow-lg shadow-violet-500/20"
          >
            <ArrowUpIcon class="w-5 h-5" />
          </button>
        {/if}

        {#if isDone}
          <div class="flex items-center gap-2 px-3 py-2 text-green-400 text-sm font-medium">
            <div class="w-2 h-2 rounded-full bg-green-400"></div>
            <span>Merged</span>
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>
