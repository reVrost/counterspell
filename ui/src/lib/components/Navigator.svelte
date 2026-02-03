<script lang="ts">
  import { cn } from '$lib/utils';
  import InboxIcon from '@lucide/svelte/icons/inbox';
  import { SquarePen } from '@lucide/svelte';
  import FolderIcon from '@lucide/svelte/icons/folder';
  import MessageSquareIcon from '@lucide/svelte/icons/message-square';
  import LayersIcon from '@lucide/svelte/icons/layers';
  import SearchIcon from '@lucide/svelte/icons/search';
  import { appState } from '$lib/stores/app.svelte';
  import { taskStore } from '$lib/stores/tasks.svelte';

  interface Props {
    activeTab?: 'inbox' | 'sessions' | 'projects' | 'focus' | 'layers';
    onNavigate?: (tab: string) => void;
    onSearch?: () => void;
  }

  let { activeTab = 'inbox', onNavigate, onSearch }: Props = $props();

  const tabs = ['inbox', 'sessions', 'focus', 'layers'];
  const activeIndex = $derived(tabs.indexOf(activeTab || 'inbox'));
  const navIndex = $derived(activeIndex === -1 ? 0 : activeIndex);
  const navButtonSize = 56;
  const navBaseSize = 48;
  const navGap = 4;
  const navStep = navButtonSize + navGap - 11;
  const navTop = (navButtonSize - navBaseSize) / 2 - 1;

  function handleTabClick(tab: string) {
    if (onNavigate) {
      onNavigate(tab);
    }
  }

  const navBase =
    'absolute left-1.2 h-12 w-14 bg-[#2a2a2a] rounded-full transition-all gap-1 border border-white/[0.01]';
</script>

<div class="flex items-center justify-center w-full">
  <!-- Main Navigation Pill -->
  <div class="flex-1 flex items-center justify-center">
    <div
      class="relative inline-flex items-center gap-1 bg-[#1a1a1a] rounded-full px-1 border border-white/[0.06]"
    >
      <div
        class={navBase}
        style="top:{navTop}px; transform:translateX({navIndex * navStep}px);"
      ></div>
      <!-- Inbox (Home) -->
      <button
        type="button"
        onclick={() => handleTabClick('inbox')}
        class={cn(
          'relative z-10 w-14 h-14 rounded-full flex items-center justify-center transition-all duration-200',
          activeTab === 'inbox'
            ? 'text-white'
            : 'text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]'
        )}
        aria-label="Inbox"
      >
        <InboxIcon class="w-6 h-6" strokeWidth={activeTab === 'inbox' ? 2.5 : 2} />
        {#if taskStore.reviewCount > 0}
          <div
            class="absolute top-2 right-2 flex min-w-[16px] h-4 items-center justify-center rounded-full bg-violet-500 px-1 text-[10px] font-bold text-white shadow-sm ring-2 ring-[#1a1a1a]"
          >
            {taskStore.reviewCount}
          </div>
        {/if}
      </button>

      <!-- Sessions -->
      <button
        type="button"
        onclick={() => handleTabClick('sessions')}
        class={cn(
          'relative z-10 w-14 h-14 rounded-full flex items-center justify-center transition-all duration-200',
          activeTab === 'sessions'
            ? 'text-white'
            : 'text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]'
        )}
        aria-label="Sessions"
      >
        <MessageSquareIcon class="w-6 h-6" strokeWidth={activeTab === 'sessions' ? 2.5 : 2} />
      </button>

      <!-- Projects -->
      <!-- <button -->
      <!--   type="button" -->
      <!--   onclick={() => handleTabClick("projects")} -->
      <!--   class={cn( -->
      <!--     "relative z-10 w-14 h-14 rounded-full flex items-center justify-center transition-all duration-200", -->
      <!--     activeTab === "projects" -->
      <!--       ? "text-white" -->
      <!--       : "text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]", -->
      <!--   )} -->
      <!--   aria-label="Projects" -->
      <!-- > -->
      <!--   <FolderIcon -->
      <!--     class="w-6 h-6" -->
      <!--     strokeWidth={activeTab === "projects" ? 2.5 : 2} -->
      <!--   /> -->
      <!-- </button> -->

      <!-- Search -->
      <button
        type="button"
        onclick={() => {
          handleTabClick('focus');
          if (onSearch) onSearch();
        }}
        class={cn(
          'relative z-10 w-14 h-14 rounded-full flex items-center justify-center transition-all duration-200',
          activeTab === 'focus'
            ? 'text-white'
            : 'text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]'
        )}
        aria-label="Search"
      >
        <SearchIcon class="w-6 h-6" strokeWidth={activeTab === 'focus' ? 2.5 : 2} />
      </button>

      <!-- Layers -->
      <button
        type="button"
        onclick={() => handleTabClick('layers')}
        class={cn(
          'relative z-10 w-14 h-14 rounded-full flex items-center justify-center transition-all duration-200',
          activeTab === 'layers'
            ? 'text-white'
            : 'text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]'
        )}
        aria-label="Layers"
      >
        <LayersIcon class="w-6 h-6" strokeWidth={activeTab === 'layers' ? 2.5 : 2} />
      </button>
    </div>

    <!-- New Task (Pen) -->

    <div
      class="ml-4 inline-flex items-center gap-1 bg-[#1a1a1a] rounded-full border border-white/[0.06]"
    >
      <button
        type="button"
        onclick={() => appState.toggleChatInput()}
        class="relative z-10 w-14 h-14 rounded-full flex items-center justify-center text-gray-500 hover:text-gray-300 hover:bg-white/[0.04] transition-all duration-200"
        aria-label="New Task"
      >
        <SquarePen class="w-6 h-6" strokeWidth={2} />
      </button>
    </div>
  </div>
</div>
