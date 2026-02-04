<script lang="ts">
  import { cn } from '$lib/utils';
  interface Props {
    variant?: 'text' | 'rounded' | 'circular';
    class?: string;
    shimmer?: boolean;
  }

  let { variant = 'text', class: className = '', shimmer = true }: Props = $props();

  const variants = {
    text: 'h-4 bg-white/5 rounded',
    rounded: 'rounded-xl bg-white/5',
    circular: 'rounded-full bg-white/5',
  };
</script>

<div
  class={cn(
    variants[variant],
    shimmer &&
      'relative overflow-hidden before:absolute before:inset-0 before:-translate-x-full before:animate-[shimmer_2s_infinite] before:bg-gradient-to-r before:from-transparent before:via-white/[0.03] before:to-transparent',
    !shimmer && 'animate-pulse',
    className
  )}
></div>
