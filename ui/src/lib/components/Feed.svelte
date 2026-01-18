<script lang="ts">
	import type { FeedData } from '$lib/types';
	import TaskRow from './TaskRow.svelte';
	import { slide, DURATIONS } from '$lib/utils/transitions';

	interface Props {
		feedData: FeedData;
	}

	let { feedData }: Props = $props();

	// Split active into pending and in_progress
	const pendingTasks = $derived(feedData.active.filter((t) => t.status === 'pending'));
	const inProgressTasks = $derived(feedData.active.filter((t) => t.status === 'in_progress'));
</script>

<div id="feed-content">
	<!-- Review (needs user action) -->
	{#if feedData.reviews.length > 0}
		<div class="mb-6">
			<h3 class="px-2 text-xs font-bold text-blue-500 uppercase tracking-wider mb-3">
				Review
			</h3>
			<div class="space-y-3">
				{#each feedData.reviews as task, index (task.id)}
					<div transition:slide|local={{ direction: 'up', duration: DURATIONS.normal, delay: index * 50 }}>
						<TaskRow {task} project={feedData.projects[task.project_id]} variant="review" />
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<!-- In Progress -->
	{#if inProgressTasks.length > 0}
		<div class="mb-6">
			<h3 class="px-2 text-xs font-bold text-orange-500 uppercase tracking-wider mb-3">In Progress</h3>
			<div class="space-y-3">
				{#each inProgressTasks as task, index (task.id)}
					<div transition:slide|local={{ direction: 'up', duration: DURATIONS.normal, delay: index * 50 }}>
						<TaskRow {task} project={feedData.projects[task.project_id]} variant="active" />
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Pending -->
	{#if pendingTasks.length > 0}
		<div class="mb-6">
			<h3 class="px-2 text-xs font-bold text-gray-500 uppercase tracking-wider mb-3">Pending</h3>
			<div class="space-y-3">
				{#each pendingTasks as task, index (task.id)}
					<div transition:slide|local={{ direction: 'up', duration: DURATIONS.normal, delay: index * 50 }}>
						<TaskRow {task} project={feedData.projects[task.project_id]} variant="pending" />
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<!-- No active tasks message -->
	{#if feedData.active.length === 0 && feedData.reviews.length === 0}
		<div class="mb-6">
			<div class="px-2 py-2 text-sm text-gray-600 text-center">No active agents running</div>
		</div>
	{/if}

	<!-- Completed -->
	{#if feedData.done.length > 0}
		<div class="pt-4 border-t border-gray-800/50">
			<h3 class="px-2 text-xs font-bold text-gray-600 uppercase tracking-wider mb-3">Completed</h3>
			<div class="space-y-3">
				{#each feedData.done as task, index (task.id)}
					<div transition:slide|local={{ direction: 'up', duration: DURATIONS.normal, delay: index * 50 }}>
						<TaskRow {task} project={feedData.projects[task.project_id]} variant="completed" />
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>
