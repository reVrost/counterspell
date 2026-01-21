<script lang="ts">
	import { Avatar as AvatarPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';

	type Props = {
		class?: string;
		children?: Snippet;
	};

	let { class: className = '', children }: Props = $props();
</script>

<AvatarPrimitive.Root
	class={cn('relative flex h-10 w-10 shrink-0 overflow-hidden rounded-full', className)}
>
	{#if children}
		{@render children()}
	{/if}
</AvatarPrimitive.Root>
