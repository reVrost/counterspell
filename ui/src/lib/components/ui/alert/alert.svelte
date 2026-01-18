<script lang="ts">
	import { cn } from '$lib/utils';
	import { tv, type VariantProps } from 'tailwind-variants';
	import type { Snippet } from 'svelte';
	import { AlertCircle, CheckCircle, Info, XCircle, AlertTriangle } from 'lucide-svelte';

	const alertVariants = tv({
		base: 'relative w-full rounded-lg border p-4 [&>svg~*]:pl-7 [&>svg+div]:translate-y-[-3px] [&>svg]:absolute [&>svg]:left-4 [&>svg]:top-4 [&>svg]:text-foreground',
		variants: {
			variant: {
				default: 'bg-background text-foreground',
				destructive: 'border-destructive/50 text-destructive dark:border-destructive [&>svg]:text-destructive'
			}
		},
		defaultVariants: {
			variant: 'default'
		}
	});

	type AlertVariants = VariantProps<typeof alertVariants>;

	interface Props {
		class?: string;
		variant?: AlertVariants['variant'];
		title?: string;
		children?: Snippet;
		hideIcon?: boolean;
	}

	let { class: className = '', variant = 'default', title, children, hideIcon = false }: Props = $props();

	const icons = {
		default: Info,
		destructive: AlertCircle,
		success: CheckCircle,
		warning: AlertTriangle,
		error: XCircle
	};

	const Icon = icons[variant];
</script>

<div class={cn(alertVariants({ variant }), className)}>
	{#if !hideIcon}
		<svelte:component this={Icon} class="h-4 w-4" />
	{/if}

	{#if title}
		<h5 class="mb-1 font-medium leading-none tracking-tight">{title}</h5>
	{/if}

	{#if children}
		<div class="text-sm [&_p]:leading-relaxed">
			{@render children()}
		</div>
	{/if}
</div>
