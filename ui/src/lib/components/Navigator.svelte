<script lang="ts">
  import { cn } from '$lib/utils';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import InboxIcon from '@lucide/svelte/icons/inbox';
  import { CogIcon, MessagesSquareIcon, SquarePen } from '@lucide/svelte';
  import FolderIcon from '@lucide/svelte/icons/folder';
  import MessageSquareIcon from '@lucide/svelte/icons/message-square';
  import LayersIcon from '@lucide/svelte/icons/layers';
  import SearchIcon from '@lucide/svelte/icons/search';
  import { appState } from '$lib/stores/app.svelte';
  import { taskStore } from '$lib/stores/tasks.svelte';

  interface Props {
    onSearch?: () => void;
  }

  let { onSearch }: Props = $props();

  // Derive activeTab from URL, with fallback to appState for non-URL tabs like focus
  const activeTab = $derived.by((): 'inbox' | 'focus' | 'settings' => {
    const path = $page.url.pathname;
    if (path.startsWith('/app/settings')) return 'settings';
    if (path === '/app' || path === '/app/') return 'inbox';
    // For focus (search), it doesn't have a dedicated route yet
    return 'inbox';
  });

  const tabs = ['inbox', 'focus', 'settings'];
  const activeIndex = $derived(tabs.indexOf(activeTab));
  const navIndex = $derived(activeIndex === -1 ? 0 : activeIndex);
  const navButtonSize = 64;
  const navBaseSize = 56;
  const navGap = 6;
  const navTop = (navButtonSize - navBaseSize) / 2 - 1;

  function handleTabClick(tab: string) {
    // Navigate to the appropriate URL
    switch (tab) {
      case 'inbox':
        goto('/app');
        break;
      case 'settings':
        goto('/app/settings');
        break;
      case 'focus':
        // Focus/Search doesn't have a dedicated page yet, just update state
        appState.activeTab = 'focus';
        break;
      default:
        goto('/app');
    }
  }

  const navBase =
    'absolute h-12 w-16 bg-[#2a2a2a] rounded-full transition-all gap-1 border border-white/[0.01]';
</script>

<div class="flex items-center justify-center w-full scale-105 sm:scale-100">
  <!-- Main Navigation Pill -->
  <div class="flex-1 flex items-center justify-center">
    <div
      class="relative inline-flex items-center gap-1 bg-[#1a1a1a] rounded-full px-1 border border-white/[0.06] shadow-2xl h-14"
    >
      <div
        class={navBase}
        style="top:{navTop}px; transform:translateX({navIndex * (navButtonSize + navGap - 15)}px);"
      ></div>

      <!-- Sessions -->
      <!-- <button -->
      <!--   type="button" -->
      <!--   onclick={() => handleTabClick('sessions')} -->
      <!--   class={cn( -->
      <!--     'relative z-10 w-16 h-16 rounded-full flex items-center justify-center transition-all duration-200', -->
      <!--     activeTab === 'sessions' -->
      <!--       ? 'text-white' -->
      <!--       : 'text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]' -->
      <!--   )} -->
      <!--   aria-label="Sessions" -->
      <!-- > -->
      <!--   <MessagesSquareIcon class="w-7 h-7" strokeWidth={activeTab === 'sessions' ? 2.5 : 2} /> -->
      <!-- </button> -->

      <!-- Inbox (Home) -->
      <button
        type="button"
        onclick={() => handleTabClick('inbox')}
        class={cn(
          'relative z-10 w-16 h-16 rounded-full flex items-center justify-center transition-all duration-200',
          activeTab === 'inbox'
            ? 'text-white'
            : 'text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]'
        )}
        aria-label="Inbox"
      >
        <InboxIcon class="w-7 h-7" strokeWidth={activeTab === 'inbox' ? 2.5 : 2} />
        {#if taskStore.reviewCount > 0}
          <div
            class="absolute top-2.5 right-2.5 flex min-w-[18px] h-[18px] items-center justify-center rounded-full bg-violet-500 px-1 text-[10px] font-bold text-white shadow-sm ring-2 ring-[#1a1a1a]"
          >
            {taskStore.reviewCount}
          </div>
        {/if}
      </button>

      <!-- Search -->
      <button
        type="button"
        onclick={() => {
          handleTabClick('focus');
          if (onSearch) onSearch();
        }}
        class={cn(
          'relative z-10 w-16 h-16 rounded-full flex items-center justify-center transition-all duration-200',
          activeTab === 'focus'
            ? 'text-white'
            : 'text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]'
        )}
        aria-label="Search"
      >
        <SearchIcon class="w-7 h-7" strokeWidth={activeTab === 'focus' ? 2.5 : 2} />
      </button>

      <!-- Settings -->
      <button
        type="button"
        onclick={() => handleTabClick('settings')}
        class={cn(
          'relative z-10 w-16 h-16 rounded-full flex items-center justify-center transition-all duration-200',
          activeTab === 'settings'
            ? 'text-white'
            : 'text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]'
        )}
        aria-label="Layers"
      >
        <CogIcon class="w-7 h-7" strokeWidth={activeTab === 'settings' ? 2.5 : 2} />
      </button>

      <!-- <!-- Layers -->
      <!-- <button -->
      <!--   type="button" -->
      <!--   onclick={() => handleTabClick('layers')} -->
      <!--   class={cn( -->
      <!--     'relative z-10 w-16 h-16 rounded-full flex items-center justify-center transition-all duration-200', -->
      <!--     activeTab === 'layers' -->
      <!--       ? 'text-white' -->
      <!--       : 'text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]' -->
      <!--   )} -->
      <!--   aria-label="Layers" -->
      <!-- > -->
      <!--   <LayersIcon class="w-7 h-7" strokeWidth={activeTab === 'layers' ? 2.5 : 2} /> -->
      <!-- </button> -->
    </div>

    <!-- New Task (Pen) -->

    <div
      class="ml-4 inline-flex items-center gap-1.5 bg-[#1a1a1a] rounded-full border border-white/[0.06] shadow-2xl h-14"
    >
      <button
        type="button"
        onclick={() => appState.toggleNewTaskModal()}
        class="relative z-10 w-16 h-14 rounded-full flex items-center justify-center text-gray-500 hover:text-gray-300 hover:bg-white/[0.04] transition-all duration-200"
        aria-label="New Task"
      >
        <SquarePen class="w-7 h-7" strokeWidth={2} />
      </button>
    </div>
  </div>
</div>
