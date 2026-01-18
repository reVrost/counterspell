<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Message, MessageContent } from '$lib/types';
	import UserIcon from '@lucide/svelte/icons/user';
	import BotIcon from '@lucide/svelte/icons/bot';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import TerminalIcon from '@lucide/svelte/icons/terminal';
	import PenIcon from '@lucide/svelte/icons/pen';
	import EyeIcon from '@lucide/svelte/icons/eye';
	import SearchIcon from '@lucide/svelte/icons/search';
	import FolderOpenIcon from '@lucide/svelte/icons/folder-open';
	import SettingsIcon from '@lucide/svelte/icons/settings';

	interface Props {
		msg: Message;
	}

	let { msg }: Props = $props();

	function hasTextContent(message: Message): boolean {
		return message.content.some((c) => c.type === 'text' && c.text);
	}

	function toolIcon(name: string): typeof SettingsIcon {
		if (name?.includes('edit') || name?.includes('write')) return PenIcon;
		if (name?.includes('read') || name?.includes('view')) return EyeIcon;
		if (name?.includes('bash') || name?.includes('exec')) return TerminalIcon;
		if (name?.includes('search') || name?.includes('grep') || name?.includes('glob'))
			return SearchIcon;
		if (name?.includes('list') || name?.includes('ls')) return FolderOpenIcon;
		return SettingsIcon;
	}

	function truncateResult(s: string): string {
		if (s.length > 2000) {
			return s.slice(0, 2000) + '\n... (truncated)';
		}
		return s;
	}
</script>

{#if msg.role === 'user' && hasTextContent(msg)}
	<!-- User message -->
	<div class="px-4 py-2 border-b border-gray-800/50">
		<div class="flex gap-2">
			<div
				class="w-4 h-4 rounded-full bg-blue-500/20 border border-blue-500/30 flex items-center justify-center shrink-0"
			>
				<UserIcon class="w-2 h-2 text-blue-400" />
			</div>
			<div class="flex-1 min-w-0 -mt-px">
				{#each msg.content as content}
					{#if content.type === 'text' && content.text}
						<div class="text-sm text-gray-200 whitespace-pre-wrap leading-normal">{content.text}</div>
					{/if}
				{/each}
			</div>
		</div>
	</div>
{:else if msg.role === 'assistant'}
	<!-- Assistant message -->
	<div class="px-4 py-2 bg-[#0D1117] border-b border-gray-800/50">
		<div class="flex gap-2">
			<div
				class="w-4 h-4 rounded-full bg-purple-500/20 border border-purple-500/30 flex items-center justify-center shrink-0"
			>
				<BotIcon class="w-2 h-2 text-purple-400" />
			</div>
			<div class="flex-1 min-w-0 space-y-1 -mt-px">
				{#each msg.content as content}
					{#if content.type === 'text' && content.text}
						<div
							class="text-sm text-gray-300 leading-normal prose prose-invert prose-sm prose-p:my-3 prose-headings:font-bold prose-headings:text-sm prose-headings:mt-4 prose-headings:mb-2 prose-code:text-xs prose-pre:text-xs prose-pre:my-3 prose-ul:my-2 prose-ol:my-2 prose-li:my-1 max-w-none"
						>
							{@html content.text}
						</div>
					{:else if content.type === 'tool_use'}
						<!-- Tool Call Block -->
						{@const Icon = toolIcon(content.toolName || '')}
						<div class="my-1">
							<button
								class="flex items-center gap-1.5 text-xs text-gray-500 hover:text-gray-300 transition group"
							>
								<div
									class="w-4 h-4 rounded bg-gray-800 border border-gray-700 flex items-center justify-center group-hover:border-gray-600 transition"
								>
									<svelte:component this={Icon} class="w-2.5 h-2.5 text-gray-500 group-hover:text-gray-400" />
								</div>
								<span class="font-mono">{content.toolName}</span>
								<ChevronRightIcon class="w-3 h-3" />
							</button>
						</div>
					{:else if content.type === 'tool_result' && content.text}
						<!-- Tool Result Block -->
						<div class="my-1 ml-5">
							<button
								class="flex items-center gap-1.5 text-[10px] text-gray-600 hover:text-gray-400 transition"
							>
								<CheckCircleIcon class="w-2 h-2 text-green-600" />
								<span class="font-mono">result</span>
								<ChevronRightIcon class="w-2 h-2" />
							</button>
						</div>
					{/if}
				{/each}
			</div>
		</div>
	</div>
{/if}
