<script lang="ts">
	import { setContext } from 'svelte';
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { X } from 'lucide-svelte';
	import { cn } from '$lib/utils';

	interface Props {
		open?: boolean;
		onOpenChange?: (open: boolean) => void;
		children: any;
	}

	let { open = false, onOpenChange, children }: Props = $props();

	const ROOT = DialogPrimitive.Root;
	const TRIGGER = DialogPrimitive.Trigger;
	const PORTAL = DialogPrimitive.Portal;
	const CLOSE = DialogPrimitive.Close;
	const TITLE = DialogPrimitive.Title;
	const DESCRIPTION = DialogPrimitive.Description;

	export { ROOT, TRIGGER, PORTAL, CLOSE, TITLE, DESCRIPTION };
</script>
