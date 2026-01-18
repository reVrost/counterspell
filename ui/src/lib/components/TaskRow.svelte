<script lang="ts">
	import { appState } from '$lib/stores/app.svelte';
	import { taskTimers } from '$lib/stores/tasks.svelte';
	import { cn } from '$lib/utils';
	import type { Task, Project } from '$lib/types';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import CheckIcon from '@lucide/svelte/icons/check';
	import { onMount } from 'svelte';

	interface Props {
		task: Task;
		project: Project;
		variant: 'planning' | 'active' | 'agent_review' | 'human_review' | 'completed';
	}

	let { task, project, variant }: Props = $props();

	let elapsed = $state(0);

	onMount(() => {
		if (variant === 'active') {
			if (!taskTimers[task.id]) {
				taskTimers[task.id] = Date.now();
			}
			elapsed = Math.floor((Date.now() - taskTimers[task.id]) / 1000);

			const interval = setInterval(() => {
				elapsed = Math.floor((Date.now() - taskTimers[task.id]) / 1000);
			}, 1000);

			return () => clearInterval(interval);
		}
	});

	function handleClick() {
		appState.openModal(task.id);
	}

	const baseClasses =
		'w-full text-left bg-card border rounded-2xl p-4 transition-all duration-150 shadow-sm focus:outline-none focus:ring-2';

	const variantClasses = {
		planning:
			'border-purple-800/50 active:scale-[0.98] active:bg-purple-800/50 focus:ring-purple-500/50',
		active:
			'border-gray-800/50 active:scale-[0.98] active:bg-gray-800/50 focus:ring-orange-500/50',
		agent_review:
			'border-yellow-800/50 relative active:scale-[0.98] active:bg-yellow-800/30 focus:ring-yellow-500/50',
		human_review:
			'border-blue-800 relative active:scale-[0.98] active:bg-blue-800/80 focus:ring-blue-500/50',
		completed:
			'bg-card/60 border-gray-800/20 flex-row active:scale-[0.98] active:bg-gray-800/50 focus:ring-green-500/50'
	};
</script>

<button type="button" class="{baseClasses} {variantClasses[variant]}" onclick={handleClick}>
	{#if variant === 'completed'}
		<!-- Completed Layout -->
		<div class="flex justify-between items-center w-full">
			<div class="flex items-center gap-3">
				<div
					class="w-6 h-6 rounded-full bg-green-900/40 text-green-500 flex items-center justify-center text-xs shrink-0"
				>
					<CheckIcon class="w-3 h-3" />
				</div>
				<div class="min-w-0">
					<div class="text-base text-gray-400 leading-snug line-clamp-2">{task.description}</div>
					<div class="text-xs text-gray-600 mt-0.5">{project.name}</div>
				</div>
			</div>
			<ChevronRightIcon class="w-4 h-4 text-gray-700 ml-3 shrink-0" />
		</div>
	{:else}
		<!-- Active/Review Layout -->
		<div class="flex justify-between items-start mb-2">
			<div class="flex items-center gap-2">
				<span
					class={cn(
						project.color,
						'opacity-80 w-6 h-6 rounded-lg flex items-center justify-center text-xs',
						(variant === 'agent_review' || variant === 'human_review') && 'bg-gray-800/50 border border-gray-700/50'
					)}
				>
					<i class="fas {project.icon}"></i>
				</span>
				<span class="text-sm font-medium text-gray-400">{project.name}</span>
			</div>

			{#if variant === 'planning'}
				<div class="flex items-center gap-2">
					<span
						class="text-xs text-purple-400 bg-purple-500/10 px-2.5 py-1 rounded-lg font-medium border border-purple-500/20"
					>
						Planning
					</span>
					<span class="text-xs text-purple-400/80 font-mono tabular-nums">{elapsed}s</span>
				</div>
			{:else if variant === 'active'}
				<div class="flex items-center gap-2">
					<span
						class="text-xs text-orange-400 bg-orange-500/10 px-2.5 py-1 rounded-lg font-medium border border-orange-500/20"
					>
						In Progress
					</span>
					<span class="text-xs text-orange-400/80 font-mono tabular-nums">{elapsed}s</span>
				</div>
			{:else if variant === 'agent_review'}
				<span
					class="text-xs text-yellow-400 bg-yellow-500/10 px-2.5 py-1 rounded-lg font-medium border border-yellow-500/20"
				>
					Agent Review
				</span>
			{:else if variant === 'human_review'}
				<span
					class="text-xs text-blue-400 bg-blue-500/10 px-2.5 py-1 rounded-lg font-medium border border-blue-500/20"
				>
					Human Review
				</span>
			{/if}
		</div>

		<p class="text-base text-gray-200 font-medium leading-snug line-clamp-2" class:pr-6={variant === 'agent_review' || variant === 'human_review'}>
			{task.description}
		</p>

		{#if variant === 'agent_review' || variant === 'human_review'}
			<div class="absolute right-4 top-1/2 -translate-y-1/2 text-gray-600">
				<ChevronRightIcon class="w-4 h-4" />
			</div>
		{/if}
	{/if}
</button>
