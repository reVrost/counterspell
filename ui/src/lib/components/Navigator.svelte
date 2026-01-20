<script lang="ts">
  import { cn } from "$lib/utils";
  import InboxIcon from "@lucide/svelte/icons/inbox";
  import { SquarePen } from "@lucide/svelte";
  import FolderIcon from "@lucide/svelte/icons/folder";
  import LayersIcon from "@lucide/svelte/icons/layers";
  import SearchIcon from "@lucide/svelte/icons/search";
  import { appState } from "$lib/stores/app.svelte";

  interface Props {
    activeTab?: "inbox" | "projects" | "focus" | "layers";
    onNavigate?: (tab: string) => void;
    onSearch?: () => void;
  }

  let { activeTab = "inbox", onNavigate, onSearch }: Props = $props();

  const tabs = ["inbox", "projects", "search", "layers", "new"];
  const currentTab = $derived(
    activeTab === "focus" ? "search" : activeTab || "inbox",
  );
  const activeIndex = $derived(tabs.indexOf(currentTab));

  function handleTabClick(tab: string) {
    if (onNavigate) {
      onNavigate(tab);
    }
  }
</script>

<div class="flex items-center justify-center w-full">
  <!-- Main Navigation Pill -->
  <div
    class="relative inline-flex items-center gap-1 bg-[#1a1a1a] rounded-full p-1.5 border border-white/[0.06]"
  >
    class="absolute h-10 w-12 bg-[#2a2a2a] rounded-full transition-all
    duration-300 ease-out shadow-lg shadow-black/20" style="transform:
    translateX({(activeIndex === -1 ? 0 : activeIndex) * 52}px); left: 6px;" >
  </div>

  <!-- Inbox (Home) -->
  <button
    type="button"
    onclick={() => handleTabClick("inbox")}
    class={cn(
      "relative z-10 w-12 h-10 rounded-full flex items-center justify-center transition-all duration-200",
      activeTab === "inbox"
        ? "text-white"
        : "text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]",
    )}
    aria-label="Inbox"
  >
    <InboxIcon class="w-5 h-5" strokeWidth={activeTab === "inbox" ? 2.5 : 2} />
  </button>

  <!-- Projects -->
  <button
    type="button"
    onclick={() => handleTabClick("projects")}
    class={cn(
      "relative z-10 w-12 h-10 rounded-full flex items-center justify-center transition-all duration-200",
      activeTab === "projects"
        ? "text-white"
        : "text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]",
    )}
    aria-label="Projects"
  >
    <FolderIcon
      class="w-5 h-5"
      strokeWidth={activeTab === "projects" ? 2.5 : 2}
    />
  </button>

  <!-- Search -->
  <button
    type="button"
    onclick={() => {
      handleTabClick("focus");
      if (onSearch) onSearch();
    }}
    class={cn(
      "relative z-10 w-12 h-10 rounded-full flex items-center justify-center transition-all duration-200",
      activeTab === "focus"
        ? "text-white"
        : "text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]",
    )}
    aria-label="Search"
  >
    <SearchIcon class="w-5 h-5" strokeWidth={activeTab === "focus" ? 2.5 : 2} />
  </button>

  <!-- Layers -->
  <button
    type="button"
    onclick={() => handleTabClick("layers")}
    class={cn(
      "relative z-10 w-12 h-10 rounded-full flex items-center justify-center transition-all duration-200",
      activeTab === "layers"
        ? "text-white"
        : "text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]",
    )}
    aria-label="Layers"
  >
    <LayersIcon
      class="w-5 h-5"
      strokeWidth={activeTab === "layers" ? 2.5 : 2}
    />
  </button>

  <!-- New Task (Pen) -->
  <button
    type="button"
    onclick={() => appState.toggleChatInput()}
    class="relative z-10 w-12 h-10 rounded-full flex items-center justify-center text-gray-500 hover:text-gray-300 hover:bg-white/[0.04] transition-all duration-200"
    aria-label="New Task"
  >
    <SquarePen class="w-5 h-5" strokeWidth={2} />
  </button>
</div>
