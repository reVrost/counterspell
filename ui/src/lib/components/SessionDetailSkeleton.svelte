<script lang="ts">
  import Skeleton from './Skeleton.svelte';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
</script>

<div class="flex flex-col h-[100dvh]">
  <!-- Header Skeleton -->
  <div
    class="px-4 py-2 border-b border-white/5 flex items-center justify-between shrink-0 bg-popover"
  >
    <div class="flex items-center gap-3 min-w-0">
      <div class="w-11 h-11 rounded-full flex items-center justify-center text-gray-700">
        <ArrowLeftIcon class="w-5 h-5" />
      </div>
      <div class="min-w-0 space-y-1.5">
        <div class="flex items-center gap-2">
          <Skeleton variant="text" class="w-12 h-3" />
          <Skeleton variant="text" class="w-20 h-3" />
        </div>
        <Skeleton variant="text" class="w-48 h-5" />
      </div>
    </div>
    <Skeleton variant="rounded" class="w-20 h-8 rounded-full" />
  </div>

  <!-- Content Skeleton -->
  <div class="flex-1 overflow-hidden flex flex-col relative">
    <div class="flex-1 overflow-y-auto relative px-4 pb-28 pt-3 space-y-6">
      {#each [1, 2, 3] as i}
        <div class="flex gap-4 {i % 2 === 0 ? 'flex-row-reverse' : ''}">
          <Skeleton variant="circular" class="w-8 h-8 shrink-0" />
          <div class="space-y-2 max-w-[80%] {i % 2 === 0 ? 'flex flex-col items-end' : ''}">
            <Skeleton variant="rounded" class="h-16 w-64 sm:w-80" />
            <Skeleton variant="text" class="w-20 h-2" />
          </div>
        </div>
      {/each}
    </div>

    <!-- Input Area Skeleton -->
    <div class="absolute bottom-0 inset-x-0 z-10 pb-4 px-4 bg-gradient-to-t from-[#0D1117] pt-10">
      <div class="relative mx-auto max-w-3xl">
        <Skeleton variant="rounded" class="w-full h-14 rounded-2xl" />
      </div>
    </div>
  </div>
</div>
