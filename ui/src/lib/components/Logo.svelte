<script lang="ts">
  import { cn } from '$lib/utils';
  let { class: className = '' } = $props();
</script>

<svg
  viewBox="0 0 24 24"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
  class={cn('w-6 h-6', className)}
>
  <defs>
    <linearGradient id="logo-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#a855f7" />
      <stop offset="100%" style="stop-color:#ec4899" />
    </linearGradient>
    <filter id="logo-glow" x="-40%" y="-40%" width="180%" height="180%">
      <feGaussianBlur stdDeviation="1.5" result="blur" />
      <feComposite in="SourceGraphic" in2="blur" operator="over" />
    </filter>
  </defs>
  <path
    d="M5 7H17V19"
    stroke="url(#logo-gradient)"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
    filter="url(#logo-glow)"
    class="drop-shadow-[0_0_8px_rgba(168,85,247,0.5)]"
  />
</svg>
