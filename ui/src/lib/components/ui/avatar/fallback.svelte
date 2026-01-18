<script lang="ts">
	import { Avatar as AvatarPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';

	type Props = {
		class?: string;
		children: Snippet;
	};

	let { class: className = '', children }: Props = $props();
</script>

<AvatarPrimitive.Fallback
	class={cn('flex h-full w-full items-center justify-center rounded-full bg-muted', className)}
>
	{@render children()}
</AvatarPrimitive.Fallback>
