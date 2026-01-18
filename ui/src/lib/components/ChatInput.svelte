<script lang="ts">
	import { appState } from '$lib/stores/app.svelte';
	import { MODELS, type Project } from '$lib/types';
	import { cn } from '$lib/utils';
	import { tasksAPI, filesAPI } from '$lib/api';
	import FolderIcon from '@lucide/svelte/icons/folder';
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import PaperclipIcon from '@lucide/svelte/icons/paperclip';
	import ZapIcon from '@lucide/svelte/icons/zap';
	import MicIcon from '@lucide/svelte/icons/mic';
	import ArrowUpIcon from '@lucide/svelte/icons/arrow-up';
	import XIcon from '@lucide/svelte/icons/x';
	import SearchIcon from '@lucide/svelte/icons/search';
	import FileIcon from '@lucide/svelte/icons/file';
	import CheckIcon from '@lucide/svelte/icons/check';
	import LoaderIcon from '@lucide/svelte/icons/loader-2';

	interface Props {
		mode: 'create' | 'chat';
		taskId?: string;
		placeholder?: string;
		onSubmit?: (message: string, modelId: string) => void;
		onClose?: () => void;
	}

	let { mode, taskId, placeholder = 'What do you want to build?', onSubmit, onClose }: Props = $props();

	let text = $state('');
	let inputRef = $state<HTMLTextAreaElement | null>(null);
	let modelOpen = $state(false);
	let showFileMenu = $state(false);
	let files = $state<string[]>([]);
	let selectedIndex = $state(0);
	let projectSearch = $state('');

	function resize() {
		if (!inputRef) return;
		inputRef.style.height = 'auto';
		let newHeight = inputRef.scrollHeight;
		const maxHeight = window.innerHeight * (mode === 'create' ? 0.4 : 0.35);
		if (newHeight > maxHeight) newHeight = maxHeight;
		inputRef.style.height = `${newHeight}px`;
	}

	async function checkMention(e: KeyboardEvent) {
		const match = text.match(/@([^ ]*)$/);
		if (match) {
			showFileMenu = true;
			const query = match[1] || '';
			await searchFiles(query);
		} else {
			showFileMenu = false;
			files = [];
		}
		if (e.key === 'Escape') {
			showFileMenu = false;
			files = [];
		}
	}

	async function searchFiles(query: string) {
		const projectId = appState.activeProjectId;
		if (!projectId) return;
		try {
			files = await filesAPI.search(projectId, query);
			selectedIndex = 0;
		} catch (e) {
			console.error('File search failed:', e);
			files = [];
		}
	}

	function handleFileNav(e: KeyboardEvent) {
		if (!showFileMenu || !files || files.length === 0) return;
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			selectedIndex = Math.min(selectedIndex + 1, files.length - 1);
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			selectedIndex = Math.max(selectedIndex - 1, 0);
		} else if (e.key === 'Enter' || e.key === 'Tab') {
			e.preventDefault();
			if (files[selectedIndex]) {
				insertFile(files[selectedIndex]);
			}
		}
	}

	function insertFile(f: string) {
		text = text.replace(/@[^ ]*$/, '') + f + ' ';
		showFileMenu = false;
		files = [];
		selectedIndex = 0;
		inputRef?.focus();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (showFileMenu && files && files.length > 0 && ['ArrowUp', 'ArrowDown', 'Tab'].includes(e.key)) {
			handleFileNav(e);
		} else if (showFileMenu && files && files.length > 0 && e.key === 'Enter' && !e.shiftKey) {
			handleFileNav(e);
		} else if (!showFileMenu && e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			submit();
		}
	}

	async function submit() {
		if (!text.trim()) return;
		if (mode === 'create' && !appState.activeProjectId) {
			appState.showToast('Select a project first');
			return;
		}

		const msg = text.trim();
		text = '';
		if (inputRef) inputRef.style.height = 'auto';

		if (onSubmit) {
			onSubmit(msg, appState.activeModelId);
		} else if (mode === 'create') {
			// Handle task creation
			try {
				await tasksAPI.create(msg, appState.activeProjectId, appState.activeModelId);
			} catch (e) {
				console.error('Failed to create task:', e);
				appState.showToast('Failed to create task');
			}
		} else if (mode === 'chat' && taskId) {
			// Handle chat submission
			try {
				await tasksAPI.chat(taskId, msg, appState.activeModelId);
			} catch (e) {
				console.error('Failed to send message:', e);
				appState.showToast('Failed to send message');
			}
		}

		navigator.vibrate?.(30);
	}

	function handleSubmitClick() {
		if (text.length > 0) {
			navigator.vibrate?.(30);
			submit();
		}
	}

	// Props for project selector (create mode only)
	interface ProjectSelectorProps {
		projects?: Record<string, Project>;
	}
