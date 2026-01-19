<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { authAPI } from '$lib/api';
	import { appState } from '$lib/stores/app.svelte';
	import { signInWithGitHub } from '$lib/supabase';
	import GithubIcon from '@lucide/svelte/icons/github';
	import XIcon from '@lucide/svelte/icons/x';
	import { browser } from '$app/environment';

	let loading = $state(false);
	let errorMsg = $state('');
	let showError = $state(false);
	let checkingAuth = $state(true);
	let needsReset = $state(false);

	async function handleGitHubLogin() {
		loading = true;
		showError = false;
		try {
			await signInWithGitHub();
		} catch (err) {
			console.error('Login failed:', err);
			loading = false;
			errorMsg = 'Failed to initiate GitHub login';
			showError = true;
		}
	}

	async function clearAllCookies() {
		if (!browser) return;

		// Clear all cookies
		document.cookie.split(';').forEach((c) => {
			const domain = window.location.hostname;
			const domains = [domain, `.${domain}`, 'localhost'];
			domains.forEach((d) => {
				document.cookie = c
					.replace(/^ +/, '')
					.replace(/=.*/, `=;expires=${new Date(0).toUTCString()};path=/;domain=${d}`);
				document.cookie = c
					.replace(/^ +/, '')
					.replace(/=.*/, `=;expires=${new Date(0).toUTCString()};path=/;`);
			});
		});

		console.log('✅ All cookies cleared');
	}



	function dismissError() {
		showError = false;
		errorMsg = '';
	}

	onMount(async () => {
		if (!browser) return;

		console.log('📍 Landing page mounted, checking auth...');

		// Check for OAuth errors in URL
		const urlParams = new URLSearchParams(window.location.search);
		const error = urlParams.get('error');
		const errorDesc = urlParams.get('error_description');

		if (error) {
			errorMsg = errorDesc || `Login error: ${error}`;
			showError = true;
			// Clear error from URL
			window.history.replaceState({}, '', '/');
			checkingAuth = false;
			return;
		}

		// Check if already authenticated
		try {
			const session = await authAPI.checkSession();
			console.log('✅ Auth check result:', session);

			if (session.authenticated) {
				console.log('🚀 Redirecting to dashboard...');
				window.location.href = '/dashboard';
			} else {
				console.log('❓ Not authenticated, staying on landing page');
			}
		} catch (e) {
			console.log('❌ Auth check failed:', e);
			// Check if it's a 401 error (token expired)
			if (e instanceof Error && e.message.includes('401')) {
				console.log('⚠️ Token expired, clearing cookies...');
				await clearAllCookies();
			}
			// Not authenticated, stay on landing page
		} finally {
			checkingAuth = false;
		}
	});
</script>

<svelte:head>
	{#if browser}
		<script>
			// Global reset function
			// window.forceReset = async () => {
			// 	console.log('🔄 Global force reset...');
			// 	localStorage.clear();
			// 	sessionStorage.clear();
			//
			// 	// Clear all cookies
			// 	document.cookie.split(';').forEach((c) => {
			// 		const domain = window.location.hostname;
			// 		const domains = [domain, `.${domain}`, 'localhost'];
			// 		domains.forEach((d) => {
			// 			document.cookie = c
			// 				.replace(/^ +/, '')
			// 				.replace(/=.*/, `=;expires=${new Date(0).toUTCString()};path=/;domain=${d}`);
			// 			document.cookie = c
			// 				.replace(/^ +/, '')
			// 				.replace(/=.*/, `=;expires=${new Date(0).toUTCString()};path=/;`);
			// 		});
			// 	});
			//
			// 	console.log('✅ Reset complete, reloading...');
			// 	setTimeout(() => window.location.reload(), 500);
			// };
		</script>
	{/if}
</svelte:head>

<div class="h-screen flex flex-col overflow-hidden bg-[#0C0E12]">
	<!-- Background Effects -->
	<div class="absolute inset-0 overflow-hidden pointer-events-none">
		<div class="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-[100px] animate-pulse"></div>
		<div class="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-[100px] animate-pulse" style="animation-delay: 2s;"></div>
	</div>

	<!-- Error Toast -->
	{#if showError}
		<div class="fixed top-4 left-1/2 -translate-x-1/2 z-[200] bg-red-500/90 backdrop-blur text-white px-4 py-3 rounded-lg shadow-lg flex items-center gap-3 animate-in fade-in slide-in-from-top-4 duration-300">
			<XIcon class="w-4 h-4 cursor-pointer hover:opacity-80" onclick={dismissError} />
			<span class="text-sm font-medium">{errorMsg}</span>
		</div>
	{/if}

	<!-- Landing Content -->
	<div class="fixed inset-0 z-[100] bg-[#0C0E12] flex flex-col items-center justify-center text-center px-6">
		<!-- Content -->
		<div class="relative z-10 max-w-md w-full space-y-8">
			{#if checkingAuth}
				<!-- Loading state while checking auth -->
				<div class="space-y-4">
					<div class="w-16 h-16 border-4 border-purple-500 border-t-transparent rounded-full animate-spin mx-auto"></div>
					<p class="text-gray-400">Checking authentication...</p>
				</div>
			{:else}
				<div class="space-y-4">
					<div class="w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl mx-auto flex items-center justify-center shadow-lg shadow-blue-500/20 mb-6">
						<GithubIcon class="w-8 h-8 text-white" />
					</div>
					<h1 class="text-3xl font-bold text-white tracking-tight">Welcome to Counterspell</h1>
					<p class="text-gray-400 text-sm leading-relaxed">
						Mobile-first, hosted AI agent Kanban.
						<br />
						Orchestrate from your pocket.
					</p>
				</div>

				<!-- GitHub Login Button -->
				{#if !loading}
					<div>
						<button
							onclick={handleGitHubLogin}
							class="w-full bg-white text-black font-bold h-12 rounded-lg hover:bg-gray-200 transition active:scale-95 flex items-center justify-center gap-2"
						>
							<GithubIcon class="w-5 h-5" />
							Continue with GitHub
						</button>
						<p class="mt-4 text-[10px] text-gray-600">By continuing, you agree to Developer Protocol v2.1</p>
					</div>
				{:else}
					<!-- Loading State -->
					<div class="space-y-4">
						<div class="bg-gray-900/50 rounded-xl p-4 border border-gray-800 text-left space-y-3 font-mono text-xs">
							<div class="flex items-center gap-3">
								<div class="w-4 h-4 rounded-full flex items-center justify-center bg-purple-500/20 text-purple-400">
									<i class="fas fa-circle-notch fa-spin"></i>
								</div>
								<span class="text-gray-200">Redirecting to GitHub...</span>
							</div>
						</div>
					</div>
				{/if}
			{/if}
		</div>

		<!-- Footer -->
		<div class="absolute bottom-8 text-center space-y-2">
			<p class="text-xs text-gray-600">
				<a href="https://github.com/revrost/counterspell" target="_blank" class="hover:text-gray-400 transition">
					<GithubIcon class="w-3 h-3 inline mr-1" />
					Open Source
				</a>
			</p>
		</div>
	</div>
</div>

<style>
	:global(body) {
		-webkit-font-smoothing: antialiased;
		-moz-osx-font-smoothing: grayscale;
	}
	:global(button, :global(a)) {
		-webkit-tap-highlight-color: transparent;
		touch-action: manipulation;
	}
</style>
