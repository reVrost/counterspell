<script lang="ts">
	import { taskStore } from '$lib/stores/tasks.svelte';
	import { cn } from '$lib/utils';
	import ListChecksIcon from '@lucide/svelte/icons/list-checks';
	import CheckIcon from '@lucide/svelte/icons/check';
	import XIcon from '@lucide/svelte/icons/x';
	import ClipboardListIcon from '@lucide/svelte/icons/clipboard-list';

	let showTodos = $state(false);
</script>

{#if taskStore.todos.length > 0}
	<!-- Floating Button -->
	<div class="fixed bottom-24 right-4 z-30">
		<button
			onclick={() => (showTodos = true)}
			class="group flex items-center gap-2 h-9 pl-3 pr-4 bg-popover hover:bg-card border border-white/[0.08] hover:border-purple-500/30 rounded-full shadow-xl shadow-black/40 transition-all"
		>
			<!-- Progress circle -->
			<div class="relative w-5 h-5">
				<svg class="w-5 h-5 -rotate-90" viewBox="0 0 20 20">
					<circle
						cx="10"
						cy="10"
						r="8"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						class="text-gray-700"
					/>
					<circle
						cx="10"
						cy="10"
						r="8"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						class="text-green-500 transition-all duration-300"
						stroke-dasharray="{(taskStore.completedCount / taskStore.todos.length) * 50.26} 50.26"
					/>
				</svg>
				<span
					class="absolute inset-0 flex items-center justify-center text-[8px] font-bold text-white"
				>
					{taskStore.completedCount}/{taskStore.todos.length}
				</span>
			</div>
			<!-- Current task -->
			{#if taskStore.inProgressTask}
				<span
					class="text-[11px] text-gray-400 group-hover:text-gray-300 max-w-[120px] truncate transition-colors"
				>
					{taskStore.inProgressTask}
				</span>
			{:else if taskStore.completedCount < taskStore.todos.length}
				<span class="text-[11px] text-gray-500">Tasks</span>
			{:else}
				<span class="text-[11px] text-green-400">Done!</span>
			{/if}
		</button>
	</div>
{/if}

<!-- Todo Modal -->
{#if showTodos}
	<div
		class="fixed inset-0 z-[200] flex items-start justify-center pt-[15vh] bg-black/60 backdrop-blur-sm"
		onclick={(e) => e.target === e.currentTarget && (showTodos = false)}
	>
		<div
			class="bg-popover border border-white/[0.08] rounded-2xl w-[380px] max-h-[60vh] shadow-2xl shadow-black/50 overflow-hidden flex flex-col"
		>
			<!-- Modal Header -->
			<div
				class="px-4 py-3 border-b border-white/[0.06] flex items-center justify-between shrink-0"
			>
				<div class="flex items-center gap-2">
					<div class="w-6 h-6 rounded-lg bg-purple-500/10 flex items-center justify-center">
						<ListChecksIcon class="w-3 h-3 text-purple-400" />
					</div>
					<h3 class="text-sm font-semibold text-white">Agent Tasks</h3>
				</div>
				<div class="flex items-center gap-2">
					<span class="text-[10px] text-gray-500 font-mono">
						{taskStore.completedCount} / {taskStore.todos.length} done
					</span>
					<button
						onclick={() => (showTodos = false)}
						class="w-6 h-6 rounded-lg hover:bg-white/[0.06] flex items-center justify-center text-gray-500 hover:text-gray-300 transition-colors"
					>
						<XIcon class="w-3 h-3" />
					</button>
				</div>
			</div>

			<!-- Progress Bar -->
			<div class="h-0.5 bg-gray-800">
				<div
					class="h-full bg-gradient-to-r from-purple-500 to-green-500 transition-all duration-300"
					style="width: {taskStore.todos.length
						? (taskStore.completedCount / taskStore.todos.length) * 100
						: 0}%"
				></div>
			</div>

			<!-- Task List -->
			<div class="flex-1 overflow-y-auto p-2 space-y-1">
				{#each taskStore.todos as todo, index}
					<div
						class={cn(
							'flex items-start gap-3 px-3 py-2 rounded-xl transition-colors',
							todo.status === 'completed' && 'bg-green-500/5',
							todo.status === 'in_progress' && 'bg-purple-500/10 border border-purple-500/20',
							todo.status === 'pending' && 'hover:bg-white/[0.02]'
						)}
					>
						<!-- Status Icon -->
						<div class="w-5 h-5 shrink-0 flex items-center justify-center mt-0.5">
							{#if todo.status === 'completed'}
								<div
									class="w-4 h-4 rounded-full bg-green-500/20 flex items-center justify-center"
								>
									<CheckIcon class="w-2 h-2 text-green-400" />
								</div>
							{:else if todo.status === 'in_progress'}
								<div
									class="w-4 h-4 rounded-full bg-purple-500/20 flex items-center justify-center relative"
								>
									<div class="w-1.5 h-1.5 rounded-full bg-purple-400 animate-pulse"></div>
									<div
										class="absolute inset-0 rounded-full border border-purple-400/50 animate-ping"
										style="animation-duration: 2s;"
									></div>
								</div>
							{:else}
								<div class="w-4 h-4 rounded-full border border-gray-600"></div>
							{/if}
						</div>

						<!-- Content -->
						<div class="flex-1 min-w-0">
							<p
								class="text-sm leading-tight"
								class:text-gray-400={todo.status === 'completed'}
								class:line-through={todo.status === 'completed'}
								class:text-white={todo.status === 'in_progress'}
								class:text-gray-300={todo.status === 'pending'}
							>
								{todo.status === 'in_progress' ? todo.activeForm || todo.content : todo.content}
							</p>
						</div>
					</div>
				{/each}

				<!-- Empty State -->
				{#if taskStore.todos.length === 0}
					<div class="py-8 text-center">
						<div
							class="w-12 h-12 rounded-full bg-gray-800 flex items-center justify-center mx-auto mb-3"
						>
							<ClipboardListIcon class="w-5 h-5 text-gray-600" />
						</div>
						<p class="text-sm text-gray-500">No tasks yet</p>
						<p class="text-xs text-gray-600 mt-1">The agent will create tasks as needed</p>
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}
