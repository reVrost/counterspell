<script lang="ts">
	import { cn } from '$lib/utils';
	import type { LogEntry } from '$lib/types';

	interface Props {
		log: LogEntry;
	}

	let { log }: Props = $props();

	const dotColors: Record<string, string> = {
		error: 'bg-red-500',
		success: 'bg-green-500',
		plan: 'bg-yellow-500',
		code: 'bg-purple-500',
		info: 'bg-blue-500'
	};

	const textColors: Record<string, string> = {
		error: 'text-red-400',
		success: 'text-green-400',
		plan: 'text-yellow-400',
		code: 'text-purple-400',
		info: 'text-blue-400'
	};

	function formatTime(timestamp: string): string {
		const date = new Date(timestamp);
		return date.toLocaleTimeString('en-US', {
			hour12: false,
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit'
		});
	}
</script>

<div class="ml-4 relative">
	<div
		class={cn(
			'absolute -left-[21px] top-1 h-2.5 w-2.5 rounded-full border border-[#0D1117]',
			dotColors[log.type] || 'bg-gray-500'
		)}
	></div>
	<div class="flex justify-between items-start">
		<span
			class={cn('text-xs font-bold block mb-0.5', textColors[log.type] || 'text-gray-300')}
		>
			{log.type}
		</span>
		<span class="text-[10px] text-gray-600 font-mono">{formatTime(log.timestamp)}</span>
	</div>
	<p class={cn('text-xs', log.type === 'error' ? 'text-red-300' : 'text-gray-400')}>
		{log.message}
	</p>
</div>
