<script lang="ts">
  import Skeleton from './Skeleton.svelte';
</script>

<div class="divide-y divide-white/[0.04] border-t border-white/[0.04]">
  {#each Array(5) as _}
    <div class="w-full py-5 px-4 sm:px-6 flex items-start justify-between gap-4">
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2.5 mb-2">
          <Skeleton variant="text" class="w-48 h-5" />
          <Skeleton variant="rounded" class="w-16 h-4 rounded-full" />
        </div>
        <div class="flex items-center gap-2">
          <Skeleton variant="circular" class="w-3 h-3 opacity-50" />
          <Skeleton variant="text" class="w-20 h-3" />
        </div>
      </div>
      <div class="flex flex-col items-end pt-1">
        <Skeleton variant="text" class="w-10 h-3" />
      </div>
    </div>
  {/each}
</div>
