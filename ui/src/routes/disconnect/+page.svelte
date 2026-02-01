<script lang="ts">
	import { appState } from '$lib/stores/app.svelte';

	$effect(async () => {
		// Clear auth state
		appState.isAuthenticated = false;
		appState.userEmail = '';

		try {
			// Call backend disconnect endpoint
			await fetch('/disconnect', { method: 'POST', credentials: 'include' });
		} catch (err) {
			console.error('Disconnect failed:', err);
		} finally {
			// Always redirect to home
			window.location.href = '/';
		}
	});
</script>

<div class="h-screen flex items-center justify-center bg-[#0C0E12]">
	<div class="text-center space-y-4">
		<div class="w-12 h-12 border-4 border-purple-500 border-t-transparent rounded-full animate-spin mx-auto"></div>
		<p class="text-gray-400">Disconnecting...</p>
	</div>
</div>
