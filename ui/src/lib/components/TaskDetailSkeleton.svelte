<script lang="ts">
  import Skeleton from './Skeleton.svelte';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
</script>

<div class="flex flex-col h-full bg-popover">
  <!-- Header Skeleton -->
  <div
    class="px-4 py-2 border-b border-white/5 flex items-center justify-between shrink-0 bg-popover"
  >
    <div class="flex items-center gap-3">
      <div class="w-11 h-11 rounded-full flex items-center justify-center text-gray-700">
        <ArrowLeftIcon class="w-5 h-5" />
      </div>
      <div class="space-y-1.5">
        <div class="flex items-center gap-2">
          <Skeleton variant="text" class="w-16 h-3" />
          <Skeleton variant="text" class="w-20 h-3" />
        </div>
        <Skeleton variant="text" class="w-48 h-5" />
      </div>
    </div>
    <Skeleton variant="rounded" class="w-24 h-7 rounded-full" />
  </div>

  <!-- Tabs Skeleton -->
  <div
    class="flex items-center justify-between p-2 px-7 bg-popover shrink-0 border-b border-white/5"
  >
    <div class="flex bg-gray-900 rounded-lg p-0.5 border border-gray-700/50 gap-1">
      {#each [1, 2, 3] as _}
        <Skeleton variant="rounded" class="h-8 w-16" />
      {/each}
    </div>
    <div class="flex items-center gap-2">
      <Skeleton variant="circular" class="w-8 h-8" />
      <Skeleton variant="rounded" class="w-16 h-8" />
      <Skeleton variant="rounded" class="w-14 h-8" />
    </div>
  </div>

  <!-- Main Content Skeleton -->
  <div class="flex-1 overflow-y-auto p-6 space-y-6">
    <div class="space-y-3">
      <Skeleton variant="text" class="w-full h-4" />
      <Skeleton variant="text" class="w-5/6 h-4" />
      <Skeleton variant="text" class="w-2/3 h-4" />
    </div>

    <div class="pt-8 space-y-4">
      {#each [1, 2] as _}
        <div class="flex gap-4">
          <Skeleton variant="circular" class="w-8 h-8 shrink-0" />
          <div class="space-y-2 flex-1 pt-1">
            <Skeleton variant="text" class="w-1/4 h-3" />
            <Skeleton variant="rounded" class="h-24 w-full" />
          </div>
        </div>
      {/each}
    </div>
  </div>

  <!-- Chat Input Skeleton -->
  <div class="p-6 border-t border-white/5 bg-gradient-to-t from-[#0D1117]">
    <Skeleton variant="rounded" class="w-full h-12 rounded-2xl" />
  </div>
</div>
