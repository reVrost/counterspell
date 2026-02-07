<script lang="ts">
  import { appState } from '$lib/stores/app.svelte';
  import { workspacesAPI } from '$lib/api';
  import { backdropFade, modalSlideUp } from '$lib/utils/transitions';
  import FolderIcon from '@lucide/svelte/icons/folder';
  import FolderPlusIcon from '@lucide/svelte/icons/folder-plus';
  import LoaderIcon from '@lucide/svelte/icons/loader-2';
  import XIcon from '@lucide/svelte/icons/x';

  type WorkspaceMode = 'import_current' | 'new_folder';

  let workspaceName = $state('');
  let workspaceMode = $state<WorkspaceMode>('import_current');
  let cwdPath = $state('');
  let cwdName = $state('');
  let loadingSetup = $state(false);
  let creating = $state(false);

  const resolvedPath = $derived.by(() => {
    if (!cwdPath) return '';
    if (workspaceMode === 'import_current') return cwdPath;
    const trimmedName = workspaceName.trim();
    if (!trimmedName) return cwdPath;

    const separator = cwdPath.includes('\\') && !cwdPath.includes('/') ? '\\' : '/';
    const base = cwdPath.endsWith('/') || cwdPath.endsWith('\\') ? cwdPath.slice(0, -1) : cwdPath;
    return `${base}${separator}${trimmedName}`;
  });

  $effect(() => {
    if (appState.showNewWorkspaceModal) {
      void loadSetup();
    }
  });

  async function loadSetup() {
    loadingSetup = true;
    try {
      const setup = await workspacesAPI.getSetup();
      cwdPath = setup.cwd_path;
      cwdName = setup.cwd_name;
      workspaceName = setup.cwd_name;
      workspaceMode = 'import_current';
    } catch (e) {
      console.error('Failed to load workspace setup:', e);
      appState.showToast(e instanceof Error ? e.message : 'Failed to load workspace info', 'error');
    } finally {
      loadingSetup = false;
    }
  }

  function close(e: any) {
    e.stopPropagation();

    if (creating) return;
    appState.closeNewWorkspaceModal();
  }

  function onModeChange(mode: WorkspaceMode) {
    workspaceMode = mode;
    if (mode === 'import_current' && !workspaceName.trim()) {
      workspaceName = cwdName;
    }
  }

  async function createWorkspace(e: any) {
    const name = workspaceName.trim();
    if (!name) {
      appState.showToast('Workspace name is required', 'error');
      return;
    }

    creating = true;
    try {
      const workspace = await workspacesAPI.create({
        name,
        mode: workspaceMode,
      });
      await appState.loadProjects();
      await appState.setActiveWorkspace(workspace.id, workspace.name);
      appState.showToast(`Workspace "${workspace.name}" created`, 'success');
      close(e);
    } catch (e) {
      console.error('Failed to create workspace:', e);
      appState.showToast(e instanceof Error ? e.message : 'Failed to create workspace', 'error');
    } finally {
      creating = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      close(e);
      return;
    }
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      void createWorkspace(e);
    }
  }
</script>

<div
  class="fixed inset-0 z-[70] flex items-end sm:items-center justify-center pointer-events-none"
  role="dialog"
  aria-modal="true"
  tabindex="-1"
  onkeydown={handleKeydown}
>
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <!-- <div -->
  <!--   transition:backdropFade -->
  <!--   onclick={close} -->
  <!--   class="absolute inset-0 bg-black/80 backdrop-blur-sm pointer-events-auto" -->
  <!-- ></div> -->

  <div
    transition:modalSlideUp
    class="pointer-events-auto relative w-full h-[100dvh] bg-[#0C0E12] flex flex-col sm:h-auto sm:max-w-xl sm:rounded-2xl sm:border sm:border-white/10 shadow-2xl overflow-hidden"
  >
    <div class="px-4 py-3 border-b border-white/5 flex items-center justify-between shrink-0">
      <h2 class="text-xl font-semibold text-zinc-100">New Workspace</h2>
      <button
        onclick={close}
        class="w-8 h-8 rounded-full flex items-center justify-center text-gray-500 hover:text-white hover:bg-white/10 transition"
      >
        <XIcon class="w-5 h-5" />
      </button>
    </div>

    <div class="flex-1 overflow-y-auto px-5 py-5 space-y-5">
      <div class="space-y-2">
        <label for="workspace-name" class="text-base text-zinc-300">Workspace Name</label>
        <input
          id="workspace-name"
          type="text"
          bind:value={workspaceName}
          placeholder="workspace-name"
          class="w-full bg-white/[0.03] border border-white/10 rounded-xl px-3 py-2.5 text-base text-zinc-100 focus:outline-none focus:border-violet-500/60"
        />
      </div>

      <div class="space-y-2">
        <p class="text-base text-zinc-300">Workspace Path</p>

        <button
          type="button"
          onclick={() => onModeChange('import_current')}
          class="w-full rounded-xl border px-3 py-3 text-left transition flex gap-3 items-start {workspaceMode ===
          'import_current'
            ? 'border-violet-400/50 bg-violet-500/10'
            : 'border-white/10 bg-white/[0.02] hover:bg-white/[0.04]'}"
        >
          <FolderIcon class="w-4 h-4 mt-0.5 text-zinc-300" />
          <div class="min-w-0">
            <p class="text-base font-medium text-zinc-100">Import current workspace</p>
            <p class="text-base text-zinc-400 break-all">
              Use the running app path: {cwdPath || 'Loading...'}
            </p>
          </div>
        </button>

        <button
          type="button"
          onclick={() => onModeChange('new_folder')}
          class="w-full rounded-xl border px-3 py-3 text-left transition flex gap-3 items-start {workspaceMode ===
          'new_folder'
            ? 'border-violet-400/50 bg-violet-500/10'
            : 'border-white/10 bg-white/[0.02] hover:bg-white/[0.04]'}"
        >
          <FolderPlusIcon class="w-4 h-4 mt-0.5 text-zinc-300" />
          <div class="min-w-0">
            <p class="text-base font-medium text-zinc-100">New folder</p>
            <p class="text-base text-zinc-400 break-all">
              Create a folder under current path using the workspace name.
            </p>
          </div>
        </button>
      </div>

      <div class="space-y-2">
        <label for="workspace-path" class="text-base text-zinc-300">Resolved Path</label>
        <input
          id="workspace-path"
          type="text"
          value={resolvedPath}
          readonly
          class="w-full bg-white/[0.02] border border-white/10 rounded-xl px-3 py-2.5 text-base text-zinc-400 focus:outline-none"
        />
      </div>
    </div>

    <div class="px-5 py-4 border-t border-white/5 flex items-center justify-end gap-2 shrink-0">
      <button
        type="button"
        onclick={close}
        class="px-4 py-2 rounded-xl border border-white/10 text-base text-zinc-200 hover:bg-white/5 transition"
      >
        Cancel
      </button>
      <button
        type="button"
        onclick={createWorkspace}
        disabled={creating || loadingSetup || !workspaceName.trim() || !resolvedPath}
        class="px-4 py-2 rounded-xl text-base font-semibold text-white transition flex items-center gap-2 {creating ||
        loadingSetup ||
        !workspaceName.trim() ||
        !resolvedPath
          ? 'bg-violet-900/50 cursor-not-allowed'
          : 'bg-violet-600 hover:bg-violet-500'}"
      >
        {#if creating}
          <LoaderIcon class="w-4 h-4 animate-spin" />
          Creating...
        {:else}
          Create Workspace
        {/if}
      </button>
    </div>
  </div>
</div>
