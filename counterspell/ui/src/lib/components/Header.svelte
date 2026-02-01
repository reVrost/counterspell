<script lang="ts">
  import { appState } from "$lib/stores/app.svelte";
  import { cn, getInitial } from "$lib/utils";
  import SettingsIcon from "@lucide/svelte/icons/settings";
  import DownloadIcon from "@lucide/svelte/icons/download";
  import RefreshCwIcon from "@lucide/svelte/icons/refresh-cw";
  import LogOutIcon from "@lucide/svelte/icons/log-out";
  import CheckIcon from "@lucide/svelte/icons/check";
  import SearchIcon from "@lucide/svelte/icons/search";
  import PlusIcon from "@lucide/svelte/icons/plus";
  import * as DropdownMenu from "$lib/components/ui/dropdown-menu";
  import Title from "./ui/title/title.svelte";

  let projectSearch = $state("");
  let { activeTab } = $props();

  // const activeTabToIcon = {
  //   inbox: Inbox,
  //   projects: Folder,
  //   search: Search,
  //   layers: Layers,
  //   new: Newspaper,
  // };
  // let CurrentIcon = (activeTabToIcon[activeTab] as any) ?? null;

  // let CurrentIcon = null;

  const filteredProjects = $derived(
    appState.projects.filter((p) =>
      p.name.toLowerCase().includes(projectSearch.toLowerCase()),
    ),
  );

  async function handleSignOut() {
    await appState.logout();
  }

  async function handleSettings() {
    appState.openSettings();
  }

  let syncing = $state(false);
  async function handleSyncRepos() {
    syncing = true;
    try {
      const res = await fetch("/api/v1/github/sync", { method: "POST" });
      if (res.ok) {
        // Refresh the feed to get updated projects
        // await appState.fetchFeed();
      }
    } catch (e) {
      console.error("Failed to sync repos:", e);
    } finally {
      syncing = false;
    }
  }
</script>

<header
  class="h-14 border-linear-border backdrop-blur-md flex items-center justify-between px-4 z-20 shrink-0 fixed top-0 left-0 right-0 border-b border-white/[0.06] shadow-[0_1px_0_0_rgba(0,0,0,0.3)]"
