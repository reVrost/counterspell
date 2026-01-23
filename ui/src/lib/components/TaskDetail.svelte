<script lang="ts">
  import { goto } from "$app/navigation";
  import { appState } from "$lib/stores/app.svelte";
  import { taskStore } from "$lib/stores/tasks.svelte";
  import { tasksAPI } from "$lib/api";
  import { cn } from "$lib/utils";
  import {
    modalSlideUp,
    backdropFade,
    slide,
    DURATIONS,
  } from "$lib/utils/transitions";
  import type { Task } from "$lib/types";
  import ChatInput from "./ChatInput.svelte";
  import TodoIndicator from "./TodoIndicator.svelte";
  import ArrowLeftIcon from "@lucide/svelte/icons/arrow-left";
  import TrashIcon from "@lucide/svelte/icons/trash";
  import RotateCcwIcon from "@lucide/svelte/icons/rotate-ccw";
  import EraserIcon from "@lucide/svelte/icons/eraser";
  import GitMergeIcon from "@lucide/svelte/icons/git-merge";
  import MessageSquareIcon from "@lucide/svelte/icons/message-square";
  import GithubIcon from "@lucide/svelte/icons/github";
  import SparklesIcon from "@lucide/svelte/icons/sparkles";

  interface Props {
    task: Task;
    agentContent: string;
    diffContent: string;
    logContent: string[];
  }

  let { task, agentContent, diffContent, logContent }: Props = $props();

  let activeTab = $state<"agent" | "diff" | "activity">("agent");
  let confirmAction = $state<string | null>(null);

  function handleBack() {
    goto("/dashboard");
  }

  async function handleChatSubmit(message: string, modelId: string) {
    try {
      const response = await tasksAPI.chat(task.id, message, modelId);
      if (response.message) {
        appState.showToast(response.message, "success");
      }
    } catch (err) {
      console.error("Failed to send message:", err);
      appState.showToast(
        err instanceof Error ? err.message : "Failed to send message",
        "error",
      );
    }
  }

  async function handleAction(action: string) {
    confirmAction = null;
    try {
      if (action === "retry") {
        const response = await tasksAPI.retry(task.id);
        appState.showToast(response.message || "Task retry started", "success");
      } else if (action === "clear") {
        const response = await tasksAPI.clear(task.id);
        appState.showToast(response.message || "History cleared", "success");
      } else if (action === "pr") {
        const response = await tasksAPI.createPR(task.id);
        if (response.pr_url) {
          appState.showToast("Pull request created!", "success");
          window.open(response.pr_url, "_blank");
        } else {
          appState.showToast(
            response.message || "Pull request created",
            "success",
          );
        }
      } else if (action === "merge") {
        const response = await tasksAPI.merge(task.id);
        if (response.status === "conflict") {
          appState.showToast(
            "Merge has conflicts - resolve them to continue",
            "info",
          );
        } else {
          appState.showToast(response.message || "Changes merged", "success");
        }
      } else if (action === "review") {
        const response = await tasksAPI.chat(task.id, "Please review the changes made in this task", appState.activeModelId);
        if (response.message) {
          appState.showToast(response.message, "success");
        }
      } else if (action === "discard") {
        const response = await tasksAPI.discard(task.id);
        appState.showToast(response.message || "Task discarded", "success");
        goto("/dashboard");
      }
    } catch (err) {
      console.error(`Failed to ${action}:`, err);
      appState.showToast(
        err instanceof Error ? err.message : `Failed to ${action}`,
        "error",
      );
    }
  }
</script>

