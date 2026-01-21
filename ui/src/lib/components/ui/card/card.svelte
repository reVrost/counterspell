<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	type Props = {
		class?: string;
		children: Snippet;
	};

	let { class: className = '', children }: Props = $props();
</script>

<div class={cn('rounded-xl border bg-card text-card-foreground shadow', className)}>
	{@render children()}
</div>
