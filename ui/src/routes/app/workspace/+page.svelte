<script lang="ts">
  import { appState } from '$lib/stores/app.svelte';
  import { workspacesAPI } from '$lib/api';
  import type { WorkspaceFile, WorkspaceFileFilter } from '$lib/types';
  import UploadIcon from '@lucide/svelte/icons/upload';
  import FileIcon from '@lucide/svelte/icons/file';
  import FileTextIcon from '@lucide/svelte/icons/file-text';
  import ImageIcon from '@lucide/svelte/icons/image';
  import FolderIcon from '@lucide/svelte/icons/folder';
  import CopyIcon from '@lucide/svelte/icons/copy';
  import MoreVerticalIcon from '@lucide/svelte/icons/more-vertical';
  import { formatBytes, formatDate } from '$lib/utils';

  let files = $state<WorkspaceFile[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let activeFilter = $state<WorkspaceFileFilter>('all');
  let uploading = $state(false);
  let selectedFile = $state<WorkspaceFile | null>(null);
  let showActionSheet = $state(false);

  const filters: { id: WorkspaceFileFilter; label: string }[] = [
    { id: 'all', label: 'All' },
    { id: 'recent', label: 'Recent' },
    { id: 'images', label: 'Images' },
    { id: 'docs', label: 'Docs' },
  ];

  const currentWorkspace = $derived(
    appState.workspaces.find((w) => w.id === appState.activeWorkspaceId)
  );

  async function loadFiles() {
    if (!appState.activeWorkspaceId) return;

    try {
      loading = true;
      error = null;
      files = await workspacesAPI.listFiles(
        appState.activeWorkspaceId,
        '',
        activeFilter,
        200
      );
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load files';
      console.error('Failed to load workspace files:', err);
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    loadFiles();
  });

  async function handleUpload(event: Event) {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file || !appState.activeWorkspaceId) return;

    uploading = true;
    try {
      await workspacesAPI.uploadFile(appState.activeWorkspaceId, file);
      await loadFiles();
      appState.showToast('File uploaded successfully');
    } catch (err) {
      appState.showToast('Failed to upload file', 'error');
      console.error('Upload error:', err);
    } finally {
      uploading = false;
      input.value = '';
    }
  }

  function getFileIcon(file: WorkspaceFile) {
    if (file.is_dir) return FolderIcon;
    if (
      ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.svg', '.ico'].includes(
        file.extension || ''
      )
    ) {
      return ImageIcon;
    }
    if (['.txt', '.md', '.pdf', '.doc', '.docx'].includes(file.extension || '')) {
      return FileTextIcon;
    }
    return FileIcon;
  }

  function handleFileClick(file: WorkspaceFile) {
    selectedFile = file;
    showActionSheet = true;
  }

  function copyPath(path: string) {
    navigator.clipboard.writeText(path);
    appState.showToast('Path copied to clipboard');
    showActionSheet = false;
  }

  function copyFileReference(path: string) {
    navigator.clipboard.writeText(`@${path}`);
    appState.showToast('File reference copied');
    showActionSheet = false;
  }
</script>

<svelte:head>
  <title>Workspace | Counterspell</title>
</svelte:head>