<div class="flex flex-col h-full">
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
    <button
      onclick={() => (confirmAction = "discard")}
      class="w-11 h-11 flex items-center justify-center text-gray-600 hover:text-red-400 transition focus:outline-none focus:ring-2 focus:ring-red-500/50 rounded-lg"
      aria-label="Discard task"
    >
      <TrashIcon class="w-5 h-5" />
    </button>
  </div>

  <!-- Tabs Container -->
  <div
    class="flex items-center justify-between p-2 bg-popover shrink-0 border-b border-white/5"
  >
    <div class="w-6"></div>
    <div class="flex bg-gray-900 rounded-lg p-0.5 border border-gray-700/50">
      {#each ["agent", "diff", "activity"] as tab}
        <button
          onclick={() => (activeTab = tab as typeof activeTab)}
          class={cn(
            "px-4 py-2 text-[11px] font-medium rounded-md transition-all focus:outline-none focus:ring-2 focus:ring-purple-500/50 relative overflow-hidden",
            activeTab === tab
              ? "bg-gray-800 text-white shadow"
              : "text-gray-500",
          )}
        >
          {#if activeTab === tab}
            <span
              class="absolute inset-0 bg-purple-500/10"
              style:transition="opacity 200ms ease-out"
            ></span>
          {/if}
          <span class="relative z-10">
            {tab === "agent" ? "Agent" : tab === "diff" ? "Diff" : "Log"}
          </span>
        </button>
      {/each}
    </div>
    <!-- Status Indicator -->
    <div class="w-6 flex justify-end pr-2">
      {#if task.status === "pending"}
        <div class="w-1.5 h-1.5 rounded-full bg-gray-400" title="Pending"></div>
      {:else if task.status === "in_progress"}
        <div
          class="w-1.5 h-1.5 rounded-full bg-orange-400"
          title="In Progress"
        ></div>
      {:else if task.status === "review"}
        <div class="w-1.5 h-1.5 rounded-full bg-blue-400" title="Review"></div>
      {:else if task.status === "done"}
        <div class="w-1.5 h-1.5 rounded-full bg-green-400" title="Done"></div>
      {:else if task.status === "failed"}
        <div class="w-1.5 h-1.5 rounded-full bg-red-400" title="Failed"></div>
      {/if}
    </div>
  </div>

  <!-- Floating Todo Indicator -->
  {#if taskStore.todos.length > 0}
    <TodoIndicator />
  {/if}

  <!-- Main Content Area -->
  <div
    class="flex-1 overflow-y-auto bg-[#0D1117] relative w-full"
    id="content-scroll"
  >
    <!-- Agent Tab -->
    {#if activeTab === "agent"}
      <div class="pb-32">
        <div id="agent-content" class="mt-1">
          {@html agentContent}
        </div>
      </div>
    {/if}

    <!-- Diff Tab -->
    {#if activeTab === "diff"}
      <div class="p-0 min-h-full pb-32">
        <div
          class="px-4 py-3 border-b border-gray-800 sticky top-0 bg-[#0D1117] z-10 flex justify-between"
        >
          <span class="text-sm text-gray-400 font-mono">changes</span>
          <span class="text-[10px] text-green-500 font-mono">git diff</span>
        </div>
        <div class="p-3 diff-container">
          {@html diffContent}
        </div>
      </div>
    {/if}

    <!-- Activity Tab -->
    {#if activeTab === "activity"}
      <div class="p-5 pb-32 space-y-6">
        <div
          class="relative border-l border-gray-800 ml-2 space-y-6"
          id="log-content"
        >
          {#if logContent.length === 0}
            <div class="ml-4 text-sm text-gray-500 italic">No activity yet</div>
          {/if}
          {#each logContent as logHtml}
            <div>
              {@html logHtml}
            </div>
          {/each}
        </div>
      </div>
    {/if}
  </div>

  <!-- Chat Input Overlay - Always visible -->
  <div
    class="absolute bottom-20 inset-x-0 z-20 pb-6 px-3"
  >
    <div
      class="absolute inset-0 bg-gradient-to-t from-[#0D1117] via-[#0D1117]/95 to-transparent pointer-events-none"
    ></div>
    <div class="relative mx-auto max-w-xl">
      <ChatInput
        mode="chat"
        taskId={task.id}
        placeholder="Continue the conversation..."
        onSubmit={handleChatSubmit}
      />
    </div>
  </div>

  <!-- Bottom Actions Toolbar - Navigator Style -->
  <div
    class="shrink-0 px-4 py-3 border-t border-white/[0.06] bg-popover/95 backdrop-blur-sm pb-6"
  >
    <div class="flex items-center justify-center gap-2">
      <div class="flex-1 flex items-center justify-center">
        <div
          class="inline-flex items-center gap-1 bg-[#1a1a1a] rounded-full px-1 border border-white/[0.06]"
        >
          <!-- Chat -->
          <button
            type="button"
            onclick={() => {}}
            class="relative z-10 w-14 h-14 rounded-full flex items-center justify-center transition-all duration-200 text-gray-400 hover:text-gray-300 hover:bg-white/[0.04]"
            aria-label="Chat"
          >
            <MessageSquareIcon class="w-6 h-6" strokeWidth={2} />
          </button>

          <!-- Merge to Main -->
          <button
            type="button"
            onclick={() => (confirmAction = "merge")}
            class="relative z-10 w-14 h-14 rounded-full flex items-center justify-center transition-all duration-200 text-gray-400 hover:text-gray-300 hover:bg-white/[0.04]"
            aria-label="Merge to main"
          >
            <GitMergeIcon class="w-6 h-6" strokeWidth={2} />
          </button>

          <!-- Create PR -->
          <button
            type="button"
            onclick={() => (confirmAction = "pr")}
            class="relative z-10 w-14 h-14 rounded-full flex items-center justify-center transition-all duration-200 text-gray-400 hover:text-gray-300 hover:bg-white/[0.04]"
            aria-label="Create pull request"
          >
            <GithubIcon class="w-6 h-6" strokeWidth={2} />
          </button>

          <!-- Request Review -->
          <button
            type="button"
            onclick={() => (confirmAction = "review")}
            class="relative z-10 w-14 h-14 rounded-full flex items-center justify-center transition-all duration-200 text-gray-400 hover:text-gray-300 hover:bg-white/[0.04]"
            aria-label="Request AI review"
          >
            <SparklesIcon class="w-6 h-6" strokeWidth={2} />
          </button>
        </div>
      </div>
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
      onkeydown={(e) => e.key === "Escape" && (confirmAction = null)}
      aria-label="Close confirmation dialog"
    >
      <div
        transition:modalSlideUp|local={{ duration: DURATIONS.quick }}
        class="bg-popover border border-gray-700/50 rounded-xl p-5 w-[320px] shadow-2xl"
      >
        {#if confirmAction === "retry"}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div
                class="w-10 h-10 rounded-full bg-amber-500/10 flex items-center justify-center"
              >
                <RotateCcwIcon class="w-5 h-5 text-amber-500" />
              </div>
              <div>
                <h3 class="font-semibold text-white">Retry Task</h3>
                <p class="text-xs text-gray-400">Re-run with the same prompt</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will retry the previous prompt and overwrite any existing
              changes.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction("retry")}
                class="flex-1 h-9 rounded-lg bg-amber-600 hover:bg-amber-500 text-white text-sm font-medium transition-colors"
              >
                Retry
              </button>
            </div>
          </div>
        {:else if confirmAction === "clear"}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div
                class="w-10 h-10 rounded-full bg-red-500/10 flex items-center justify-center"
              >
                <EraserIcon class="w-5 h-5 text-red-500" />
              </div>
              <div>
                <h3 class="font-semibold text-white">Clear History</h3>
                <p class="text-xs text-gray-400">Reset memory and context</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will clear the chat history and agent output. The task will
              start fresh without any prior context.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction("clear")}
                class="flex-1 h-9 rounded-lg bg-red-600 hover:bg-red-500 text-white text-sm font-medium transition-colors"
              >
                Clear
              </button>
            </div>
          </div>
        {:else if confirmAction === "pr"}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div
                class="w-10 h-10 rounded-full bg-purple-500/10 flex items-center justify-center"
              >
                <GithubIcon class="w-5 h-5 text-purple-500" />
              </div>
              <div>
                <h3 class="font-semibold text-white">Create Pull Request</h3>
                <p class="text-xs text-gray-400">Push changes to GitHub</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will create a new pull request on GitHub with all the changes
              from this task.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction("pr")}
                class="flex-1 h-9 rounded-lg bg-purple-600 hover:bg-purple-500 text-white text-sm font-medium transition-colors"
              >
                Create PR
              </button>
            </div>
          </div>
        {:else if confirmAction === "merge"}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div
                class="w-10 h-10 rounded-full bg-green-500/10 flex items-center justify-center"
              >
                <GitMergeIcon class="w-5 h-5 text-green-500" />
              </div>
              <div>
                <h3 class="font-semibold text-white">Merge to Main</h3>
                <p class="text-xs text-gray-400">Apply changes directly</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will merge all changes directly into the main branch without
              creating a pull request.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction("merge")}
                class="flex-1 h-9 rounded-lg bg-green-600 hover:bg-green-500 text-white text-sm font-medium transition-colors"
              >
                Merge
              </button>
            </div>
          </div>
        {:else if confirmAction === "review"}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div
                class="w-10 h-10 rounded-full bg-purple-500/10 flex items-center justify-center"
              >
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
                onclick={() => handleAction("review")}
                class="flex-1 h-9 rounded-lg bg-purple-600 hover:bg-purple-500 text-white text-sm font-medium transition-colors"
              >
                Request Review
              </button>
            </div>
          </div>
        {:else if confirmAction === "discard"}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div
                class="w-10 h-10 rounded-full bg-red-500/10 flex items-center justify-center"
              >
                <TrashIcon class="w-5 h-5 text-red-500" />
              </div>
              <div>
                <h3 class="font-semibold text-white">Discard Task</h3>
                <p class="text-xs text-gray-400">Permanently delete task</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will permanently delete this task and all its data. This
              action cannot be undone.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction("discard")}
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
