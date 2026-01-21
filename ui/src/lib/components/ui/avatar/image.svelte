<script lang="ts">
	import { Avatar as AvatarPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils';

	type Props = {
		class?: string;
		src?: string;
		alt?: string;
	};

	let { class: className = '', src, alt = '' }: Props = $props();
</script>

<AvatarPrimitive.Image
	{src}
	{alt}
	class={cn('aspect-square h-full w-full', className)}
/>
