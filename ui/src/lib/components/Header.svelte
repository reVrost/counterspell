<script lang="ts">
  import { appState } from '$lib/stores/app.svelte';
  import { cn, getInitial } from '$lib/utils';
  import SettingsIcon from '@lucide/svelte/icons/settings';
  import DownloadIcon from '@lucide/svelte/icons/download';
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
  import LogOutIcon from '@lucide/svelte/icons/log-out';
  import CheckIcon from '@lucide/svelte/icons/check';
  import SearchIcon from '@lucide/svelte/icons/search';
  import PlusIcon from '@lucide/svelte/icons/plus';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
  import Logo from './Logo.svelte';

  let projectSearch = $state('');
  let userMenuOpen = $state(false);
  let avatarImageFailed = $state(false);
  let signingOut = $state(false);
  let { activeTab } = $props();

  const filteredProjects = $derived(
    appState.projects.filter((p) => p.name.toLowerCase().includes(projectSearch.toLowerCase()))
  );
  const accountName = $derived(appState.githubLogin || appState.userEmail);
  const avatarSrc = $derived(
    appState.githubLogin
      ? `https://avatars.githubusercontent.com/${appState.githubLogin}?size=64`
      : ''
  );

  $effect(() => {
    appState.githubLogin;
    avatarImageFailed = false;
  });

  async function handleSignOut() {
    if (signingOut) return;
    signingOut = true;
    try {
      await appState.logout();
    } catch (e) {
      signingOut = false;
      console.error('Failed to sign out:', e);
    }
  }

  let syncing = $state(false);
  async function handleSyncRepos() {
    syncing = true;
    try {
      const res = await fetch('/api/v1/github/sync', { method: 'POST' });
      if (res.ok) {
        // Refresh or notify
      }
    } catch (e) {
      console.error('Failed to sync repos:', e);
    } finally {
      syncing = false;
    }
  }
</script>

<header
  class="h-16 flex items-center justify-between px-6 z-30 shrink-0 fixed top-0 left-0 right-0 backdrop-blur-lg bg-zinc-950/30 transition-all duration-300"
