<script lang="ts">
  import { goto } from '$app/navigation';
  import { appState } from '$lib/stores/app.svelte';
  import { taskStore } from '$lib/stores/tasks.svelte';
  import { tasksAPI } from '$lib/api';
  import { cn } from '$lib/utils';
  import { modalSlideUp, backdropFade, slide, DURATIONS } from '$lib/utils/transitions';
  import type { Task } from '$lib/types';
  import ChatInput from './ChatInput.svelte';
  import TodoIndicator from './TodoIndicator.svelte';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
  import TrashIcon from '@lucide/svelte/icons/trash';
  import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';
  import EraserIcon from '@lucide/svelte/icons/eraser';
  import GitMergeIcon from '@lucide/svelte/icons/git-merge';
  import MessageSquareIcon from '@lucide/svelte/icons/message-square';
  import GithubIcon from '@lucide/svelte/icons/github';
  import SparklesIcon from '@lucide/svelte/icons/sparkles';

  interface Props {
    task: Task;
    messages: Message[];
    diffContent: string;
    logContent: string[];
    isInProgress?: boolean;
  }

  let { task, messages, diffContent, logContent, isInProgress }: Props = $props();

  // Group messages for rendering (e.g., grouping tool/result into thinking blocks)
  type DisplayItem =
    | { type: 'message'; message: Message }
    | { type: 'thinking'; items: { command: string; result: string }[] };

  let displayItems = $derived.by(() => {
    const items: DisplayItem[] = [];
    let i = 0;
    while (i < messages.length) {
      const msg = messages[i];
      if (msg.role === 'tool' || msg.role === 'tool_result') {
        const thinkingItems: { command: string; result: string }[] = [];
        while (
          i < messages.length &&
          (messages[i].role === 'tool' || messages[i].role === 'tool_result')
        ) {
          const toolMsg = messages[i];
          if (toolMsg.role === 'tool') {
            let result = '';
            if (i + 1 < messages.length && messages[i + 1].role === 'tool_result') {
              result = messages[i + 1].content;
              i++;
            }
            thinkingItems.push({ command: toolMsg.content, result });
          } else {
            thinkingItems.push({ command: 'Command Trace', result: toolMsg.content });
          }
          i++;
        }
        items.push({ type: 'thinking', items: thinkingItems });
      } else {
        items.push({ type: 'message', message: msg });
        i++;
      }
    }
    return items;
  });

  function escapeHtml(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  let activeTab = $state<'agent' | 'diff'>('agent');
  let confirmAction = $state<string | null>(null);

  function handleBack() {
    goto('/dashboard');
  }

  async function handleChatSubmit(message: string, modelId: string) {
    try {
      const response = await tasksAPI.chat(task.id, message, modelId);
      if (response.message) {
        appState.showToast(response.message, 'success');
      }
    } catch (err) {
      console.error('Failed to send message:', err);
      appState.showToast(err instanceof Error ? err.message : 'Failed to send message', 'error');
    }
  }

  async function handleAction(action: string) {
    confirmAction = null;
    try {
      if (action === 'retry') {
        const response = await tasksAPI.retry(task.id);
        appState.showToast(response.message || 'Task retry started', 'success');
      } else if (action === 'clear') {
        const response = await tasksAPI.clear(task.id);
        appState.showToast(response.message || 'History cleared', 'success');
      } else if (action === 'pr') {
        const response = await tasksAPI.createPR(task.id);
        if (response.pr_url) {
          appState.showToast('Pull request created!', 'success');
          window.open(response.pr_url, '_blank');
        } else {
          appState.showToast(response.message || 'Pull request created', 'success');
        }
      } else if (action === 'merge') {
        const response = await tasksAPI.merge(task.id);
        if (response.status === 'conflict') {
          appState.showToast('Merge has conflicts - resolve them to continue', 'info');
        } else {
          appState.showToast(response.message || 'Changes merged', 'success');
        }
      } else if (action === 'review') {
        const response = await tasksAPI.chat(
          task.id,
          'Please review the changes made in this task',
          appState.activeModelId
        );
        if (response.message) {
          appState.showToast(response.message, 'success');
        }
      } else if (action === 'discard') {
        const response = await tasksAPI.discard(task.id);
        appState.showToast(response.message || 'Task discarded', 'success');
        goto('/dashboard');
      }
    } catch (err) {
      console.error(`Failed to ${action}:`, err);
      appState.showToast(err instanceof Error ? err.message : `Failed to ${action}`, 'error');
    }
  }
</script>

<div class="flex flex-col h-screen">
  <!-- Modal Header -->
  <div
    class="px-4 py-2 border-b border-white/5 flex items-center justify-between shrink-0 bg-popover"
  >
    <div class="flex items-center gap-3">
      <button
        onclick={handleBack}
        class="w-11 h-11 rounded-full hover:bg-white/5 flex items-center justify-center text-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500/50"
        aria-label="Go back"
      >
        <ArrowLeftIcon class="w-5 h-5" />
      </button>
      <div>
        <div class="flex items-center gap-2">
          <span class="text-gray-500 text-[10px]">
            <i class="fas fa-folder"></i>
            {task.repository_name || 'Unknown'}
          </span>
          <span class="text-[10px] text-gray-600 font-mono">#{task.id}</span>
        </div>
        <h2 class="text-sm font-bold text-gray-200 line-clamp-1 w-48">
          {task.title}
        </h2>
      </div>
    </div>

    <!-- Status Badge -->
    <div
      class="flex items-center px-2 py-1 rounded-full bg-white/5 border border-white/10 gap-1.5 h-7"
    >
      {#if task.status === 'pending'}
        <div class="w-1.5 h-1.5 rounded-full bg-gray-400"></div>
        <span class="text-[9px] uppercase font-bold tracking-wider text-gray-400">Pending</span>
      {:else if task.status === 'in_progress'}
        <div class="w-1.5 h-1.5 rounded-full bg-orange-400 pulse-glow"></div>
        <span class="text-[9px] uppercase font-bold tracking-wider text-orange-400">Running</span>
      {:else if task.status === 'review'}
        <div class="w-1.5 h-1.5 rounded-full bg-blue-400 pulse-glow"></div>
        <span class="text-[9px] uppercase font-bold tracking-wider text-blue-400">Ready</span>
      {:else if task.status === 'done'}
        <div class="w-1.5 h-1.5 rounded-full bg-green-400"></div>
        <span class="text-[9px] uppercase font-bold tracking-wider text-green-400">Merged</span>
      {:else if task.status === 'failed'}
        <div class="w-1.5 h-1.5 rounded-full bg-red-400"></div>
        <span class="text-[9px] uppercase font-bold tracking-wider text-red-400">Failed</span>
      {/if}
    </div>
  </div>

  <!-- Tabs Container -->
  <div
    class="flex items-center justify-between p-2 px-7 bg-popover shrink-0 border-b border-white/5"
  >
    <div class="flex bg-gray-900 rounded-lg p-0.5 border border-gray-700/50">
      {#each ['agent', 'diff'] as tab}
        <button
          onclick={() => (activeTab = tab as typeof activeTab)}
          class={cn(
            'px-4 py-2 text-[11px] font-medium rounded-md transition-all focus:outline-none focus:ring-2 focus:ring-purple-500/50 relative overflow-hidden',
            activeTab === tab ? 'bg-gray-800 text-white shadow' : 'text-gray-500'
          )}
        >
          {#if activeTab === tab}
            <span
              class="absolute inset-0 bg-purple-500/10"
              style:transition="opacity 200ms ease-out"
            ></span>
          {/if}
          <span class="relative z-10">
            {tab === 'agent' ? 'Agent' : tab === 'diff' ? 'Diff' : 'Log'}
          </span>
        </button>
      {/each}
    </div>

    <!-- Status Indicator -->

    <div class="flex items-center justify-end gap-2">
      <button
        onclick={() => (confirmAction = 'discard')}
        class="w-8 h-8 flex items-center justify-center text-gray-500 hover:text-red-400 transition focus:outline-none rounded-lg"
        aria-label="Discard task"
      >
        <TrashIcon class="w-4 h-4" />
      </button>
      <!-- Action Buttons -->
      {#if task.status !== 'in_progress' && task.status !== 'done'}
        <div class="flex items-center gap-2">
          <button
            onclick={() => (confirmAction = 'merge')}
            class="h-8 px-3 rounded-md bg-white/5 hover:bg-white/10 border border-white/10 text-[11px] font-medium text-gray-300 transition-all flex items-center gap-1.5"
            title="Merge directly to main"
          >
            <GithubIcon class="w-3.5 h-3.5" />
            <span class="hidden sm:inline">Merge</span>
          </button>

          <button
            onclick={() => (confirmAction = 'pr')}
            class="h-8 px-3 rounded-md bg-white hover:bg-gray-100 text-black text-[11px] font-bold transition-all shadow-sm flex items-center gap-1.5"
            title="Create a Pull Request"
          >
            <GitMergeIcon class="w-3.5 h-3.5" />
            <span>Merge</span>
          </button>
        </div>
      {/if}
    </div>
  </div>

  <!-- Floating Todo Indicator -->
  {#if taskStore.todos.length > 0}
    <TodoIndicator />
  {/if}

  <!-- Main Content Area -->
  <div class="flex-1 overflow-y-auto relative w-full h-screen mb-10" id="content-scroll">
    <!-- Agent Tab -->
    {#if activeTab === 'agent'}
      <div class="pb-32">
        <div id="agent-content" class="mt-1 space-y-1">
          {#each displayItems as item}
            {#if item.type === 'message'}
              {@render messageSnippet(item.message)}
            {:else if item.type === 'thinking'}
              {@render thinkingSnippet(item.items)}
            {/if}
          {/each}

          {#if isInProgress}
            <div class="flex items-center gap-3 px-12 py-3">
              <div class="relative shrink-0">
                <div
                  class="w-8 h-8 rounded-lg bg-violet-500/10 border border-violet-500/20 flex items-center justify-center"
                >
                  <i class="fas fa-robot text-base text-violet-400 pulse-glow"></i>
                </div>
                <div class="absolute inset-0 animate-spin" style="animation-duration: 3s;">
                  <div
                    class="absolute -top-0.5 left-1/2 -translate-x-1/2 w-1 h-1 bg-violet-400 rounded-full"
                  ></div>
                </div>
              </div>
              <div>
                <p class="text-sm font-medium shimmer text-gray-300">Agent is thinking...</p>
                <p class="text-sm text-gray-600">Analyzing context</p>
              </div>
            </div>
          {/if}

          {#if messages.length === 0}
            <div class="p-5 text-gray-500 italic text-xs">No agent output</div>
          {/if}
        </div>
      </div>
    {/if}

    {#snippet messageSnippet(msg: Message)}
      {#if msg.role === 'user'}
        <div class="flex gap-4 px-4 py-2 items-start">
          <div
            class="w-8 h-8 rounded-full bg-white/10 flex items-center justify-center shrink-0 overflow-hidden border border-white/5"
          >
            <img
              src="https://api.dicebear.com/7.x/avataaars/svg?seed=user&backgroundColor=b6e3f4"
              alt="User"
              class="w-full h-full"
            />
          </div>
          <div
            class="flex-1 min-w-0 bg-[#1e1e1e]/60 border border-white/10 rounded-xl px-4 py-3 text-white shadow-lg"
          >
            <p class="text-base leading-relaxed">{msg.content}</p>
          </div>
        </div>
      {:else}
        <div class="px-12 py-2 pr-4">
          <p class="text-base text-gray-200 leading-relaxed font-sans">{msg.content}</p>
        </div>
      {/if}
    {/snippet}

    {#snippet thinkingSnippet(items: { command: string; result: string }[])}
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
          {#each items as item}
            <div
              class="bg-[#0a0a0a] border border-white/5 rounded-lg overflow-hidden font-mono shadow-xl"
            >
              <div
                class="flex items-center justify-between px-3 py-1.5 bg-white/5 border-b border-white/5"
              >
                <div class="flex items-center gap-2 text-gray-400">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="12"
                    height="12"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2.5"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    class="text-gray-500"
                    ><polyline points="4 17 10 11 4 5" /><line
                      x1="12"
                      y1="19"
                      x2="20"
                      y2="19"
                    /></svg
                  >
                  <span class="text-xs font-bold text-gray-500 tracking-tight">{item.command}</span>
                </div>
                <div class="text-gray-700 hover:text-gray-500 transition-colors">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="12"
                    height="12"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2.5"
                    stroke-linecap="round"
                    stroke-linejoin="round"><path d="M18 6 6 18" /><path d="m6 6 12 12" /></svg
                  >
                </div>
              </div>
              {#if item.result}
                <div class="p-3 text-[11px] text-gray-400 overflow-x-auto max-h-[400px]">
                  <pre class="whitespace-pre-wrap leading-tight antialiased">{item.result}</pre>
                </div>
              {/if}
            </div>
          {/each}
        </div>
      </details>
    {/snippet}

    <!-- Diff Tab -->
    {#if activeTab === 'diff'}
      <div class="p-0 min-h-full pb-32">
        <div
          class="px-4 py-3 border-b border-gray-800 sticky top-0 bg-[#0D1117] z-10 flex justify-between"
        >
          <span class="text-sm text-gray-400 font-mono">changes</span>
          <span class="text-xs text-green-500 font-mono">git diff</span>
        </div>
        <div class="p-3 diff-container">
          {@html diffContent}
        </div>
      </div>
    {/if}
  </div>

  <!-- Chat Input Overlay - Always visible -->
  <div class="absolute bottom-0 inset-x-0 z-20 pb-6 px-3">
    <div
      class="absolute inset-0 bg-gradient-to-t from-[#0D1117] via-[#0D1117]/95 to-transparent pointer-events-none"
    ></div>
    <div class="relative mx-auto max-w-4xl">
      <ChatInput
        mode="chat"
        taskId={task.id}
        placeholder="Continue the conversation..."
        onSubmit={handleChatSubmit}
      />
    </div>
  </div>

  <!-- Confirmation Modal -->
  {#if confirmAction}
    <div
      transition:backdropFade|local={{ duration: DURATIONS.normal }}
      class="fixed inset-0 z-[200] flex items-start justify-center pt-[25vh] bg-black/60 backdrop-blur-sm"
      role="button"
      tabindex="-1"
      onclick={(e) => e.target === e.currentTarget && (confirmAction = null)}
      onkeydown={(e) => e.key === 'Escape' && (confirmAction = null)}
      aria-label="Close confirmation dialog"
    >
      <div
        transition:modalSlideUp|local={{ duration: DURATIONS.quick }}
        class="bg-popover border border-gray-700/50 rounded-xl p-5 w-[320px] shadow-2xl"
      >
        {#if confirmAction === 'retry'}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-amber-500/10 flex items-center justify-center">
                <RotateCcwIcon class="w-5 h-5 text-amber-500" />
              </div>
              <div>
                <h3 class="font-semibold text-white">Retry Task</h3>
                <p class="text-xs text-gray-400">Re-run with the same prompt</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will retry the previous prompt and overwrite any existing changes.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction('retry')}
                class="flex-1 h-9 rounded-lg bg-amber-600 hover:bg-amber-500 text-white text-sm font-medium transition-colors"
              >
                Retry
              </button>
            </div>
          </div>
        {:else if confirmAction === 'clear'}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-red-500/10 flex items-center justify-center">
                <EraserIcon class="w-5 h-5 text-red-500" />
              </div>
              <div>
                <h3 class="font-semibold text-white">Clear History</h3>
                <p class="text-xs text-gray-400">Reset memory and context</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will clear the chat history and agent output. The task will start fresh without
              any prior context.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction('clear')}
                class="flex-1 h-9 rounded-lg bg-red-600 hover:bg-red-500 text-white text-sm font-medium transition-colors"
              >
                Clear
              </button>
            </div>
          </div>
        {:else if confirmAction === 'pr'}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-purple-500/10 flex items-center justify-center">
                <GithubIcon class="w-5 h-5 text-purple-500" />
              </div>
              <div>
                <h3 class="font-semibold text-white">Create Pull Request</h3>
                <p class="text-xs text-gray-400">Push changes to GitHub</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will create a new pull request on GitHub with all the changes from this task.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction('pr')}
                class="flex-1 h-9 rounded-lg bg-purple-600 hover:bg-purple-500 text-white text-sm font-medium transition-colors"
              >
                Create PR
              </button>
            </div>
          </div>
        {:else if confirmAction === 'merge'}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-green-500/10 flex items-center justify-center">
                <GitMergeIcon class="w-5 h-5 text-green-500" />
              </div>
              <div>
                <h3 class="font-semibold text-white">Merge to Main</h3>
                <p class="text-xs text-gray-400">Apply changes directly</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will merge all changes directly into the main branch without creating a pull
              request.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction('merge')}
                class="flex-1 h-9 rounded-lg bg-green-600 hover:bg-green-500 text-white text-sm font-medium transition-colors"
              >
                Merge
              </button>
            </div>
          </div>
        {:else if confirmAction === 'review'}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-purple-500/10 flex items-center justify-center">
                <SparklesIcon class="w-5 h-5 text-purple-500" />
              </div>
              <div>
                <h3 class="font-semibold text-white">Request AI Review</h3>
                <p class="text-xs text-gray-400">Get code review feedback</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will request an AI code review of the changes made in this task.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction('review')}
                class="flex-1 h-9 rounded-lg bg-purple-600 hover:bg-purple-500 text-white text-sm font-medium transition-colors"
              >
                Request Review
              </button>
            </div>
          </div>
        {:else if confirmAction === 'discard'}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-red-500/10 flex items-center justify-center">
                <TrashIcon class="w-5 h-5 text-red-500" />
              </div>
              <div>
                <h3 class="font-semibold text-white">Discard Task</h3>
                <p class="text-xs text-gray-400">Permanently delete task</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will permanently delete this task and all its data. This action cannot be undone.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction('discard')}
                class="flex-1 h-9 rounded-lg bg-red-600 hover:bg-red-500 text-white text-sm font-medium transition-colors"
              >
                Discard
              </button>
            </div>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>
