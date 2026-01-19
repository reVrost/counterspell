<script lang="ts">
	import { onMount } from "svelte";
	import "../app.css";
	import { QueryClient, QueryClientProvider } from "@tanstack/svelte-query";
	import { appState } from "$lib/stores/app.svelte";
	import { page } from "$app/stores";
	import { browser } from "$app/environment";
	import { authAPI } from "$lib/api";
	import { initGlobalErrorHandlers } from "$lib/utils/logger";

	const queryClient = new QueryClient({
		defaultOptions: {
			queries: {
				staleTime: 1000 * 60 * 5, // 5 minutes
				refetchOnWindowFocus: false,
				retry: false, // Don't retry on 401 auth errors
			},
		},
	});

	let { children } = $props();
	let isInitialized = $state(false);

	// Force logout function - only assign to window in browser
	if (browser) {
		(window as any).forceLogout = async () => {
			console.log("🔄 Force logout triggered");
			appState.isAuthenticated = false;
			appState.userEmail = "";
			localStorage.clear();
			sessionStorage.clear();

			// Clear cookies
			document.cookie.split(";").forEach((c) => {
				document.cookie = c
					.replace(/^ +/, "")
					.replace(
						/=.*/,
						"=;expires=" + new Date().toUTCString() + ";path=/",
					);
			});

			try {
				await fetch("/auth/logout", {
					method: "POST",
					credentials: "include",
				});
			} catch (e) {
				console.error("Logout error:", e);
			}

			window.location.href = "/";
		};
	}

	// Initialize app state on mount
	onMount(async () => {
		// Initialize global error handlers first
		initGlobalErrorHandlers();

		console.log("🚀 App layout mounting...");
		await appState.init();
		isInitialized = true;
		console.log(
			"✅ App state initialized, isAuth:",
			appState.isAuthenticated,
		);

		// Register service worker for PWA
		if ("serviceWorker" in navigator) {
			try {
				const registration =
					await navigator.serviceWorker.register("/sw.js");
				console.log(
					"✅ Service worker registered:",
					registration.scope,
				);
			} catch (err) {
				console.warn("Service worker registration failed:", err);
			}
		}
	});

	// Auth guard - handle authentication
	// With GitHub-only auth, isAuthenticated = githubConnected
	let hasRedirected = false;
	$effect(() => {
		if (!isInitialized || !browser || hasRedirected) return;

		const path = $page.url.pathname;

		console.log(
			"📍 Navigation:",
			path,
			"auth:",
			appState.isAuthenticated,
			"github:",
			appState.githubConnected,
		);

		// Dashboard requires authentication (GitHub connected)
		if (path.startsWith("/dashboard")) {
			if (!appState.isAuthenticated) {
				console.log("🔒 Not authenticated, redirecting to home...");
				hasRedirected = true;
				window.location.href = "/";
				return;
			}
		}

		// Landing page - if authenticated, redirect to dashboard
		if (path === "/") {
			if (appState.isAuthenticated) {
				console.log("✅ Authenticated, redirecting to dashboard...");
				hasRedirected = true;
				window.location.href = "/dashboard";
			}
		}
	});
</script>

<QueryClientProvider client={queryClient}>
	{@render children()}
</QueryClientProvider>
