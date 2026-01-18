<script lang="ts">
	import { appState } from '$lib/stores/app.svelte';
	import { Button } from '$lib/components/ui/button';
	import GithubIcon from '@lucide/svelte/icons/github';
	import LoaderIcon from '@lucide/svelte/icons/loader-2';

	let loading = $state(false);
</script>

<svelte:head>
	<title>Counterspell - AI-Native Engineering Orchestration</title>
</svelte:head>

<div
	class="fixed inset-0 z-[100] bg-background flex flex-col items-center justify-center text-center px-6"
>
	<!-- Background Effects -->
	<div class="absolute inset-0 overflow-hidden pointer-events-none">
		<div
			class="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-[100px] animate-pulse"
		></div>
		<div
			class="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-[100px] animate-pulse"
			style="animation-delay: 2s;"
		></div>
	</div>

	<!-- Content -->
	<div class="relative z-10 max-w-md w-full space-y-8">
		<div class="space-y-4">
			<img
				src="/icon-144.png"
				class="w-16 h-16 rounded-2xl mx-auto border border-gray-800 shadow-lg shadow-blue-500/20"
				alt="Counterspell Logo"
			/>
			<h1 class="text-3xl font-bold text-white tracking-tight">Welcome to Counterspell</h1>
			<p class="text-gray-400 text-sm leading-relaxed">
				Mobile-first, hosted AI agent Kanban.
				<br />
				Orchestrate from your pocket.
			</p>
		</div>

		<!-- Initial Action -->
		{#if !loading}
			<div>
				<a
					href="/api/v1/auth/oauth/github"
					onclick={() => (loading = true)}
					class="w-full bg-white text-black font-bold h-12 rounded-lg hover:bg-gray-200 transition active:scale-95 flex items-center justify-center gap-2"
				>
					<GithubIcon class="w-5 h-5" /> Continue with GitHub
				</a>
				<p class="mt-4 text-[10px] text-gray-600">
					By continuing, you agree to the Developer Protocol v2.1
				</p>
			</div>
		{:else}
			<!-- Loading Sequence -->
			<div class="space-y-4">
				<div
					class="bg-gray-900/50 rounded-xl p-4 border border-gray-800 text-left space-y-3 font-mono text-xs"
				>
					<div class="flex items-center gap-3">
						<div
							class="w-4 h-4 rounded-full flex items-center justify-center bg-purple-500/20 text-purple-400"
						>
							<LoaderIcon class="w-3 h-3 animate-spin" />
						</div>
						<span class="text-gray-200">Redirecting to GitHub...</span>
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- Footer -->
	<div class="absolute bottom-8 text-center">
		<p class="text-xs text-gray-600">
			<a
				href="https://github.com/revrost/counterspell"
				target="_blank"
				class="hover:text-gray-400 transition flex items-center justify-center gap-1"
			>
				<GithubIcon class="w-3 h-3" /> Open Source
			</a>
		</p>
	</div>
</div>
