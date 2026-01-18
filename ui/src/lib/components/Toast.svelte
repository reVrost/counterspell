<script lang="ts">
	import { appState } from '$lib/stores/app.svelte';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
</script>

{#if appState.toastOpen}
	<div
		class="fixed top-6 left-1/2 -translate-x-1/2 z-[60] bg-gray-900 border border-gray-700/50 text-white px-4 py-2 rounded-full shadow-2xl flex items-center gap-3 text-sm font-medium transition-all duration-300"
		class:translate-y-0={appState.toastOpen}
		class:-translate-y-full={!appState.toastOpen}
		class:opacity-100={appState.toastOpen}
		class:opacity-0={!appState.toastOpen}
	>
		<CheckCircleIcon class="w-4 h-4 text-green-500" />
		<span>{appState.toastMsg}</span>
	</div>
{/if}