>
  <!-- Left: Workspace Selector -->
  <div class="flex items-center">
    <DropdownMenu.Root bind:open={appState.projectMenuOpen}>
      <DropdownMenu.Trigger
        class="flex items-center gap-1.5 cursor-pointer group hover:bg-white/[0.04] active:bg-white/[0.06] px-2 py-1.5 rounded-xl transition-all duration-200 outline-none"
      >
        <Logo class="w-6 h-6" />
        <span
          class="text-lg font-semibold tracking-tight text-white/90 group-hover:text-white transition-colors"
        >
          {appState.activeWorkspaceName || 'All Workspaces'}
        </span>
        <ChevronDownIcon
          class="w-4 h-4 text-zinc-500 group-hover:text-zinc-300 transition-colors"
        />
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content
          class="w-72 bg-zinc-950/95 backdrop-blur-xl border border-white/10 rounded-2xl shadow-[0_20px_50px_rgba(0,0,0,0.5)] overflow-hidden flex flex-col mt-2 ml-6 z-50 p-1"
          sideOffset={8}
        >
          <!-- Search Header -->
          <div class="p-2">
            <div class="relative">
              <SearchIcon
                class="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500 w-3.5 h-3.5"
              />
              <input
                bind:value={projectSearch}
                type="text"
                placeholder="Filter workspaces..."
                class="w-full bg-white/[0.03] border border-white/5 rounded-xl pl-9 pr-3 py-2 text-sm text-zinc-200 focus:outline-none focus:border-violet-500/50 focus:bg-white/[0.05] placeholder-zinc-600 transition-all"
              />
            </div>
          </div>

          <!-- Scrollable List -->
          <div class="max-h-[320px] overflow-y-auto py-1 custom-scrollbar">
            <DropdownMenu.Item
              onSelect={() => appState.setActiveWorkspace('', '')}
              class={cn(
                'w-full px-3 py-2.5 hover:bg-white/5 cursor-pointer rounded-lg flex items-center justify-between group transition text-left focus:bg-white/5 outline-none mb-0.5',
                !appState.activeWorkspaceId && 'bg-white/[0.08] text-white'
              )}
            >
              <span
                class="text-sm {appState.activeWorkspaceId
                  ? 'text-zinc-400'
                  : 'text-zinc-100'} group-hover:text-zinc-100 transition">All Workspaces</span
              >
              {#if !appState.activeWorkspaceId}
                <CheckIcon class="w-3.5 h-3.5 text-violet-400" />
              {/if}
            </DropdownMenu.Item>

            {#each filteredProjects as p}
              <DropdownMenu.Item
                onSelect={() => appState.setActiveWorkspace(p.id, p.name)}
                class={cn(
                  'w-full px-3 py-2.5 hover:bg-white/5 cursor-pointer rounded-lg flex items-center gap-3 group transition text-left focus:bg-white/5 outline-none mb-0.5',
                  appState.activeWorkspaceId === p.id && 'bg-white/[0.08] text-white'
                )}
              >
                <div
                  class="w-6 h-6 rounded-md bg-zinc-900 border border-white/5 flex items-center justify-center shrink-0"
                >
                  <span class="text-[10px] {p.color}">
                    <i class="fas {p.icon}"></i>
                  </span>
                </div>
                <div class="flex-1 min-w-0">
                  <div class="text-sm text-zinc-400 group-hover:text-zinc-100 truncate transition">
                    {p.name}
                  </div>
                </div>
                {#if appState.activeWorkspaceId === p.id}
                  <CheckIcon class="w-3.5 h-3.5 text-violet-400" />
                {/if}
              </DropdownMenu.Item>
            {/each}

            {#if filteredProjects.length === 0}
              <div class="px-4 py-8 text-center text-zinc-600 text-sm">No workspaces found.</div>
            {/if}
          </div>

          <!-- Footer -->
          <div
            class="mt-1 px-3 py-2 bg-white/[0.02] border-t border-white/5 text-sm text-zinc-500 flex justify-between items-center rounded-b-xl"
          >
            <span class="font-medium">{appState.projects.length} Workspaces</span>
            <button
              class="hover:text-violet-400 font-semibold cursor-pointer flex items-center gap-1 transition-colors"
            >
              <PlusIcon class="w-3 h-3" /> New
            </button>
          </div>
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  </div>

  <!-- Right: Actions & User -->
  <div class="flex items-center gap-2 pointer-events-auto">
    {#if appState.canInstallPWA}
      <button
        onclick={() => appState.installPWA()}
        class="flex items-center gap-2 px-3 py-1.5 rounded-full bg-white/[0.05] hover:bg-white/[0.08] border border-white/[0.05] transition-all group mr-1"
      >
        <DownloadIcon
          class="w-3.5 h-3.5 text-violet-400 group-hover:scale-110 transition-transform"
        />
        <span class="text-[11px] font-semibold text-zinc-400 group-hover:text-zinc-200"
          >Install App</span
        >
      </button>
    {/if}

    <DropdownMenu.Root bind:open={userMenuOpen}>
      <DropdownMenu.Trigger
        class="group flex items-center gap-2.5 cursor-pointer hover:bg-white/[0.04] active:bg-white/[0.06] p-1 pr-3 rounded-full transition-all outline-none border border-transparent hover:border-white/5"
        aria-label="Open user menu"
      >
        <div class="relative group">
          <div
            class="absolute pointer-events-none -inset-0.5 bg-gradient-to-tr from-violet-600 to-pink-600 rounded-full opacity-0 group-hover:opacity-40 blur-sm transition-opacity"
          ></div>

          <div
            class="w-7 h-7 rounded-full border border-white/10 relative z-10 bg-zinc-900 shadow-xl overflow-hidden flex items-center justify-center"
          >
            {#if appState.githubLogin && !avatarImageFailed}
              <img
                src={avatarSrc}
                alt={appState.githubLogin}
                class="w-full h-full object-cover animate-in fade-in duration-300"
                draggable="false"
                onerror={() => {
                  avatarImageFailed = true;
                }}
              />
            {:else if accountName}
              <div
                class="w-full h-full flex items-center justify-center text-[10px] font-bold text-violet-400 animate-in fade-in duration-300"
              >
                {getInitial(accountName)}
              </div>
            {:else}
              <div class="w-full h-full bg-zinc-800 animate-pulse"></div>
            {/if}
          </div>

          <div
            class="absolute pointer-events-none -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full bg-emerald-500 border-2 border-zinc-950 z-20 shadow-[0_0_8px_rgba(16,185,129,0.5)] scale-75"
          ></div>
        </div>
        <ChevronDownIcon
          class={cn(
            'w-3.5 h-3.5 text-zinc-500 group-hover:text-zinc-300 transition-all duration-300',
            userMenuOpen && 'rotate-180 text-zinc-300'
          )}
        />
      </DropdownMenu.Trigger>

      <DropdownMenu.Portal>
        <DropdownMenu.Content
          class="w-60 bg-zinc-950/95 backdrop-blur-xl border border-white/10 rounded-2xl shadow-[0_20px_50px_rgba(0,0,0,0.5)] overflow-hidden py-1.5 z-50"
          align="end"
          sideOffset={8}
        >
          <div class="px-4 py-3 border-b border-white/5 mb-1 bg-white/[0.02]">
            <p class="text-[10px] text-zinc-500 uppercase tracking-widest font-black">Account</p>
            <p class="text-sm font-semibold text-zinc-100 mt-1 truncate">
              {accountName}
            </p>
          </div>
          <DropdownMenu.Group class="px-1.5">
            <DropdownMenu.Item
              onSelect={() => (appState.settingsOpen = true)}
              class="w-full px-2.5 py-2 hover:bg-white/5 rounded-lg text-sm text-zinc-400 flex items-center gap-3 transition-colors text-left cursor-pointer focus:bg-white/5 outline-none"
            >
              <SettingsIcon class="w-4 h-4" /> Settings
            </DropdownMenu.Item>
            <DropdownMenu.Item
              onSelect={handleSyncRepos}
              disabled={syncing}
              class="w-full px-2.5 py-2 hover:bg-white/5 rounded-lg text-sm text-zinc-400 flex items-center gap-3 transition-colors text-left cursor-pointer focus:bg-white/5 outline-none disabled:opacity-50"
            >
              <RefreshCwIcon class="w-4 h-4 {syncing ? 'animate-spin' : ''}" />
              {syncing ? 'Syncing...' : 'Sync Repos'}
            </DropdownMenu.Item>
          </DropdownMenu.Group>
          <DropdownMenu.Separator class="h-px bg-white/5 my-1.5 mx-2" />
          <DropdownMenu.Group class="px-1.5 pb-1">
            <DropdownMenu.Item
              onSelect={handleSignOut}
              disabled={signingOut}
              data-testid="logout-button"
              class="w-full px-2.5 py-2 hover:bg-red-500/10 rounded-lg text-sm text-red-400 hover:text-red-300 flex items-center gap-3 transition-colors text-left cursor-pointer focus:bg-red-500/10 outline-none disabled:opacity-50"
            >
              <LogOutIcon class="w-4 h-4" />
              {signingOut ? 'Logging out...' : 'Logout'}
            </DropdownMenu.Item>
          </DropdownMenu.Group>
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  </div>
</header>

<style>
  .custom-scrollbar::-webkit-scrollbar {
    width: 4px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.05);
    border-radius: 10px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: rgba(255, 255, 255, 0.1);
  }
</style>
