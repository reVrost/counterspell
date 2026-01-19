<script lang="ts">
    import { cn } from "$lib/utils";
    import InboxIcon from "@lucide/svelte/icons/inbox";
    import FolderIcon from "@lucide/svelte/icons/folder";
    import FocusIcon from "@lucide/svelte/icons/focus";
    import LayersIcon from "@lucide/svelte/icons/layers";
    import SearchIcon from "@lucide/svelte/icons/search";

    interface Props {
        activeTab?: "inbox" | "projects" | "focus" | "layers";
        onNavigate?: (tab: string) => void;
        onSearch?: () => void;
    }

    let { activeTab = "inbox", onNavigate, onSearch }: Props = $props();

    function handleTabClick(tab: string) {
        if (onNavigate) {
            onNavigate(tab);
        }
    }
</script>

<div class="flex items-center gap-3">
    <!-- Main Navigation Pill -->
    <div class="flex-1 flex items-center justify-center">
        <div
            class="inline-flex items-center gap-1 bg-[#1a1a1a] rounded-full p-1.5 border border-white/[0.06]"
        >
            <!-- Inbox (Home) -->
            <button
                type="button"
                onclick={() => handleTabClick("inbox")}
                class={cn(
                    "w-12 h-12 rounded-full flex items-center justify-center transition-all duration-200",
                    activeTab === "inbox"
                        ? "bg-[#2a2a2a] text-white shadow-lg"
                        : "text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]",
                )}
                aria-label="Inbox"
            >
                <InboxIcon
                    class="w-5 h-5"
                    strokeWidth={activeTab === "inbox" ? 2.5 : 2}
                />
            </button>

            <!-- Projects -->
            <button
                type="button"
                onclick={() => handleTabClick("projects")}
                class={cn(
                    "w-12 h-12 rounded-full flex items-center justify-center transition-all duration-200",
                    activeTab === "projects"
                        ? "bg-[#2a2a2a] text-white shadow-lg"
                        : "text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]",
                )}
                aria-label="Projects"
            >
                <FolderIcon
                    class="w-5 h-5"
                    strokeWidth={activeTab === "projects" ? 2.5 : 2}
                />
            </button>

            <!-- Focus -->
            <button
                type="button"
                onclick={() => handleTabClick("focus")}
                class={cn(
                    "w-12 h-12 rounded-full flex items-center justify-center transition-all duration-200",
                    activeTab === "focus"
                        ? "bg-[#2a2a2a] text-white shadow-lg"
                        : "text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]",
                )}
                aria-label="Focus"
            >
                <FocusIcon
                    class="w-5 h-5"
                    strokeWidth={activeTab === "focus" ? 2.5 : 2}
                />
            </button>

            <!-- Layers -->
            <button
                type="button"
                onclick={() => handleTabClick("layers")}
                class={cn(
                    "w-12 h-12 rounded-full flex items-center justify-center transition-all duration-200",
                    activeTab === "layers"
                        ? "bg-[#2a2a2a] text-white shadow-lg"
                        : "text-gray-500 hover:text-gray-300 hover:bg-white/[0.04]",
                )}
                aria-label="Layers"
            >
                <LayersIcon
                    class="w-5 h-5"
                    strokeWidth={activeTab === "layers" ? 2.5 : 2}
                />
            </button>
        </div>
    </div>

    <!-- Search Button (Separate) -->
    <button
        type="button"
        onclick={onSearch}
        class="w-14 h-14 rounded-full bg-[#1a1a1a] border border-white/[0.06] flex items-center justify-center text-gray-500 hover:text-gray-300 hover:bg-[#222] transition-all duration-200"
        aria-label="Search"
    >
        <SearchIcon class="w-5 h-5" strokeWidth={2} />
    </button>
</div>