<div class="flex flex-col h-dvh bg-[#0a0a0a]">
  <header class="flex-shrink-0 px-4 pt-12 pb-4 border-b border-white/[0.06]">
    <div class="flex items-center justify-between mb-4">
      <div>
        <h1 class="text-xl font-semibold text-white">Workspace</h1>
        {#if currentWorkspace}
          <p class="text-sm text-zinc-500 mt-0.5">{currentWorkspace.name}</p>
        {:else}
          <p class="text-sm text-zinc-500 mt-0.5">Select a workspace</p>
        {/if}
      </div>
      <label
        class="flex items-center gap-2 px-4 py-2.5 bg-violet-600 hover:bg-violet-500 text-white text-sm font-medium rounded-full transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <UploadIcon class="w-4 h-4" />
        <span>Upload</span>
        <input
          type="file"
          class="hidden"
          onchange={handleUpload}
          disabled={uploading || !appState.activeWorkspaceId}
        />
      </label>
    </div>

    <div class="flex gap-2 bg-[#1a1a1a] p-1 rounded-full border border-white/[0.06]">
      {#each filters as filter}
        <button
          type="button"
          onclick={() => (activeFilter = filter.id)}
          class="flex-1 px-4 py-2 text-sm font-medium rounded-full transition-all {activeFilter === filter.id
            ? 'bg-white text-black'
            : 'text-zinc-400 hover:text-zinc-300'}"
        >
          {filter.label}
        </button>
      {/each}
    </div>
  </header>

  <main class="flex-1 overflow-y-auto px-4 py-4">
    {#if !appState.activeWorkspaceId}
      <div class="flex flex-col items-center justify-center h-full text-center py-12">
        <div
          class="w-16 h-16 rounded-2xl bg-zinc-900/50 border border-white/[0.06] flex items-center justify-center mb-4"
        >
          <FolderIcon class="w-8 h-8 text-zinc-600" />
        </div>
        <h3 class="text-base font-medium text-zinc-300 mb-1">No Workspace Selected</h3>
        <p class="text-sm text-zinc-500 max-w-[240px]">
          Select a workspace to browse and upload files
        </p>
      </div>
    {:else if loading}
      <div class="space-y-3">
        {#each Array(6) as _}
          <div
            class="h-16 bg-[#1a1a1a] rounded-xl border border-white/[0.04] animate-pulse"
          ></div>
        {/each}
      </div>
    {:else if error}
      <div
        class="flex flex-col items-center justify-center h-full text-center py-12"
      >
        <div
          class="w-16 h-16 rounded-2xl bg-red-500/10 border border-red-500/20 flex items-center justify-center mb-4"
        >
          <FileIcon class="w-8 h-8 text-red-400" />
        </div>
        <h3 class="text-base font-medium text-zinc-300 mb-1">Failed to Load Files</h3>
        <p class="text-sm text-zinc-500 max-w-[240px]">{error}</p>
      </div>
    {:else if files.length === 0}
      <div class="flex flex-col items-center justify-center h-full text-center py-12">
        <div
          class="w-16 h-16 rounded-2xl bg-zinc-900/50 border border-white/[0.06] flex items-center justify-center mb-4"
        >
          <FileIcon class="w-8 h-8 text-zinc-600" />
        </div>
        <h3 class="text-base font-medium text-zinc-300 mb-1">No Files Yet</h3>
        <p class="text-sm text-zinc-500 max-w-[240px]">
          Upload files to use in tasks
        </p>
      </div>
    {:else}
      <div class="space-y-2 pb-20">
        {#each files as file}
          {@const Icon = getFileIcon(file)}
          <button
            type="button"
            onclick={() => handleFileClick(file)}
            class="w-full flex items-center gap-3 px-4 py-3 bg-[#1a1a1a] hover:bg-[#252525] rounded-xl border border-white/[0.04] hover:border-white/[0.08] transition-all group"
          >
            <div
              class="flex-shrink-0 w-10 h-10 rounded-lg bg-zinc-900/50 border border-white/[0.06] flex items-center justify-center"
            >
              <Icon class="w-5 h-5 text-zinc-400 group-hover:text-zinc-300 transition-colors" />
            </div>

            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <p class="text-sm font-medium text-zinc-200 truncate">
                  {file.name}
                </p>
                {#if file.is_dir}
                  <span
                    class="flex-shrink-0 px-1.5 py-0.5 bg-zinc-800 text-[10px] font-medium text-zinc-500 rounded-full"
                    >DIR</span
                  >
                {/if}
              </div>
              <p class="text-xs text-zinc-500 truncate mt-0.5">
                {file.path}
              </p>
            </div>

            <div class="flex-shrink-0 text-right">
              <p class="text-xs text-zinc-500">
                {!file.is_dir ? formatBytes(file.size) : ''}
              </p>
              <p class="text-[10px] text-zinc-600 mt-0.5">
                {formatDate(file.mod_time)}
              </p>
            </div>

            <div
              class="flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center text-zinc-600 hover:text-zinc-400 hover:bg-white/[0.04] transition-all"
            >
              <MoreVerticalIcon class="w-4 h-4" />
            </div>
          </button>
        {/each}
      </div>
    {/if}
  </main>

  {#if showActionSheet && selectedFile}
    <div
      class="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-end sm:items-center justify-center p-4"
      onclick={() => (showActionSheet = false)}
      role="presentation"
    >
      <div
        class="bg-[#1a1a1a] w-full max-w-sm rounded-2xl border border-white/[0.06] shadow-2xl overflow-hidden"
        onclick={(e) => e.stopPropagation()}
        onkeydown={(e) => e.key === 'Escape' && (showActionSheet = false)}
        role="dialog"
        aria-modal="true"
        tabindex="0"
      >
        <div class="px-4 py-4 border-b border-white/[0.04]">
          <p class="text-sm font-medium text-zinc-300 truncate">
            {selectedFile.name}
          </p>
          <p class="text-xs text-zinc-500 truncate mt-0.5">
            {selectedFile.path}
          </p>
        </div>

        <div class="py-2">
          <button
            type="button"
            onclick={() => copyPath(selectedFile.path)}
            class="w-full flex items-center gap-3 px-4 py-3 hover:bg-white/[0.04] transition-colors"
          >
            <div
              class="flex-shrink-0 w-8 h-8 rounded-lg bg-zinc-900/50 border border-white/[0.06] flex items-center justify-center"
            >
              <CopyIcon class="w-4 h-4 text-zinc-400" />
            </div>
            <span class="text-sm text-zinc-300">Copy Path</span>
          </button>

          <button
            type="button"
            onclick={() => copyFileReference(selectedFile.path)}
            class="w-full flex items-center gap-3 px-4 py-3 hover:bg-white/[0.04] transition-colors"
          >
            <div
              class="flex-shrink-0 w-8 h-8 rounded-lg bg-zinc-900/50 border border-white/[0.06] flex items-center justify-center"
            >
              <CopyIcon class="w-4 h-4 text-zinc-400" />
            </div>
            <div class="flex-1">
              <span class="text-sm text-zinc-300">Copy @file reference</span>
              <p class="text-xs text-zinc-500 mt-0.5">
                Paste into task input
              </p>
            </div>
          </button>
        </div>

        <div class="px-4 py-2 border-t border-white/[0.04]">
          <button
            type="button"
            onclick={() => (showActionSheet = false)}
            class="w-full px-4 py-3 text-sm font-medium text-zinc-400 hover:text-zinc-300 hover:bg-white/[0.04] rounded-lg transition-colors"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>
