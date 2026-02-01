<script lang="ts">
	import { appState } from '$lib/stores/app.svelte';
	import { page } from '$app/stores';

	$effect(() => {
		// Get auth status from URL parameters if present
		const params = new URLSearchParams(window.location.search);

		if (params.get('error')) {
			// OAuth error - redirect to home with error
			window.location.href = `/?error=${params.get('error')}`;
			return;
		}

		// OAuth callback was successful - redirect to dashboard
		// The auth cookie should already be set by the backend
		appState.isAuthenticated = true;

		// Redirect to dashboard
		window.location.href = '/dashboard';
	});
</script>

<div class="h-screen flex items-center justify-center bg-[#0C0E12]">
	<div class="text-center space-y-4">
		<div class="w-12 h-12 border-4 border-purple-500 border-t-transparent rounded-full animate-spin mx-auto"></div>
		<p class="text-gray-400">Completing login...</p>
	</div>
</div>
