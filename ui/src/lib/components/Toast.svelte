<script lang="ts">
	import { appState } from '$lib/stores/app.svelte';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import InfoIcon from '@lucide/svelte/icons/info';
	import ExternalLinkIcon from '@lucide/svelte/icons/external-link';

	const iconClasses = {
		success: 'text-green-500',
		error: 'text-red-500',
		info: 'text-blue-500'
	};

	const borderClasses = {
		success: 'border-green-500/30',
		error: 'border-red-500/30',
		info: 'border-blue-500/30'
	};
</script>

{#if appState.toastOpen}
	<div
		class="fixed top-6 left-1/2 -translate-x-1/2 z-[60] bg-gray-900 border text-white px-4 py-2 rounded-full shadow-2xl flex items-center gap-3 text-sm font-medium transition-all duration-300 {borderClasses[appState.toastType]}"
		class:translate-y-0={appState.toastOpen}
		class:-translate-y-full={!appState.toastOpen}
		class:opacity-100={appState.toastOpen}
		class:opacity-0={!appState.toastOpen}
	>
		{#if appState.toastType === 'success'}
			<CheckCircleIcon class="w-4 h-4 {iconClasses.success}" />
		{:else if appState.toastType === 'error'}
			<XCircleIcon class="w-4 h-4 {iconClasses.error}" />
		{:else}
			<InfoIcon class="w-4 h-4 {iconClasses.info}" />
		{/if}
		<span>{appState.toastMsg}</span>
	</div>
{/if}
