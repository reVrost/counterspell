<script lang="ts">
  import Feed from '$lib/components/Feed.svelte';
  import FeedSkeleton from '$lib/components/FeedSkeleton.svelte';
  import SessionsView from '$lib/components/SessionsView.svelte';
  import type { FeedData } from '$lib/types';
  import { tasksAPI } from '$lib/api';
  import { createFeedSSE } from '$lib/utils/sse';
  import { appState } from '$lib/stores/app.svelte';
  import { taskStore } from '$lib/stores/tasks.svelte';
  import ErrorView from '$lib/components/ErrorView.svelte';

  let feedData = $state<FeedData>({
    active: [],
    reviews: [],
    done: [],
    todo: [],
    planning: [],
    projects: {},
  });

  let loading = $state(true);
  let error = $state<string | null>(null);
  let eventSource: EventSource | null = null;

  async function loadFeedData() {
    try {
      loading = true;
      error = null;
      const data = await tasksAPI.getFeed();

      // Defensive check: ensure data is an object and has required properties
      if (!data || typeof data !== 'object') {
        throw new Error('Invalid feed data received');
      }

      // Ensure arrays exist
      data.active = data.active || [];
      data.reviews = data.reviews || [];
      data.done = data.done || [];
      data.todo = data.todo || [];

      feedData = data;

      // Update review count
      taskStore.reviewCount = data.reviews.length;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load feed';
      console.error('Feed load error:', err, 'stack:', err instanceof Error ? err.stack : '');
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    // Load feed data immediately
    loadFeedData();

    // Set up SSE for real-time updates
    eventSource = createFeedSSE(() => {
      loadFeedData();
    });

    return () => {
      if (eventSource) {
        eventSource.close();
      }
    };
  });
</script>

<svelte:head>
  <title>Dashboard | Counterspell</title>
</svelte:head>

{#if appState.activeTab === 'sessions'}
  <SessionsView />
{:else if loading}
  <div data-testid="loading-state" class="flex flex-col gap-6">
    <div class="flex flex-col items-center gap-4 py-4">
      <div
        class="w-12 h-12 rounded-2xl bg-violet-500/10 border border-violet-500/20 flex items-center justify-center animate-pulse-glow"
      >
        <i class="fas fa-bolt text-xl text-violet-400"></i>
      </div>
      <div class="flex flex-col items-center text-center">
        <h3 class="text-sm font-semibold text-gray-200">Updating Feed</h3>
        <p class="text-[11px] text-gray-500 font-medium">Preparing your workspace...</p>
      </div>
    </div>
    <FeedSkeleton />
  </div>
{:else if error}
  <div data-testid="error-state">
    <ErrorView
      title="Error"
      message="Failed to load feed"
      description={error}
      onRetry={loadFeedData}
      homeLink="/app"
    />
  </div>
{:else}
  <div data-testid="feed-loaded">
    <Feed {feedData} />
  </div>
{/if}
