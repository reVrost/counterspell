<script lang="ts">
  import Skeleton from './Skeleton.svelte';
</script>

<div class="space-y-3">
  {#each Array(3) as _}
    <div
      class="w-full bg-card border border-white/5 rounded-sm p-4 shadow-sm flex justify-between items-start gap-4"
    >
      <div class="flex-1 min-w-0 space-y-2">
        <Skeleton variant="text" class="w-3/4 h-5" />
        <div class="flex items-center gap-1.5">
          <Skeleton variant="circular" class="w-3 h-3 opacity-50" />
          <Skeleton variant="text" class="w-24 h-3" />
        </div>
        <div class="pt-1 space-y-1.5">
          <Skeleton variant="text" class="w-full h-3 opacity-70" />
          <Skeleton variant="text" class="w-5/6 h-3 opacity-70" />
        </div>
      </div>
      <div class="shrink-0 flex flex-col items-end gap-1.5">
        <Skeleton variant="rounded" class="w-20 h-5 rounded-full" />
        <Skeleton variant="text" class="w-12 h-3" />
      </div>
    </div>
  {/each}
</div>