</script>

<div class="relative">
	<div
		class="bg-secondary/60 border border-white/10 rounded-[24px] shadow-2xl relative backdrop-blur-xl transition-all duration-200 ring-1 ring-white/5 flex flex-col group focus-within:border-white/20 focus-within:ring-white/10"
	>
		<!-- Voice Recording Visualization -->
		{#if appState.isRecording}
			<div class="absolute inset-0 bg-secondary rounded-3xl flex items-center px-4 z-10">
				<button
					type="button"
					class="w-8 h-8 rounded-full bg-gray-800 hover:bg-gray-700 flex items-center justify-center text-gray-400 hover:text-white transition shrink-0"
				>
					<XIcon class="w-4 h-4" />
				</button>
				<div class="flex-1 flex items-center justify-center gap-[3px] h-10 mx-4">
					{#each appState.audioLevels as level, i}
						<div
							class="w-1 bg-red-500 rounded-full transition-all duration-75"
							style="height: {Math.max(4, level * 0.4)}px; opacity: {0.4 + (level / 100) * 0.6}"
						></div>
					{/each}
				</div>
				<div class="flex items-center gap-2 shrink-0">
					<div class="w-2 h-2 rounded-full bg-red-500 animate-pulse"></div>
					<span class="text-sm text-gray-300 font-mono"
						>{Math.floor(appState.recordedDuration / 60)}:{String(appState.recordedDuration % 60).padStart(2, '0')}</span
					>
				</div>
			</div>
		{/if}

		<!-- Transcribing Overlay -->
		{#if appState.isTranscribing}
			<div
				class="absolute inset-0 bg-secondary rounded-3xl flex items-center justify-center z-10"
			>
				<LoaderIcon class="w-4 h-4 text-blue-400 mr-2 animate-spin" />
				<span class="text-gray-400 text-sm">Transcribing...</span>
			</div>
		{/if}

		<!-- Input Area -->
		<div class="relative px-4 pt-4 pb-2">
			<!-- File Menu Popover -->
			{#if showFileMenu && files && files.length > 0}
				<div
					class="absolute bottom-full left-0 mb-2 w-80 bg-popover border border-gray-700 rounded-xl shadow-2xl overflow-hidden max-h-48 overflow-y-auto z-50"
				>
					<div
						class="px-3 py-2 text-[10px] text-gray-500 font-bold uppercase tracking-wider border-b border-gray-800 flex items-center justify-between"
					>
						<span>Files</span>
						<span class="text-gray-600 font-normal normal-case"
							>↑↓ to navigate, Enter to select</span
						>
					</div>
					{#each files as file, idx}
						<button
							type="button"
							onclick={() => insertFile(file)}
							onmouseenter={() => (selectedIndex = idx)}
							class={cn(
								'w-full px-3 py-2 text-xs font-mono cursor-pointer transition flex items-center gap-2 text-left',
								idx === selectedIndex
									? 'bg-blue-500/20 text-blue-300'
									: 'text-gray-300 hover:bg-white/5'
							)}
						>
							<FileIcon class="w-2.5 h-2.5 opacity-40" />
							<span class="truncate">{file}</span>
						</button>
					{/each}
				</div>
			{/if}

			<!-- Empty State -->
			{#if showFileMenu && files && files.length === 0}
				<div
					class="absolute bottom-full left-0 mb-2 w-64 bg-popover border border-gray-700 rounded-xl shadow-2xl overflow-hidden z-50"
				>
					<div class="px-3 py-4 text-xs text-gray-500 text-center">
						<SearchIcon class="w-4 h-4 mx-auto mb-2 opacity-50" />
						<div>No files found</div>
					</div>
				</div>
			{/if}

			<textarea
				bind:this={inputRef}
				bind:value={text}
				oninput={resize}
				onkeyup={checkMention}
				onkeydown={handleKeydown}
				rows="1"
				{placeholder}
				aria-label={placeholder}
				class="bg-transparent border-none focus:ring-0 focus:outline-none text-white text-[15px] placeholder-gray-500 w-full resize-none font-medium p-0 leading-relaxed max-h-[40vh] min-h-[24px]"
			></textarea>
		</div>

		<!-- Toolbar -->
		<div class="flex items-center justify-between px-2 pb-2 mt-1">
			<!-- Left Side -->
			<div class="flex items-center gap-1">
				{#if mode === 'create'}
					<!-- Project Selector -->
					<div class="relative">
						<button
							type="button"
							onclick={() => (appState.inputProjectMenuOpen = !appState.inputProjectMenuOpen)}
							class={cn(
								'flex items-center gap-2 px-2.5 py-1.5 rounded-full text-xs font-semibold transition-all duration-200 border border-transparent',
								appState.activeProjectId
									? 'bg-blue-500/20 text-blue-400 hover:bg-blue-500/30'
									: 'bg-gray-800/50 text-gray-400 hover:bg-gray-800'
							)}
						>
							<div
								class={cn(
									'flex items-center justify-center w-3.5 h-3.5 rounded-full',
									appState.activeProjectId ? 'bg-blue-400/20' : 'bg-gray-700/50'
								)}
							>
								<FolderIcon
									class={cn(
										'w-2.5 h-2.5',
										appState.activeProjectId ? 'text-blue-400' : 'text-gray-400'
									)}
								/>
							</div>
							<span class="max-w-[120px] truncate">
								{appState.activeProjectName
									? appState.activeProjectName.split('/').pop()
									: 'Select project'}
							</span>
							<ChevronDownIcon class="w-2.5 h-2.5 opacity-50 ml-0.5" />
						</button>

						{#if appState.inputProjectMenuOpen}
							<div
								class="absolute bottom-full left-0 mb-3 w-72 bg-popover border border-gray-700 rounded-2xl shadow-[0_0_50px_rgba(0,0,0,0.5)] overflow-hidden flex flex-col z-50 animate-in fade-in slide-in-from-bottom-2 duration-200"
							>
								<!-- Search Header -->
								<div class="p-3 border-b border-gray-800 bg-gray-900/50">
									<div class="relative">
										<SearchIcon class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500 w-3 h-3" />
										<input
											bind:value={projectSearch}
											type="text"
											placeholder="Search repositories..."
											class="w-full bg-black/40 border border-gray-700 rounded-xl pl-8 pr-3 py-2 text-xs text-white focus:outline-none focus:border-blue-500 placeholder-gray-600 transition-all font-medium"
										/>
									</div>
								</div>

								<!-- Projects/Repos List -->
								<div class="max-h-64 overflow-y-auto py-2 px-1 scrollbar-thin">
									{#if appState.projects.length > 0}
										<div class="px-3 py-1 text-[10px] font-bold text-gray-500 uppercase tracking-widest mb-1">
											Active Projects
										</div>
										{#each appState.projects.filter(p => p.name.toLowerCase().includes(projectSearch.toLowerCase())) as p}
											<button
												type="button"
												onclick={() => appState.setActiveProject(p.id, p.name)}
												class={cn(
													'w-full px-3 py-2 rounded-xl flex items-center justify-between group transition-all duration-150',
													appState.activeProjectId === p.id 
														? 'bg-blue-500/10 text-white' 
														: 'text-gray-400 hover:bg-white/5 hover:text-white'
												)}
											>
												<div class="flex items-center gap-3 overflow-hidden">
													<div class="w-2 h-2 rounded-full bg-blue-500/40 group-hover:bg-blue-500 transition-colors"></div>
													<span class="text-xs font-medium truncate">{p.name}</span>
												</div>
												{#if appState.activeProjectId === p.id}
													<CheckIcon class="w-3 h-3 text-blue-500" />
												{/if}
											</button>
										{/each}
										<div class="h-px bg-gray-800 my-2 mx-2"></div>
									{/if}

									<div class="px-3 py-1 text-[10px] font-bold text-gray-500 uppercase tracking-widest mb-1">
										All Repositories
									</div>
									{#each appState.repos.filter(r => r.full_name.toLowerCase().includes(projectSearch.toLowerCase()) && !appState.projects.some(p => p.name === r.full_name)) as r}
										<button
											type="button"
											onclick={() => appState.setActiveProject(r.id.toString(), r.full_name)}
											class="w-full px-3 py-2 rounded-xl flex items-center gap-3 text-gray-400 hover:bg-white/5 hover:text-white group transition-all duration-150"
										>
											<div class="w-2 h-2 rounded-full bg-gray-700 group-hover:bg-gray-500 transition-colors"></div>
											<span class="text-xs font-medium truncate">{r.full_name}</span>
										</button>
									{:else}
										<div class="px-4 py-8 text-center text-gray-600 text-[11px] italic">
											No repositories found matching "{projectSearch}"
										</div>
									{/each}
								</div>

								<!-- Footer -->
								<div class="px-4 py-2 border-t border-gray-800 bg-gray-900/30">
									<p class="text-[9px] text-gray-600 text-center font-medium">
										{appState.repos.length} REPOSITORIES SYNCED
									</p>
								</div>
							</div>
						{/if}
					</div>
				{:else}
					<!-- Chat mode buttons -->
					<button
						type="button"
						onclick={onClose}
						aria-label="Close chat"
						class="w-11 h-11 flex items-center justify-center rounded-xl text-gray-400 active:text-white active:bg-white/10 transition-all duration-150"
					>
						<XIcon class="w-4 h-4" />
					</button>
					<button
						type="button"
						aria-label="Attach file"
						class="w-11 h-11 flex items-center justify-center rounded-xl text-gray-400 active:text-white active:bg-white/10 transition-all duration-150"
					>
						<PaperclipIcon class="w-4 h-4" />
					</button>
					<!-- Model Selector (chat mode inline) -->
					<div class="relative">
						<button
							type="button"
							onclick={() => (modelOpen = !modelOpen)}
							aria-label="Select model"
							class="w-11 h-11 flex items-center justify-center rounded-xl text-gray-400 active:text-white active:bg-white/10 transition-all duration-150"
						>
							<ZapIcon class="w-4 h-4" />
						</button>
						{#if modelOpen}
							<div
								class="absolute bottom-full left-0 mb-3 w-52 bg-popover/95 border border-white/10 rounded-xl shadow-2xl overflow-hidden py-1 z-50 backdrop-blur-xl"
							>
								<div
									class="px-3 py-2 text-[10px] text-gray-500 font-bold uppercase tracking-wider border-b border-white/5 mb-1"
								>
									Select Model
								</div>
								<div class="p-1.5 space-y-0.5">
									{#each MODELS as m}
										<button
											type="button"
											onclick={() => {
												appState.setModel(m.id);
												modelOpen = false;
											}}
											class={cn(
												'w-full flex items-center justify-between px-3 py-2 rounded-lg cursor-pointer transition text-left group',
												appState.activeModelId === m.id
													? 'bg-blue-500/10 text-blue-400 border border-blue-500/20'
													: 'hover:bg-white/5 text-gray-400 hover:text-white border border-transparent'
											)}
										>
											<span class="text-xs font-medium">{m.name}</span>
											{#if appState.activeModelId === m.id}
												<div
													class="w-1.5 h-1.5 rounded-full bg-blue-500 shadow-[0_0_8px_rgba(59,130,246,0.8)]"
												></div>
											{/if}
										</button>
									{/each}
								</div>
							</div>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Right Side -->
			<div class="flex items-center gap-1">
				{#if mode === 'create'}
					<button
						type="button"
						aria-label="Attach file"
						class="w-11 h-11 flex items-center justify-center rounded-xl text-gray-400 active:text-white active:bg-white/10 transition-all duration-150"
					>
						<PaperclipIcon class="w-4 h-4" />
					</button>
					<!-- Model Selector (create mode) -->
					<div class="relative">
						<button
							type="button"
							onclick={() => (modelOpen = !modelOpen)}
							aria-label="Select model"
							class="w-11 h-11 flex items-center justify-center rounded-xl text-gray-400 active:text-white active:bg-white/10 transition-all duration-150"
						>
							<ZapIcon class="w-4 h-4" />
						</button>
						{#if modelOpen}
							<div
								class="absolute bottom-full right-0 mb-3 w-52 bg-popover/95 border border-white/10 rounded-xl shadow-2xl overflow-hidden py-1 z-50 backdrop-blur-xl"
							>
								<div
									class="px-3 py-2 text-[10px] text-gray-500 font-bold uppercase tracking-wider border-b border-white/5 mb-1"
								>
									Select Model
								</div>
								<div class="p-1.5 space-y-0.5">
									{#each MODELS as m}
										<button
											type="button"
											onclick={() => {
												appState.setModel(m.id);
												modelOpen = false;
											}}
											class={cn(
												'w-full flex items-center justify-between px-3 py-2 rounded-lg cursor-pointer transition text-left group',
												appState.activeModelId === m.id
													? 'bg-blue-500/10 text-blue-400 border border-blue-500/20'
													: 'hover:bg-white/5 text-gray-400 hover:text-white border border-transparent'
											)}
										>
											<span class="text-xs font-medium">{m.name}</span>
											{#if appState.activeModelId === m.id}
												<div
													class="w-1.5 h-1.5 rounded-full bg-blue-500 shadow-[0_0_8px_rgba(59,130,246,0.8)]"
												></div>
											{/if}
										</button>
									{/each}
								</div>
							</div>
						{/if}
					</div>
				{/if}

				<!-- Divider -->
				<div class="w-px h-4 bg-gray-700 mx-1"></div>

				<!-- Submit/Voice Button -->
				<button
					type="button"
					aria-label="Send message or hold to record voice"
					onclick={handleSubmitClick}
					class={cn(
						'w-9 h-9 rounded-xl flex items-center justify-center text-base transition-all duration-150 select-none',
						appState.isRecording
							? 'bg-red-500 shadow-[0_0_20px_rgba(239,68,68,0.5)] scale-110'
							: text.length > 0
								? 'bg-blue-500 text-white shadow-[0_0_15px_rgba(59,130,246,0.4)]'
								: 'bg-white/10 text-gray-300 hover:bg-white/15'
					)}
				>
					{#if text.length > 0}
						<ArrowUpIcon class="w-4 h-4" />
					{:else if appState.isRecording}
						<MicIcon class="w-4 h-4 animate-pulse" />
					{:else}
						<MicIcon class="w-4 h-4" />
					{/if}
				</button>
			</div>
		</div>
	</div>
</div>
