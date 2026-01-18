<script lang="ts">
	import { cn } from '$lib/utils';
	import { tv, type VariantProps } from 'tailwind-variants';

	const badgeVariants = tv({
		base: 'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
		variants: {
			variant: {
				default: 'border-transparent bg-primary text-primary-foreground hover:bg-primary/80',
				secondary: 'border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80',
				destructive: 'border-transparent bg-destructive text-destructive-foreground hover:bg-destructive/80',
				outline: 'text-foreground',
				success: 'border-transparent bg-green-500/10 text-green-400 border-green-500/20',
				warning: 'border-transparent bg-yellow-500/10 text-yellow-400 border-yellow-500/20',
				info: 'border-transparent bg-blue-500/10 text-blue-400 border-blue-500/20'
			}
		},
		defaultVariants: {
			variant: 'default'
		}
	});

	export type BadgeVariants = VariantProps<typeof badgeVariants>;

	type Props = {
		class?: string;
		variant?: BadgeVariants['variant'];
		children: any;
	};

	let { class: className = '', variant = 'default', children }: Props = $props();
</script>

<div class={cn(badgeVariants({ variant }), className)}>
	{children}
</div>
