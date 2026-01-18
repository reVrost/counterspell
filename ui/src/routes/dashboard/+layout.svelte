<script lang="ts">
	import Header from '$lib/components/Header.svelte';
	import Toast from '$lib/components/Toast.svelte';
	import ChatInput from '$lib/components/ChatInput.svelte';
	import TaskDetail from '$lib/components/TaskDetail.svelte';
	import { appState } from '$lib/stores/app.svelte';
	import { taskStore } from '$lib/stores/tasks.svelte';
	import { tasksAPI } from '$lib/api';
	import { createTaskSSE } from '$lib/utils/sse';
	import type { Project, Task, Message, LogEntry } from '$lib/types';
	import { onDestroy, tick } from 'svelte';

	let { children } = $props();

	// Task detail state for modal
	let currentTask = $state<Task | null>(null);
	let currentProject = $state<Project | null>(null);
	let loadingTask = $state(false);
	let taskError = $state<string | null>(null);
	let agentContent = $state('');
	let diffContent = $state('');
	let logContent = $state<string[]>([]);
	let eventSource: EventSource | null = null;

	async function loadTaskDetail(taskId: string) {
		if (!taskId) return;

		try {
			loadingTask = true;
			taskError = null;
			agentContent = '';
			diffContent = '';
			logContent = [];

			// Close existing SSE connection
			if (eventSource) {
				eventSource.close();
				eventSource = null;
			}

			const data = await tasksAPI.get(taskId);
			currentTask = data.task;
			currentProject = data.project;
			taskStore.currentTask = data.task;

			// Build initial agent content HTML from messages
			if (data.messages && data.messages.length > 0) {
				agentContent = renderMessagesHTML(data.messages, data.task.status === 'in_progress');
			}

			// Build initial diff content
			if (data.task.status === 'in_progress') {
				diffContent = renderLoadingDiff();
			} else if (data.task.gitDiff) {
				diffContent = renderDiffHTML(data.task.gitDiff);
			} else {
				diffContent = '<div class="text-gray-500 italic">No changes made</div>';
			}

			// Build logs
			if (data.logs && data.logs.length > 0) {
				logContent = data.logs.map((log) => renderLogEntryHTML(log));
			}

			// Set up SSE for real-time updates
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
				onStatus: (html: string) => {
					// Status indicator updated - could update task status
				},
				onComplete: (status: string) => {
					// Task completed - reload task details
					if (currentTask) {
						currentTask.status = status as Task['status'];
					}
				},
				onError: (error) => {
					console.error('Task SSE error:', error);
				}
			});
		} catch (err) {
			taskError = err instanceof Error ? err.message : 'Failed to load task';
			console.error('Task load error:', err);
		} finally {
			loadingTask = false;
		}
	}

	// Helper to render messages as HTML (matches Go backend rendering)
	function renderMessagesHTML(messages: Message[], isInProgress: boolean): string {
		if (messages.length === 0) {
			return '<div class="p-5 text-gray-500 italic text-xs">No agent output</div>';
		}

		let html = '<div class="space-y-0">';
		for (const msg of messages) {
			html += renderMessageBubbleHTML(msg);
		}
		html += '</div>';

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
		const isUser = msg.role === 'user';
		const bgClass = isUser ? 'bg-violet-500/10 border-violet-500/20' : 'bg-gray-800/50 border-gray-700/50';
		const icon = isUser ? 'fa-user' : 'fa-robot';
		const iconColor = isUser ? 'text-violet-400' : 'text-blue-400';

		let contentHtml = '';
		for (const block of msg.content) {
			if (block.type === 'text' && block.text) {
				contentHtml += `<p class="text-sm text-gray-300 leading-normal">${escapeHtml(block.text)}</p>`;
			} else if (block.type === 'tool_use' && block.toolName) {
				contentHtml += `
					<div class="flex items-center gap-2 my-2">
						<span class="px-2 py-0.5 bg-blue-500/10 border border-blue-500/20 rounded text-[10px] font-mono text-blue-300">
							${escapeHtml(block.toolName)}
						</span>
					</div>
				`;
			} else if (block.type === 'tool_result') {
				contentHtml += `<pre class="text-xs text-gray-400 font-mono whitespace-pre-wrap bg-gray-900/50 rounded p-2 my-2">${escapeHtml(block.text || '')}</pre>`;
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

	function renderLoadingDiff(): string {
		return `
			<div class="flex flex-col items-center justify-center h-48 text-gray-500 space-y-4">
				<i class="fas fa-cog fa-spin text-3xl opacity-50"></i>
				<p class="text-xs font-mono">Generating changes...</p>
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

	// Watch for modal open/close
	$effect(() => {
		if (appState.modalOpen && appState.modalTaskId) {
			loadTaskDetail(appState.modalTaskId);
		} else {
			// Close SSE when modal closes
			if (eventSource) {
				eventSource.close();
				eventSource = null;
			}
			currentTask = null;
			currentProject = null;
			agentContent = '';
			diffContent = '';
			logContent = [];
		}
	});

	$effect(() => {
		// Sync task data from store
		if (taskStore.currentTask && appState.modalTaskId === taskStore.currentTask.id) {
			currentTask = { ...currentTask, ...taskStore.currentTask } as Task;
		}
	});

	onDestroy(() => {
		if (eventSource) {
			eventSource.close();
		}
	});
</script>

<div class="h-screen flex flex-col overflow-hidden bg-background">
	<Toast />
	<Header />

	<main class="flex-1 overflow-y-auto bg-background relative pt-14" id="feed-container">
		<div class="px-3 pt-6 pb-40">{@render children()}</div>
	</main>

	<!-- Input Bar -->
	<div class="fixed bottom-6 left-4 right-4 z-20 mx-auto max-w-3xl">
		<ChatInput mode="create" placeholder="What do you want to build?" />
	</div>

	<!-- Task Detail Modal -->
	{#if appState.modalOpen && appState.modalTaskId}
		<div
			class="fixed inset-0 z-50 bg-popover flex flex-col overflow-hidden transition-all duration-300"
			class:translate-x-0={appState.modalOpen}
			class:translate-x-full={!appState.modalOpen}
		>
			{#if loadingTask}
				<div class="flex items-center justify-center h-full">
					<div class="flex flex-col items-center gap-3">
						<div class="w-8 h-8 rounded-lg bg-violet-500/10 border border-violet-500/20 flex items-center justify-center">
							<i class="fas fa-spinner fa-spin text-sm text-violet-400"></i>
						</div>
						<p class="text-xs text-gray-500">Loading task...</p>
					</div>
				</div>
			{:else if taskError}
				<div class="flex items-center justify-center h-full">
					<div class="text-center">
						<p class="text-sm text-red-400 mb-2">{taskError}</p>
						<button
							onclick={() => loadTaskDetail(appState.modalTaskId!)}
							class="px-4 py-2 bg-violet-500/20 border border-violet-500/30 rounded-lg text-xs text-violet-300 hover:bg-violet-500/30 transition-colors"
						>
							Retry
						</button>
					</div>
				</div>
			{:else if currentTask && currentProject}
				<TaskDetail
					task={currentTask}
					project={currentProject}
					{agentContent}
					{diffContent}
					{logContent}
				/>
			{/if}
		</div>
	{/if}
</div>
