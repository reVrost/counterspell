<script lang="ts">
  import Skeleton from './Skeleton.svelte';
</script>

<div class="space-y-4 p-2">
  <div class="flex items-center justify-between mb-4">
    <Skeleton variant="text" class="w-32 h-4" />
    <Skeleton variant="text" class="w-16 h-3" />
  </div>
  {#each [1, 2, 3] as _}
    <div class="space-y-2">
      <Skeleton variant="text" class="w-2/3 h-5 opacity-40" />
      <div class="pl-4 space-y-1">
        <Skeleton variant="rounded" class="w-full h-4 bg-red-500/5" />
        <Skeleton variant="rounded" class="w-full h-4 bg-green-500/5" />
        <Skeleton variant="rounded" class="w-full h-4 bg-green-500/5" />
        <Skeleton variant="text" class="w-5/6 h-4 opacity-20" />
      </div>
    </div>
  {/each}
</div>
