<script lang="ts">
	import { cn } from '$lib/utils';
	import { Tabs as TabsPrimitive } from 'bits-ui';

	export const Tabs = TabsPrimitive.Root;
	export const List = TabsPrimitive.List;
	export const Trigger = TabsPrimitive.Trigger;
	export const Content = TabsPrimitive.Content;

	interface Props = {
		class?: string;
		value?: string;
		children?: any;
	}

	let { class: className = '', value, children }: Props = $props();
</script>

<TabsPrimitive.Root {value} class={cn('', className)}>
	{@render children()}
</TabsPrimitive.Root>
