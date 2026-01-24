<script lang="ts">
	import { cn } from '$lib/utils';
	import { Tabs as TabsPrimitive } from 'bits-ui';

	interface Props {
		class?: string;
		children?: any;
	}

	let { class: className = '', children }: Props = $props();
</script>

<TabsPrimitive.List class={cn('inline-flex h-9 items-center justify-center rounded-lg bg-muted p-1 text-muted-foreground', className)}>
	{children}
</TabsPrimitive.List>
