<script lang="ts">
	import type { FeedData } from '$lib/types';
	import TaskRow from './TaskRow.svelte';

	interface Props {
		feedData: FeedData;
	}

	let { feedData }: Props = $props();
</script>

<div id="feed-content">
	<!-- Needs Review -->
	{#if feedData.reviews.length > 0}
		<div class="mb-6">
			<h3 class="px-2 text-xs font-bold text-gray-500 uppercase tracking-wider mb-3">
				Needs Review
			</h3>
			<div class="space-y-3">
				{#each feedData.reviews as task (task.id)}
					<TaskRow {task} project={feedData.projects[task.projectId]} variant="review" />
				{/each}
			</div>
		</div>
	{/if}

	<!-- In Progress -->
	<div class="mb-6">
		<h3 class="px-2 text-xs font-bold text-gray-500 uppercase tracking-wider mb-3">In Progress</h3>
		<div class="space-y-3">
			{#each feedData.active as task (task.id)}
				<TaskRow {task} project={feedData.projects[task.projectId]} variant="active" />
			{/each}
			{#if feedData.active.length === 0}
				<div class="px-2 py-2 text-sm text-gray-600 text-center">No active agents running</div>
			{/if}
		</div>
	</div>

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
