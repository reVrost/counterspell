<script lang="ts">
	import { cn } from '$lib/utils';
	import type { HTMLInputAttributes } from 'svelte/elements';

	interface Props extends HTMLInputAttributes {
		class?: string;
		value?: string;
	}

	let { class: className, value = $bindable(''), ...restProps }: Props = $props();
</script>

<input
	bind:value
	class={cn(
		'flex h-10 w-full rounded-lg border border-input bg-input px-3 py-2 text-sm text-foreground ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:border-primary focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 transition-colors',
		className
	)}
	{...restProps}
/>
