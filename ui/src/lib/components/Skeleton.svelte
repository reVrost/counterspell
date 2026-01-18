<script lang="ts">
	interface Props {
		variant?: 'text' | 'rounded' | 'circular';
		class?: string;
	}

	let { variant = 'text', class: className = '' }: Props = $props();

	const variants = {
		text: 'h-4 bg-gray-800 rounded animate-pulse',
		rounded: 'rounded-xl bg-gray-800 animate-pulse',
		circular: 'rounded-full bg-gray-800 animate-pulse'
	};
</script>

<div class={variants[variant]} {className}></div>
