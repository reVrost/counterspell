<script lang="ts">
  import { cn } from "$lib/utils";
  import type { Snippet } from "svelte";

  type Props = {
    class?: string;
    children: Snippet;
  };

  let { class: className = "", children }: Props = $props();
</script>

<div class={cn("text-lg font-semibold leading-none tracking-tight", className)}>
  {@render children()}
</div>