>
  <!-- Project Selector -->
  <DropdownMenu.Root bind:open={appState.projectMenuOpen}>
    <DropdownMenu.Trigger
      class="flex items-center gap-2 cursor-pointer active:opacity-70 transition"
    >
      <img src="/favicon-16.png" alt="Counterspell" class="h-6 w-6" />
      <Title>
        {activeTab.charAt(0).toUpperCase() + activeTab.slice(1)}
      </Title>
    </DropdownMenu.Trigger>
    <DropdownMenu.Portal>
      <DropdownMenu.Content
        class="w-72 bg-popover border border-gray-700 rounded-xl shadow-[0_0_50px_rgba(0,0,0,0.5)] overflow-hidden flex flex-col mt-2 z-50"
        sideOffset={8}
      >
        <!-- Search Header -->
        <div class="p-3 border-b border-gray-700 bg-popover">
          <div class="relative">
            <SearchIcon
              class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500 w-3 h-3"
            />
            <input
              bind:value={projectSearch}
              type="text"
              placeholder="Filter repositories..."
              class="w-full bg-gray-900 border border-gray-700 rounded-lg pl-8 pr-3 py-1.5 text-sm text-white focus:outline-none focus:border-primary placeholder-gray-600 transition-colors"
            />
          </div>
        </div>

        <!-- Scrollable List -->
        <div class="max-h-[320px] overflow-y-auto py-1">
          <DropdownMenu.Item
            class="w-full px-4 py-2 hover:bg-white/5 cursor-pointer text-sm font-bold text-white border-b border-gray-800/50 mb-1 text-left focus:bg-white/5 outline-none"
          >
            All Projects
          </DropdownMenu.Item>

          {#each filteredProjects as p}
            <DropdownMenu.Item
              onSelect={() => appState.setActiveProject(p.id, p.name)}
              class={cn(
                "w-full px-4 py-2 hover:bg-white/5 cursor-pointer flex items-center gap-3 group transition text-left focus:bg-white/5 outline-none",
                appState.activeProjectId === p.id && "bg-white/5",
              )}
            >
              <div
                class="w-6 h-6 rounded bg-gray-800 border border-gray-700 flex items-center justify-center shrink-0"
              >
                <span class="text-[10px] {p.color}">
                  <i class="fas {p.icon}"></i>
                </span>
              </div>
              <div class="flex-1 min-w-0">
                <div
                  class="text-sm text-gray-400 group-hover:text-white truncate transition"
                >
                  {p.name}
                </div>
              </div>
              {#if appState.activeProjectId === p.id}
                <CheckIcon class="w-3 h-3 text-green-500" />
              {/if}
            </DropdownMenu.Item>
          {/each}

          {#if filteredProjects.length === 0}
            <div class="px-4 py-8 text-center text-gray-600 text-sm">
              No projects found.
            </div>
          {/if}
        </div>

        <!-- Footer -->
        <div
          class="px-3 py-2 bg-gray-900/50 border-t border-gray-800 text-[10px] text-gray-500 flex justify-between"
        >
          <span>{appState.projects.length} Repositories</span>
          <button
            class="hover:text-blue-400 cursor-pointer flex items-center gap-1"
          >
            <PlusIcon class="w-2.5 h-2.5" /> New
          </button>
        </div>
      </DropdownMenu.Content>
    </DropdownMenu.Portal>
  </DropdownMenu.Root>

  <!-- User Menu -->
  <DropdownMenu.Root>
    <DropdownMenu.Trigger
      class="flex items-center gap-3 cursor-pointer hover:opacity-80 transition p-1"
    >
      <div
        class="h-2 w-2 rounded-full bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.4)]"
      ></div>
      <div
        class="w-6 h-6 rounded-full bg-gray-800 border border-gray-700 flex items-center justify-center text-[10px] font-bold text-gray-300"
      >
        {getInitial(appState.githubLogin || appState.userEmail)}
      </div>
    </DropdownMenu.Trigger>
    <DropdownMenu.Portal>
      <DropdownMenu.Content
        class="w-56 bg-popover border border-gray-700 rounded-xl shadow-2xl overflow-hidden py-1 z-50"
        align="end"
        sideOffset={8}
      >
        <div class="px-4 py-3 border-b border-gray-800 mb-1">
          <p class="text-sm text-gray-500 uppercase tracking-wider font-bold">
            Signed in as
          </p>
          <p class="text-sm font-medium text-gray-200 mt-1 truncate">
            {appState.githubLogin || appState.userEmail}
          </p>
        </div>
        <DropdownMenu.Group class="px-2">
          <DropdownMenu.Item
            onSelect={() => (appState.settingsOpen = true)}
            class="w-full px-2 py-1.5 hover:bg-white/5 rounded text-sm text-gray-400 flex items-center gap-2 transition-colors text-left cursor-pointer focus:bg-white/5 outline-none"
          >
            <SettingsIcon class="w-4 h-4" /> Settings
          </DropdownMenu.Item>
          <DropdownMenu.Item
            onSelect={handleSyncRepos}
            disabled={syncing}
            class="w-full px-2 py-1.5 hover:bg-white/5 rounded text-sm text-gray-400 flex items-center gap-2 transition-colors text-left cursor-pointer focus:bg-white/5 outline-none disabled:opacity-50"
          >
            <RefreshCwIcon class="w-4 h-4 {syncing ? 'animate-spin' : ''}" />
            {syncing ? "Syncing..." : "Sync Repos"}
          </DropdownMenu.Item>
          {#if appState.canInstallPWA}
            <DropdownMenu.Item
              onSelect={() => appState.installPWA()}
              class="w-full px-2 py-1.5 hover:bg-purple-500/10 rounded text-sm text-purple-400 hover:text-purple-300 flex items-center gap-2 transition-colors text-left cursor-pointer focus:bg-purple-500/10 outline-none"
            >
              <DownloadIcon class="w-4 h-4" /> Install App
            </DropdownMenu.Item>
          {/if}
        </DropdownMenu.Group>
        <DropdownMenu.Separator class="h-px bg-gray-800 my-1 mx-2" />
        <DropdownMenu.Group class="px-2 pb-1">
          <DropdownMenu.Item
            onSelect={handleSignOut}
            class="w-full px-2 py-1.5 hover:bg-red-500/10 rounded text-sm text-red-400 hover:text-red-300 flex items-center gap-2 transition-colors text-left cursor-pointer focus:bg-red-500/10 outline-none"
          >
            <LogOutIcon class="w-4 h-4" /> Sign Out
          </DropdownMenu.Item>
        </DropdownMenu.Group>
      </DropdownMenu.Content>
    </DropdownMenu.Portal>
  </DropdownMenu.Root>
</header>
