<script lang="ts">
  import { fly, scale } from 'svelte/transition';

  let {
    title = '500',
    message = 'Internal Server Error',
    description = "Something went wrong on our end. We're looking into it.",
    onRetry = null as (() => void) | null,
    homeLink = '/',
  } = $props();

  function goHome() {
    window.location.href = homeLink;
  }

  let mounted = $state(false);
  $effect(() => {
    mounted = true;
  });
</script>

<div
  class="fixed inset-0 flex flex-col items-center overflow-hidden bg-background selection:bg-primary/20"
>
  <!-- Sophisticated Background -->
  <div class="absolute inset-0 pointer-events-none overflow-hidden">
    <div
      class="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-primary/5 blur-[120px] rounded-full"
    ></div>
    <div
      class="absolute bottom-[10%] right-[-5%] w-[35%] h-[35%] bg-violet-500/5 blur-[100px] rounded-full"
    ></div>
  </div>

  <!-- Main Content Area -->
  <div
    class="relative flex-1 flex flex-col items-center justify-center w-full max-w-sm px-8 pt-12 text-center"
  >
    {#if mounted}
      <div in:scale={{ duration: 800, start: 0.95, delay: 100 }} class="relative mb-12">
        <!-- Status Code Display -->
        <div class="relative z-10">
          <span
            class="text-[120px] font-black tracking-tighter leading-none opacity-[0.03] select-none block"
          >
            {title}
          </span>
          <div class="absolute inset-0 flex items-center justify-center">
            <div
              class="w-16 h-16 rounded-3xl bg-destructive/10 border border-destructive/20 flex items-center justify-center rotate-12"
            >
              <i class="fas fa-exclamation-triangle text-destructive text-2xl -rotate-12"></i>
            </div>
          </div>
        </div>
      </div>

      <div class="space-y-4" in:fly={{ y: 20, duration: 600, delay: 300 }}>
        <h1 class="text-3xl font-bold tracking-tight text-foreground">
          {message}
        </h1>
        <p class="text-base text-muted-foreground/80 leading-relaxed font-medium">
          {description}
        </p>
      </div>
    {/if}
  </div>

  <!-- Mobile-Native Action Bottom Bar -->
  <div
    class="relative w-full max-w-sm px-6 pb-12 space-y-3"
    in:fly={{ y: 40, duration: 700, delay: 500 }}
  >
    {#if onRetry}
      <button
        onclick={onRetry}
        class="w-full flex items-center justify-center gap-3 py-5 bg-foreground text-background font-bold text-sm rounded-2xl active:scale-[0.98] transition-all shadow-xl shadow-foreground/5 overflow-hidden group"
      >
        <i
          class="fas fa-redo-alt text-[12px] group-hover:rotate-180 transition-transform duration-500"
        ></i>
        Try again
      </button>
    {/if}

    <button
      onclick={goHome}
      class="w-full flex items-center justify-center gap-3 py-5 bg-secondary border border-border/50 text-foreground font-bold text-sm rounded-2xl active:scale-[0.98] transition-all"
    >
      <i class="fas fa-home text-[12px]"></i>
      Return to Dashboard
    </button>

    <div class="pt-8 flex justify-center opacity-30 select-none">
      <span class="text-[10px] uppercase font-black tracking-[0.3em]">Counterspell</span>
    </div>
  </div>
</div>

<style>
  :global(body) {
    overflow: hidden;
  }
</style>
