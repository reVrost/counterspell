<script lang="ts">
	import { Tabs as TabsPrimitive } from "bits-ui";
	import { cn } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: TabsPrimitive.ListProps = $props();
</script>

<TabsPrimitive.List
	bind:ref
	data-slot="tabs-list"
	class={cn(
		"bg-muted/30 border-border/50 inline-flex h-10 items-center justify-center rounded-lg border p-1 text-muted-foreground backdrop-blur-sm",
		className,
	)}
	{...restProps}
/>
