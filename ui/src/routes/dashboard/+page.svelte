<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import Feed from '$lib/components/Feed.svelte';
	import type { FeedData } from '$lib/types';
	import { tasksAPI } from '$lib/api';
	import { createFeedSSE } from '$lib/utils/sse';
	import { appState } from '$lib/stores/app.svelte';

	let feedData = $state<FeedData>({
		active: [],
		reviews: [],
		done: [],
		todo: [],
		projects: {}
	});

	let loading = $state(true);
	let error = $state<string | null>(null);
	let eventSource: EventSource | null = null;

	async function loadFeedData() {
		try {
			loading = true;
			error = null;
			const data = await tasksAPI.getFeed();
			feedData = data;

			// Update projects in app state
			appState.projects = Object.values(data.projects || {});
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load feed';
			console.error('Feed load error:', err);
		} finally {
			loading = false;
		}
	}

	onMount(async () => {
		await loadFeedData();

		// Set up SSE for real-time updates
		eventSource = createFeedSSE(() => {
			loadFeedData();
		});
	});

	onDestroy(() => {
		if (eventSource) {
			eventSource.close();
		}
	});
</script>

<svelte:head>
	<title>Dashboard | Counterspell</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center h-64">
		<div class="flex flex-col items-center gap-3">
			<div class="w-8 h-8 rounded-lg bg-violet-500/10 border border-violet-500/20 flex items-center justify-center">
				<i class="fas fa-spinner fa-spin text-sm text-violet-400"></i>
			</div>
			<p class="text-xs text-gray-500">Loading feed...</p>
		</div>
	</div>
{:else if error}
	<div class="flex items-center justify-center h-64">
		<div class="text-center">
			<p class="text-sm text-red-400 mb-2">{error}</p>
			<button
				onclick={loadFeedData}
				class="px-4 py-2 bg-violet-500/20 border border-violet-500/30 rounded-lg text-xs text-violet-300 hover:bg-violet-500/30 transition-colors"
			>
				Retry
			</button>
		</div>
	</div>
{:else}
	<Feed {feedData} />
{/if}
