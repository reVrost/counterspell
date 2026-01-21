<script lang="ts">
  import { taskStore } from "$lib/stores/tasks.svelte";
  import { tasksAPI } from "$lib/api";
  import { createTaskSSE } from "$lib/utils/sse";
  import type { PageData } from "./$types";
  import type { Project, Task, Message, LogEntry } from "$lib/types";
  import TaskDetail from "$lib/components/TaskDetail.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import ArrowLeftIcon from "@lucide/svelte/icons/arrow-left";
  import { onDestroy } from "svelte";

  interface Props {
    data: PageData;
  }

  let { data }: Props = $props();

  let task = $state<Task | null>(null);
  let project = $state<Project | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let agentContent = $state("");
  let diffContent = $state("");
  let logContent = $state<string[]>([]);
  let eventSource: EventSource | null = null;

  async function loadTask() {
    if (!data.taskId) return;

    loading = true;
    error = null;

    try {
      const taskData = await tasksAPI.get(data.taskId);
      agentContent = renderMessagesHTML(
        taskData.messages || [],
        taskData.status === "in_progress",
      );
      diffContent = taskData.gitDiff
        ? renderDiffHTML(taskData.gitDiff)
        : '<div class="text-gray-500 italic">No changes made</div>';
      logContent = taskData.logs?.map((log) => renderLogEntryHTML(log)) || [];

      taskStore.currentTask = task;

      // Set up SSE for real-time updates
      setupSSE(data.taskId);
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load task";
      console.error("Task load error:", err);
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
          task = { ...task, status: status as Task["status"] };
          taskStore.currentTask = task;
        }
      },
      onError: (err) => {
        console.error("Task SSE error:", err);
      },
    });
  }

  function renderMessagesHTML(
    messages: Message[],
    isInProgress: boolean,
  ): string {
    if (messages.length === 0) {
      return '<div class="p-5 text-gray-500 italic text-xs">No agent output</div>';
    }

    let html = '<div class="space-y-0">';
    for (const msg of messages) {
      html += renderMessageBubbleHTML(msg);
    }
    html += "</div>";

    if (isInProgress) {
      html += `
				<div class="flex items-center gap-3 px-4 py-3">
					<div class="relative">
						<div class="w-8 h-8 rounded-lg bg-violet-500/10 border border-violet-500/20 flex items-center justify-center">
							<i class="fas fa-robot text-sm text-violet-400 pulse-glow"></i>
						</div>
						<div class="absolute inset-0 animate-spin" style="animation-duration: 3s;">
							<div class="absolute -top-0.5 left-1/2 -translate-x-1/2 w-1 h-1 bg-violet-400 rounded-full"></div>
						</div>
					</div>
					<div>
						<p class="text-xs font-medium shimmer">Agent is thinking...</p>
						<p class="text-[10px] text-gray-600">Analyzing code</p>
					</div>
				</div>
			`;
    }

    return html;
  }

  function renderMessageBubbleHTML(msg: Message): string {
    const isUser = msg.role === "user";
    const bgClass = isUser
      ? "bg-violet-500/10 border-violet-500/20"
      : "bg-gray-800/50 border-gray-700/50";
    const icon = isUser ? "fa-user" : "fa-robot";
    const iconColor = isUser ? "text-violet-400" : "text-blue-400";

    let contentHtml = "";
    for (const block of msg.content) {
      if (block.type === "text" && block.text) {
        contentHtml += `<p class="text-sm text-gray-300 leading-normal">${escapeHtml(block.text)}</p>`;
      } else if (block.type === "tool_use" && block.toolName) {
        contentHtml += `
					<div class="flex items-center gap-2 my-2">
						<span class="px-2 py-0.5 bg-blue-500/10 border border-blue-500/20 rounded text-[10px] font-mono text-blue-300">
							${escapeHtml(block.toolName)}
						</span>
					</div>
				`;
      } else if (block.type === "tool_result") {
        contentHtml += `<pre class="text-xs text-gray-400 font-mono whitespace-pre-wrap bg-gray-900/50 rounded p-2 my-2">${escapeHtml(block.text || "")}</pre>`;
      }
    }

    return `
			<div class="flex gap-3 px-4 py-3 ${bgClass} border-b border-white/5">
				<div class="w-8 h-8 rounded-full ${iconColor} bg-white/5 flex items-center justify-center shrink-0">
					<i class="fas ${icon} text-xs"></i>
				</div>
				<div class="flex-1 min-w-0">
					${contentHtml}
				</div>
			</div>
		`;
  }

  function renderDiffHTML(diff: string): string {
    if (!diff) return '<div class="text-gray-500 italic">No changes made</div>';

    let html = "";
    for (const line of diff.split("\n")) {
      const escapedLine = escapeHtml(line);
      if (line.startsWith("+")) {
        html += `<div class="px-3 py-1 bg-green-500/10 text-green-400 font-mono text-xs border-l-2 border-green-500/50">${escapedLine.substring(1)}</div>`;
      } else if (line.startsWith("-")) {
        html += `<div class="px-3 py-1 bg-red-500/10 text-red-400 font-mono text-xs border-l-2 border-red-500/50">${escapedLine.substring(1)}</div>`;
      } else if (line.startsWith("@@")) {
        html += `<div class="px-3 py-1 bg-gray-800 text-gray-500 font-mono text-xs">${escapedLine}</div>`;
      } else if (line.trim() !== "") {
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
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
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
  <title>{task?.title || "Task"} - Counterspell</title>
</svelte:head>

<div class="min-h-screen bg-background flex flex-col">
  <!-- Navigation Header -->
  <div
    class="px-4 py-2 border-b border-white/5 flex items-center gap-3 bg-popover shrink-0"
  >
    <a
      href="/dashboard"
      class="w-11 h-11 rounded-full hover:bg-white/5 flex items-center justify-center text-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500/50"
      aria-label="Go back to dashboard"
    >
      <ArrowLeftIcon class="w-5 h-5" />
    </a>
    <div>
      <span class="text-[10px] text-gray-600 font-mono"
        >#{task?.id ?? "..."}</span
      >
      <h2 class="text-sm font-bold text-gray-200 line-clamp-1">
        {task?.title ?? "Loading..."}
      </h2>
    </div>
  </div>

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
          <p class="text-sm text-red-400 mb-2">{error}</p>
          <button
            onclick={() => loadTask()}
            class="px-4 py-2 bg-violet-500/20 border border-violet-500/30 rounded-lg text-xs text-violet-300 hover:bg-violet-500/30 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    {:else if task && project}
      <TaskDetail {task} {project} {agentContent} {diffContent} {logContent} />
    {/if}
  </div>
</div>
