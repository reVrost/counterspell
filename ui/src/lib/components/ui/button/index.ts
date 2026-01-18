import Root from './button.svelte';
import { tv, type VariantProps } from 'tailwind-variants';

export const buttonVariants = tv({
	base: 'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-lg text-sm font-medium transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 active:scale-[0.98]',
	variants: {
		variant: {
			default: 'bg-primary text-primary-foreground hover:bg-primary/90 shadow-lg shadow-primary/20',
			destructive:
				'bg-destructive text-destructive-foreground hover:bg-destructive/90 shadow-lg shadow-destructive/20',
			outline: 'border border-input bg-transparent hover:bg-accent hover:text-accent-foreground',
			secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
			ghost: 'hover:bg-accent hover:text-accent-foreground',
			link: 'text-primary underline-offset-4 hover:underline',
			// Custom variants matching your theme
			white: 'bg-white text-black hover:bg-gray-100 shadow-lg shadow-white/10',
			card: 'bg-card border border-white/[0.08] hover:border-purple-500/30 hover:bg-card/80'
		},
		size: {
			default: 'h-10 px-4 py-2',
			sm: 'h-9 rounded-md px-3',
			lg: 'h-12 rounded-xl px-8',
			xl: 'h-14 rounded-2xl px-10 text-base',
			icon: 'h-10 w-10',
			'icon-sm': 'h-8 w-8',
			'icon-lg': 'h-11 w-11 rounded-xl'
		}
	},
	defaultVariants: {
		variant: 'default',
		size: 'default'
	}
});

export type ButtonVariants = VariantProps<typeof buttonVariants>;
export { Root as Button };
