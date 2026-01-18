<script lang="ts">
	import type { FeedData } from '$lib/types';
	import TaskRow from './TaskRow.svelte';

	interface Props {
		feedData: FeedData;
	}

	let { feedData }: Props = $props();

	// Split reviews into agent_review and human_review
	const agentReviews = $derived(feedData.reviews.filter((t) => t.status === 'agent_review'));
	const humanReviews = $derived(feedData.reviews.filter((t) => t.status === 'human_review'));
	// Split active into planning and in_progress
	const planningTasks = $derived(feedData.active.filter((t) => t.status === 'planning'));
	const inProgressTasks = $derived(feedData.active.filter((t) => t.status === 'in_progress'));
</script>

<div id="feed-content">
	<!-- Human Review (highest priority - needs user action) -->
	{#if humanReviews.length > 0}
		<div class="mb-6">
			<h3 class="px-2 text-xs font-bold text-blue-500 uppercase tracking-wider mb-3">
				Human Review
			</h3>
			<div class="space-y-3">
				{#each humanReviews as task (task.id)}
					<TaskRow {task} project={feedData.projects[task.projectId]} variant="human_review" />
				{/each}
			</div>
		</div>
	{/if}

	<!-- Agent Review -->
	{#if agentReviews.length > 0}
		<div class="mb-6">
			<h3 class="px-2 text-xs font-bold text-yellow-500 uppercase tracking-wider mb-3">
				Agent Review
			</h3>
			<div class="space-y-3">
				{#each agentReviews as task (task.id)}
					<TaskRow {task} project={feedData.projects[task.projectId]} variant="agent_review" />
				{/each}
			</div>
		</div>
	{/if}

	<!-- In Progress -->
	{#if inProgressTasks.length > 0}
		<div class="mb-6">
			<h3 class="px-2 text-xs font-bold text-orange-500 uppercase tracking-wider mb-3">In Progress</h3>
			<div class="space-y-3">
				{#each inProgressTasks as task (task.id)}
					<TaskRow {task} project={feedData.projects[task.projectId]} variant="active" />
				{/each}
			</div>
		</div>
	{/if}

	<!-- Planning -->
	{#if planningTasks.length > 0}
		<div class="mb-6">
			<h3 class="px-2 text-xs font-bold text-purple-500 uppercase tracking-wider mb-3">Planning</h3>
			<div class="space-y-3">
				{#each planningTasks as task (task.id)}
					<TaskRow {task} project={feedData.projects[task.projectId]} variant="planning" />
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
				{#each feedData.done as task (task.id)}
					<TaskRow {task} project={feedData.projects[task.projectId]} variant="completed" />
				{/each}
			</div>
		</div>
	{/if}
</div>
