<script lang="ts">
  import { taskStore } from '$lib/stores/tasks.svelte';
  import { tasksAPI } from '$lib/api';
  import { createTaskSSE } from '$lib/utils/sse';
  import type { PageData } from './$types';
  import type { TaskResponse, Message } from '$lib/types';
  import TaskDetail from '$lib/components/TaskDetail.svelte';
  import Skeleton from '$lib/components/Skeleton.svelte';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
  import { onDestroy } from 'svelte';

  interface Props {
    data: PageData;
  }

  let { data }: Props = $props();

  let task = $state<TaskResponse | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let agentContent = $state('');
  let diffContent = $state('');
  let logContent = $state<string[]>([]);
  let eventSource: EventSource | null = null;

  async function loadTask() {
    if (!data.taskId) return;

    loading = true;
    error = null;

    try {
      const taskData = await tasksAPI.get(data.taskId);
      task = taskData;
      agentContent = renderMessagesHTML(
        taskData.messages || [],
        taskData.task.status === 'in_progress'
      );
      diffContent = taskData.git_diff
        ? renderDiffHTML(taskData.git_diff)
        : '<div class="text-gray-500 italic">No changes made</div>';
      logContent = [];

      taskStore.currentTask = taskData.task;

      setupSSE(data.taskId);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load task';
      console.error('Task load error:', err);
    } finally {
      loading = false;
    }
  }

  function setupSSE(taskId: string) {
    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }

    eventSource = createTaskSSE(taskId, {
      onAgentUpdate: (html: string) => {
        agentContent = html;
      },
      onDiffUpdate: (html: string) => {
        diffContent = html;
      },
      onLog: (html: string) => {
        logContent = [...logContent, html];
      },
      onStatus: (html: string) => {},
      onComplete: (status: string) => {
        if (task) {
          task = { ...task, task: { ...task.task, status: status as Task['status'] } };
          taskStore.currentTask = task.task;
        }
      },
      onError: (err) => {
        console.error('Task SSE error:', err);
      },
    });
  }

  function renderMessagesHTML(messages: Message[], isInProgress: boolean): string {
    if (messages.length === 0) {
      return '<div class="p-5 text-gray-500 italic text-xs">No agent output</div>';
    }

    let html = '<div class="space-y-4 py-4">';
    let i = 0;
    while (i < messages.length) {
      const msg = messages[i];

      if (msg.role === 'tool' || msg.role === 'tool_result') {
        // Start of a potential thinking block
        html += `
          <details class="mx-12 my-4 group" open>
            <summary class="flex items-center gap-2 cursor-pointer text-gray-500 hover:text-gray-300 transition-colors list-none outline-none">
              <div class="w-4 h-4 flex items-center justify-center group-open:rotate-90 transition-transform">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m9 18 6-6-6-6"/></svg>
              </div>
              <span class="text-[10px] font-bold tracking-widest uppercase">Thinking</span>
            </summary>
            <div class="mt-3 space-y-3">`;

        while (
          i < messages.length &&
          (messages[i].role === 'tool' || messages[i].role === 'tool_result')
        ) {
          const toolMsg = messages[i];
          if (toolMsg.role === 'tool') {
            // Look ahead for its result
            let result = '';
            if (i + 1 < messages.length && messages[i + 1].role === 'tool_result') {
              result = messages[i + 1].content;
              i++; // Skip result in next iteration
            }
            html += renderToolBlockHTML(toolMsg.content, result);
          } else {
            // Dangling tool_result
            html += renderToolBlockHTML('Command Trace', toolMsg.content);
          }
          i++;
        }

        html += `</div></details>`;
      } else {
        html += renderMessageBubbleHTML(msg);
        i++;
      }
    }
    html += '</div>';

    if (isInProgress) {
      html += `
				<div class="flex items-center gap-3 px-12 py-3">
					<div class="relative shrink-0">
						<div class="w-8 h-8 rounded-lg bg-violet-500/10 border border-violet-500/20 flex items-center justify-center">
							<i class="fas fa-robot text-base text-violet-400 pulse-glow"></i>
						</div>
						<div class="absolute inset-0 animate-spin" style="animation-duration: 3s;">
							<div class="absolute -top-0.5 left-1/2 -translate-x-1/2 w-1 h-1 bg-violet-400 rounded-full"></div>
						</div>
					</div>
					<div>
						<p class="text-xs font-medium shimmer text-gray-300">Agent is thinking...</p>
						<p class="text-[10px] text-gray-600">Analyzing context</p>
					</div>
				</div>
			`;
    }

    return html;
  }

  function renderMessageBubbleHTML(msg: Message): string {
    const isUser = msg.role === 'user';

    if (isUser) {
      return `
        <div class="flex gap-4 px-4 py-2 items-start">
          <div class="w-8 h-8 rounded-full bg-white/10 flex items-center justify-center shrink-0 overflow-hidden border border-white/5">
            <img src="https://api.dicebear.com/7.x/avataaars/svg?seed=user&backgroundColor=b6e3f4" alt="User" class="w-full h-full" />
          </div>
          <div class="flex-1 min-w-0 bg-[#1e1e1e]/60 border border-white/10 rounded-xl px-4 py-3 text-white shadow-lg">
            <p class="text-base leading-relaxed">${escapeHtml(msg.content)}</p>
          </div>
        </div>
      `;
    }

    // Default: Assistant/Agent text
    return `
      <div class="px-12 py-2 pr-4">
        <p class="text-base text-gray-200 leading-relaxed font-sans">${escapeHtml(msg.content)}</p>
      </div>
    `;
  }

  function renderToolBlockHTML(command: string, result: string): string {
    return `
      <div class="bg-[#0a0a0a] border border-white/5 rounded-lg overflow-hidden font-mono shadow-xl">
        <div class="flex items-center justify-between px-3 py-1.5 bg-white/5 border-b border-white/5">
          <div class="flex items-center gap-2 text-gray-400">
            <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="text-gray-500"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>
            <span class="text-[9px] font-bold text-gray-500 tracking-tight">${escapeHtml(command)}</span>
          </div>
          <div class="text-gray-700 hover:text-gray-500 transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
          </div>
        </div>
        ${
          result
            ? `
        <div class="p-3 text-[11px] text-gray-400 overflow-x-auto max-h-[400px]">
          <pre class="whitespace-pre-wrap leading-tight antialiased">${escapeHtml(result)}</pre>
        </div>
        `
            : ''
        }
      </div>
    `;
  }

  function renderDiffHTML(diff: string): string {
    if (!diff) return '<div class="text-gray-500 italic">No changes made</div>';

    let html = '';
    for (const line of diff.split('\n')) {
      const escapedLine = escapeHtml(line);
      if (line.startsWith('+')) {
        html += `<div class="px-3 py-1 bg-green-500/10 text-green-400 font-mono text-xs border-l-2 border-green-500/50">${escapedLine.substring(1)}</div>`;
      } else if (line.startsWith('-')) {
        html += `<div class="px-3 py-1 bg-red-500/10 text-red-400 font-mono text-xs border-l-2 border-red-500/50">${escapedLine.substring(1)}</div>`;
      } else if (line.startsWith('@@')) {
        html += `<div class="px-3 py-1 bg-gray-800 text-gray-500 font-mono text-xs">${escapedLine}</div>`;
      } else if (line.trim() !== '') {
        html += `<div class="px-3 py-1 text-gray-400 font-mono text-xs">${escapedLine}</div>`;
      }
    }
    return html;
  }

  function renderLogEntryHTML(log: LogEntry): string {
    return `
			<div class="ml-4 relative">
				<div class="absolute -left-[21px] top-1 h-2.5 w-2.5 rounded-full border border-[#0D1117] bg-blue-500"></div>
				<p class="text-xs text-gray-400">${escapeHtml(log.message)}</p>
			</div>
		`;
  }

  function escapeHtml(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  $effect(() => {
    loadTask();
  });

  onDestroy(() => {
    if (eventSource) {
      eventSource.close();
    }
  });
</script>

<svelte:head>
  <title>{task?.title || 'Task'} - Counterspell</title>
</svelte:head>

<div class="min-h-screen bg-background flex flex-col">
  <!-- Task Detail Content -->
  <div class="flex-1 overflow-hidden">
    {#if loading}
      <!-- Skeleton screens matching TaskDetail layout -->
      <div class="flex flex-col h-full">
        <!-- Tabs skeleton -->
        <div class="p-2 flex justify-between">
          <div class="w-6"></div>
          <div class="flex gap-2">
            {#each [1, 2, 3] as _}
              <Skeleton variant="rounded" class="h-8 w-16" />
            {/each}
          </div>
          <div class="w-6"></div>
        </div>

        <!-- Content skeleton -->
        <div class="flex-1 p-4 space-y-4">
          {#each [1, 2, 3, 4] as _}
            <div class="space-y-2">
              <Skeleton variant="text" class="w-full" />
              <Skeleton variant="text" class="w-5/6" />
              <Skeleton variant="text" class="w-3/4" />
            </div>
          {/each}
        </div>

        <!-- Bottom actions skeleton -->
        <div class="px-4 py-3 border-t border-white/5">
          <div class="flex gap-2">
            <Skeleton variant="rounded" class="flex-1 h-12" />
            <Skeleton variant="rounded" class="flex-1 h-12" />
          </div>
        </div>
      </div>
    {:else if error}
      <div class="flex items-center justify-center h-full">
        <div class="text-center">
          <p class="text-base text-red-400 mb-2">{error}</p>
          <button
            onclick={() => loadTask()}
            class="px-4 py-2 bg-violet-500/20 border border-violet-500/30 rounded-lg text-xs text-violet-300 hover:bg-violet-500/30 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    {:else if task}
      <TaskDetail task={task.task} {agentContent} {diffContent} {logContent} />
    {/if}
  </div>
</div>

<style>
  :global(summary::-webkit-details-marker) {
    display: none;
  }

  :global(.shimmer) {
    background: linear-gradient(
      90deg,
      rgba(255, 255, 255, 0.05) 25%,
      rgba(255, 255, 255, 0.1) 50%,
      rgba(255, 255, 255, 0.05) 75%
    );
    background-size: 200% 100%;
    animation: shimmer 2s infinite;
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  @keyframes shimmer {
    0% {
      background-position: 200% 0;
    }
    100% {
      background-position: -200% 0;
    }
  }

  :global(.pulse-glow) {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
      filter: drop-shadow(0 0 2px rgba(139, 92, 246, 0.5));
    }
    50% {
      opacity: 0.7;
      filter: drop-shadow(0 0 5px rgba(139, 92, 246, 0.8));
    }
  }
</style>
