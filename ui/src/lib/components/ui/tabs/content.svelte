<script lang="ts">
	import { cn } from '$lib/utils';
	import { Tabs as TabsPrimitive } from 'bits-ui';

	interface Props {
		class?: string;
		value?: string;
		children?: any;
	}

	let { class: className = '', value, children }: Props = $props();
</script>

<TabsPrimitive.Content
	{value}
	class={cn(
		'mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
		className
	)}
>
	{children}
</TabsPrimitive.Content>
