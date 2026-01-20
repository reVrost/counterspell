<script lang="ts">
	import { authAPI } from '$lib/api';
	import { appState } from '$lib/stores/app.svelte';

	$effect(async () => {
		// Clear auth state immediately
		appState.isAuthenticated = false;
		appState.userEmail = '';

		try {
			// Call backend logout to clear cookies
			await authAPI.logout();
		} catch (err) {
			console.error('Logout failed:', err);
		} finally {
			// Always redirect to home after logout
			window.location.href = '/';
		}
	});
</script>

<div class="h-screen flex items-center justify-center bg-[#0C0E12]">
	<div class="text-center space-y-4">
		<div class="w-12 h-12 border-4 border-purple-500 border-t-transparent rounded-full animate-spin mx-auto"></div>
		<p class="text-gray-400">Logging out...</p>
	</div>
</div>
