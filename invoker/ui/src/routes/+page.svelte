<script lang="ts">
	import { onMount } from "svelte";
	import Dashboard from "$lib/components/Dashboard.svelte";
	import { Button } from "$lib/components/ui/button";
	import {
		ArrowRight,
		MoveRight,
		AlertTriangle,
		X,
	} from "lucide-svelte";
	import { Toaster, toast } from "svelte-sonner";
	import "../app.css";

	let status = "online";
	let isAuthenticated = $state(false);
	let loading = $state(true);
	let redirecting = $state(false);

	// Waitlist state
	let waitlistEmail = $state("");
	let isSubmitting = $state(false);

	// OAuth error state
	let oauthError = $state<{
		error: string;
		errorCode: string;
		errorDescription: string;
	} | null>(null);
	let showError = $state(true);

	// Mouse tracking for spotlight effect
	let mouseX = $state(0);
	let mouseY = $state(0);

	function handleMouseMove(e: MouseEvent) {
		mouseX = e.clientX;
		mouseY = e.clientY;
	}

	function clearError() {
		const url = new URL(window.location.href);
		url.searchParams.delete("error");
		url.searchParams.delete("error_code");
		url.searchParams.delete("error_description");
		window.history.replaceState({}, "", url.toString());
		oauthError = null;
		showError = false;
	}

	onMount(() => {
		// Check for OAuth error parameters
		const urlParams = new URLSearchParams(window.location.search);
		const error = urlParams.get("error");
		const errorCode = urlParams.get("error_code");
		const errorDescription = urlParams.get("error_description");

		if (error) {
			oauthError = {
				error,
				errorCode: errorCode || "",
				errorDescription: errorDescription || "An authentication error occurred.",
			};
		}

		const token = localStorage.getItem("access_token");
		isAuthenticated = !!token;

		const checkAuth = async () => {
			if (isAuthenticated && token) {
				try {
					const res = await fetch("/api/v1/machines", {
						headers: { Authorization: `Bearer ${token}` },
					});
					if (res.ok) {
						const data = await res.json();
						if (data?.machines?.length === 1) {
							const subdomain = data.machines[0].subdomain;
							if (subdomain) {
								redirecting = true;
								window.location.href = `https://${subdomain}.counterspell.app`;
								return;
							}
						}
					}
				} catch (err) {
					// Best-effort redirect only
				}
			}

			loading = false;
			document.documentElement.classList.add("dark");
		};

		checkAuth();

		window.addEventListener("mousemove", handleMouseMove);
		return () => window.removeEventListener("mousemove", handleMouseMove);
	});

	async function handleJoinWaitlist(e: Event) {
		e.preventDefault();
		if (!waitlistEmail) return;

		isSubmitting = true;
		try {
			const res = await fetch("/api/v1/waitlist", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ email: waitlistEmail }),
			});

			if (!res.ok) throw new Error("Failed to join");

			toast.success("Welcome to the hunt! You're on the list.", {
				description:
					"We'll reach out when it's your turn to enter the autonomous age.",
			});
			waitlistEmail = "";
		} catch (err) {
			toast.error("Something went wrong. Please try again later.");
		} finally {
			isSubmitting = false;
		}
	}
</script>

<svelte:head>
	<title>Counterspell - Mobile Agent Orchestration</title>
</svelte:head>

{#if loading || redirecting}
	<div class="min-h-screen bg-black flex items-center justify-center">
		<div
			class="w-8 h-8 border-2 border-white/10 border-t-primary rounded-full animate-spin"
		></div>
	</div>
{:else if isAuthenticated}
	<Dashboard />
{:else}
	<div
		class="min-h-screen bg-black text-white font-sans selection:bg-primary/30 overflow-x-hidden"
	>
		{#if oauthError && showError}
			<!-- OAuth Error Banner -->
			<div
				class="fixed top-0 left-0 right-0 z-[100] border-b border-red-500/20 backdrop-blur-xl bg-gradient-to-r from-red-950/90 via-black/90 to-black/90"
			>
				<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
					<div class="py-4 flex items-start gap-4">
						<div class="flex-shrink-0 mt-0.5">
							<div
								class="w-8 h-8 rounded-lg bg-red-500/10 border border-red-500/20 flex items-center justify-center"
							>
								<AlertTriangle class="w-4 h-4 text-red-400" />
							</div>
						</div>
						<div class="flex-1 min-w-0">
							<h3 class="text-sm font-semibold text-red-50 mb-1">
								Authentication Error
							</h3>
							<p
								class="text-xs text-red-200/70 leading-relaxed font-mono"
							>
								{oauthError.errorDescription ||
									"An error occurred during the authentication process."}
							</p>
							{#if oauthError.errorCode}
								<code
									class="mt-2 inline-block px-2 py-0.5 text-[10px] font-mono rounded bg-red-500/10 border border-red-500/20 text-red-300/80"
								>
									{oauthError.errorCode}
								</code>
							{/if}
						</div>
						<button
							onclick={clearError}
							class="flex-shrink-0 p-1.5 rounded-lg hover:bg-white/5 transition-colors group"
							aria-label="Dismiss error"
						>
							<X
								class="w-4 h-4 text-red-300/60 group-hover:text-red-200 transition-colors"
							/>
						</button>
					</div>
				</div>
			</div>
		{/if}

		<!-- Animated Background Grid -->
		<div
			class="fixed inset-0 z-0 pointer-events-none opacity-35"
			style="background-image: linear-gradient(#444 1px, transparent 1px), linear-gradient(90deg, #444 1px, transparent 1px); background-size: 40px 40px; mask-image: radial-gradient(circle at center, black 40%, transparent 100%); animation: grid-drift 60s linear infinite;"
		></div>

		<!-- Interactive Spotlight (Brand Purple) -->
		<div
			class="fixed inset-0 z-0 pointer-events-none"
			style="background: radial-gradient(600px circle at {mouseX}px {mouseY}px, rgba(177, 151, 252, 0.15), transparent 80%);"
		></div>

		<nav
			class="sticky top-0 z-50 bg-black/80 backdrop-blur-md border-b border-white/5"
			class:pt-16={oauthError && showError}
		>
			<div
				class="container mx-auto px-6 h-16 flex items-center justify-between"
			>
				<div
					class="flex items-center gap-8 text-sm font-mono text-muted-foreground"
				>
					<a
						href="#solutions"
						class="hover:text-white transition-colors"
						aria-label="Solutions"
					></a>
					<a
						href="#ecosystem"
						class="hover:text-white transition-colors"
						aria-label="Ecosystem"
					></a>
				</div>

				<div class="flex items-center gap-2">
					<span
						class="font-semibold tracking-[0.0em] text-lg uppercase"
						>Counter<span class="text-primary">spell</span></span
					>
				</div>

				<div class="flex items-center gap-6">
					<div
						class="hidden md:flex items-center gap-6 text-sm font-mono text-muted-foreground"
					>
						<a
							href="#developers"
							class="hover:text-white transition-colors"
							>&lt;&gt; Github</a
						>
						<a
							href="#resources"
							class="hover:text-white transition-colors"
							>&lt;&gt; Resources</a
						>
					</div>
					<div class="flex items-center gap-4">
						<!-- Auth links hidden for waitlist phase
						<a
							href="/login"
							class="text-xs font-mono font-bold text-white hover:text-primary transition-colors"
							>Sign In</a
						>
						<a
							href="/signup"
							class="bg-primary text-black hover:bg-primary/90 px-4 py-2 font-mono text-xs font-bold flex items-center gap-2 transition-all"
						>
							Get Started <ArrowRight class="w-3 h-3" />
						</a>
						-->
					</div>
				</div>
			</div>
		</nav>

		<main
			class="relative z-10 mx-auto border-x border-white/5 max-w-[1400px]"
		>
			<!-- Hero Section -->
			<div
				class="relative grid grid-cols-1 lg:grid-cols-12 min-h-[90vh] border-b border-white/5"
			>
				<!-- Sidebar Menu -->
				<div
					class="hidden lg:flex flex-col gap-6 p-8 col-span-2 border-r border-white/5 text-xs font-mono text-muted-foreground pt-32"
				>
					<a href="#intro" class="hover:text-white transition-colors"
						>01/ INTRODUCTION</a
					>
					<a
						href="#ecosystem"
						class="hover:text-white transition-colors"
						>02/ AUTONOMOUS ECOSYSTEM</a
					>
					<a
						href="#principles"
						class="hover:text-white transition-colors"
						>03/ CORE PRINCIPLES</a
					>
					<a href="#faq" class="hover:text-white transition-colors"
						>04/ FAQ's</a
					>
				</div>

				<!-- Main Hero Content -->
				<div
					class="col-span-1 lg:col-span-10 relative flex flex-col justify-between p-8 lg:p-24"
				>
					<!-- Geometric decorations -->
					<div
						class="absolute inset-0 overflow-hidden pointer-events-none"
					>
						<div
							class="absolute top-1/4 left-1/4 w-96 h-96 border border-primary/20 rounded-full opacity-40 transition-transform duration-1000"
							style="animation: float 20s ease-in-out infinite; box-shadow: 0 0 50px rgba(177, 151, 252, 0.1);"
						></div>
						<div
							class="absolute top-1/3 left-1/3 w-64 h-64 border border-white/20 rounded-full opacity-50"
							style="animation: float 15s ease-in-out infinite reverse; box-shadow: 0 0 30px rgba(255, 255, 255, 0.05);"
						></div>
						<div
							class="absolute top-1/2 right-1/4 w-32 h-32 border border-dashed border-primary/40 rounded-full"
							style="animation: pulse 4s ease-in-out infinite;"
						></div>

						<!-- Grid overlay patch -->
						<div
							class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[80%] h-[60%] border border-white/10 bg-white/[0.02]"
						></div>
					</div>

					<div class="flex-grow flex items-end mb-24 relative">
						<div
							class="max-w-4xl"
							style="animation: reveal 1s cubic-bezier(0.16, 1, 0.3, 1) forwards;"
						>
							<h1
								class="text-6xl md:text-8xl font-bold tracking-tight leading-[0.9] mb-8"
							>
								The Ticket-Based Workflow<br />
								<span class="text-white"
									>For Autonomous Coding</span
								>
							</h1>
						</div>
					</div>

					<div
						class="grid grid-cols-1 md:grid-cols-2 gap-8 items-end border-t border-white/10 pt-8"
					>
						<div
							class="text-muted-foreground font-mono text-sm flex items-center gap-2"
						>
							Enter the Autonomous Age <MoveRight
								class="w-4 h-4"
							/>
						</div>
						<div
							class="flex flex-col md:items-end gap-6 w-full max-w-md"
						>
							<form
								onsubmit={handleJoinWaitlist}
								class="flex w-full gap-0 border border-white/10 focus-within:border-primary/50 transition-colors"
							>
								<input
									type="email"
									bind:value={waitlistEmail}
									placeholder="Enter your email"
									required
									class="flex-1 bg-white/5 px-6 py-4 font-mono text-sm outline-none placeholder:text-muted-foreground/30"
								/>
								<button
									type="submit"
									disabled={isSubmitting}
									class="bg-white text-black hover:bg-primary hover:text-black px-4 py-4 font-mono text-sm font-bold flex items-center gap-2 transition-all disabled:opacity-50"
								>
									{isSubmitting
										? "Joining..."
										: "Join Waitlist"}
									{#if !isSubmitting}
										<ArrowRight
											class="w-4 h-4 text-black"
										/>
									{/if}
								</button>
							</form>
							<p
								class="font-mono text-[10px] uppercase tracking-widest text-muted-foreground/40 text-right"
							>
								Limited slots available for early access.
							</p>
						</div>
					</div>
				</div>
			</div>

			<!-- Intro Section -->
			<div
				id="intro"
				class="grid grid-cols-1 lg:grid-cols-12 min-h-[60vh] border-b border-white/5 bg-black"
			>
				<div
					class="hidden lg:block col-span-2 border-r border-white/5 p-8 text-xs font-mono text-muted-foreground"
				>
					01/ INTRODUCTION
				</div>
				<div
					class="col-span-1 lg:col-span-10 p-8 lg:p-24 flex flex-col justify-center"
				>
					<p
						class="text-3xl md:text-5xl leading-tight font-light max-w-5xl mb-24"
					>
						Stop fighting chat windows. Counterspell gives your
						agents a professional task-based workflow.
						<span class="text-muted-foreground"
							>Execute on your local machine for security, or spin
							up ephemeral cloud instances for always-on autonomy.</span
						>
					</p>

					<div
						class="grid grid-cols-1 md:grid-cols-3 border border-white/10"
					>
						<div
							class="p-8 border-b md:border-b-0 md:border-r border-white/10"
						>
							<div class="text-5xl font-mono mb-2">38+</div>
							<div
								class="text-sm font-mono text-muted-foreground"
							>
								Tool Integrations in Progress
								<!-- Repositories Connected-->
							</div>
						</div>
						<div
							class="p-8 border-b md:border-b-0 md:border-r border-white/10"
						>
							<div class="text-5xl font-mono mb-2">150+</div>
							<div
								class="text-sm font-mono text-muted-foreground"
							>
								Skills Connected
							</div>
						</div>
						<div class="p-8">
							<div class="text-5xl font-mono mb-2">52+</div>
							<div
								class="text-sm font-mono text-muted-foreground"
							>
								Teams on the Waitlist
								<!-- Inspired Developer Teams -->
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Ecosystem Section -->
			<div
				id="ecosystem"
				class="grid grid-cols-1 lg:grid-cols-12 min-h-[55vh] border-b border-white/5 bg-black relative"
			>
				<div
					class="hidden lg:block col-span-2 border-r border-white/5 p-8 text-xs font-mono text-muted-foreground"
				>
					02/ AUTONOMOUS ECOSYSTEM
				</div>
				<div class="col-span-1 lg:col-span-10 p-8 lg:p-12">
					<div class="flex justify-end mb-16 text-xs font-mono">
						&lt;/&gt; Parallel Executions. <span
							class="text-white ml-2">Zero Conflicts.</span
						>
					</div>

					<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
						<!-- Card 1 -->
						<div
							class="bg-black border border-white/10 p-8 hover:border-primary/50 transition-colors group"
						>
							<h3 class="text-xl font-bold mb-4">
								Multi-Agent Orchestration
							</h3>
							<p
								class="text-muted-foreground text-sm mb-12 leading-relaxed"
							>
								Spin up parallel agents for complex tickets.
								Monitor progress, token usage, and execution
								logs in real-time.
							</p>
							<div
								class="relative h-22 w-full flex items-end gap-[2px]"
							>
								{#each Array(40) as _, i}
									<div
										class="flex-1 bg-white/20 group-hover:bg-primary/80 transition-all duration-500"
										style="height: {20 +
											Math.random() * 80}%"
									></div>
								{/each}
							</div>
							<div
								class="flex justify-between text-xs font-mono text-muted-foreground mt-4"
							>
								<span>Processing Ticket...</span>
								<span>64%</span>
							</div>
						</div>

						<!-- Card 2 -->
						<div
							class="bg-black border border-white/10 p-8 hover:border-primary/50 transition-colors group"
						>
							<h3 class="text-xl font-bold mb-4">
								High-Fidelity Code Reviews
							</h3>
							<p
								class="text-muted-foreground text-sm mb-12 leading-relaxed"
							>
								Built-in split-view diffs allow you to review
								Every. Single. Change. Approve plans before they
								execute and verify code before it merges.
							</p>
							<div class="flex items-end gap-1 h-16 mb-4">
								{#each Array(20) as _, i}
									<div
										class="w-2 bg-white group-hover:bg-primary transition-colors"
										style="height: {30 +
											Math.random() * 70}%"
									></div>
								{/each}
							</div>
							<div
								class="flex justify-between text-xs font-mono text-muted-foreground border-t border-white/10 pt-2"
							>
								<span>Approval Rate</span>
								<span>98%</span>
							</div>
						</div>

						<!-- Card 3 -->
						<div
							class="bg-black border border-white/10 p-8 hover:border-primary/50 transition-colors group"
						>
							<h3 class="text-xl font-bold mb-4">
								Global Remote Control
							</h3>
							<p
								class="text-muted-foreground text-sm mb-12 leading-relaxed"
							>
								Run the brain on your laptop, access the UI from
								anywhere. Secure tunnels give you remote control
								without sacrificing data sovereignty.
							</p>
							<div class="flex items-center justify-center py-4">
								<div class="flex gap-4">
									<div
										class="relative w-12 h-12 rounded-full border border-white/30 flex items-center justify-center"
									>
										<div
											class="w-2 h-2 bg-white rounded-full group-hover:bg-primary transition-colors"
										></div>
										<div
											class="absolute inset-0 border border-white/10 rounded-full scale-150"
										></div>
									</div>
									<div
										class="relative w-12 h-12 rounded-full border border-white/30 flex items-center justify-center"
									>
										<div
											class="w-2 h-2 bg-white rounded-full group-hover:bg-primary transition-colors"
										></div>
										<div
											class="absolute inset-0 border border-white/10 rounded-full scale-150"
										></div>
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Core Principles -->
			<div
				id="principles"
				class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 border-b border-white/5 bg-black"
			>
				<!-- Header Block -->
				<div
					class="p-8 border-r border-white/5 flex flex-col justify-between"
				>
					<div class="text-xs font-mono text-muted-foreground">
						02/ CORE PRINCIPLES
					</div>
					<div class="mt-auto">
						<div
							class="text-xs font-mono text-right text-muted-foreground mb-4"
						>
							///// Building Blocks of Local-First
						</div>
					</div>
				</div>

				<!-- Principle 1 -->
				<div
					class="p-8 border-r border-white/5 flex flex-col gap-4 hover:bg-white/[0.02] transition-colors group"
				>
					<div
						class="w-12 h-12 border border-white/20 grid grid-cols-2 gap-1 p-1"
					>
						<div class="bg-white/10"></div>
						<div class="bg-white/10"></div>
						<div class="bg-primary/50"></div>
						<div class="bg-white/10"></div>
					</div>
					<div>
						<h3 class="text-xl font-bold mb-4">
							Local-First by Design
						</h3>
						<p
							class="text-sm text-muted-foreground leading-relaxed"
						>
							Your code never leaves your machine. Execute agents
							locally with full control over your data and
							privacy.
						</p>
					</div>
				</div>

				<!-- Principle 2 -->
				<div
					class="p-8 border-r border-white/5 flex flex-col gap-4 hover:bg-white/[0.02] transition-colors group"
				>
					<div
						class="w-12 h-12 border border-dashed border-white/20 relative"
					>
						<div
							class="absolute bottom-1 left-1 w-2 h-2 bg-white"
						></div>
					</div>
					<div>
						<h3 class="text-xl font-bold mb-4">
							Control & Transparency
						</h3>
						<p
							class="text-sm text-muted-foreground leading-relaxed"
						>
							Review every change before it happens with clear
							diff views. No surprise modifications.
						</p>
					</div>
				</div>

				<!-- Principle 3 -->
				<div
					class="p-8 flex flex-col hover:bg-white/[0.02] transition-colors group gap-4"
				>
					<div
						class="w-12 h-12 relative flex items-center justify-center border border-dashed border-white/20"
					>
						<div
							class="w-0 h-0 border-l-[6px] border-l-transparent border-r-[6px] border-r-transparent border-b-[12px] border-b-white/50"
						></div>
					</div>
					<div>
						<h3 class="text-xl font-bold mb-4">
							Accessible Everywhere
						</h3>
						<p
							class="text-sm text-muted-foreground leading-relaxed"
						>
							Manage your agents from any device with a secure
							tunnel to your local environment.
						</p>
					</div>
				</div>
			</div>

			<!-- CTA Section (Explore Principles) -->
			<div class="border-b border-white/5 p-8 flex justify-center py-24">
				<div
					class="grid grid-cols-1 md:grid-cols-2 gap-12 max-w-4xl w-full items-center"
				>
					<p class="text-muted-foreground font-mono text-sm max-w-sm">
						The foundation driving local-first, privacy-focused, and
						trustworthy autonomous agent workflows.
					</p>
					<Button
						variant="outline"
						class="rounded-none h-14 border-white text-white hover:bg-white hover:text-black font-mono text-sm group justify-between px-6"
					>
						Explore Principles
						<ArrowRight
							class="w-4 h-4 group-hover:translate-x-1 transition-transform"
						/>
					</Button>
				</div>
			</div>

			<!-- Footer -->
			<footer
				class="grid grid-cols-1 lg:grid-cols-12 min-h-[50vh] bg-black pt-16 pb-8"
			>
				<div
					class="lg:col-span-4 px-8 flex flex-col justify-between border-r border-white/5"
				>
					<div>
						<div class="flex items-center gap-2 mb-8">
							<span class="font-bold tracking-widest text-base"
								>Counterspell</span
							>
						</div>
						<p
							class="text-muted-foreground text-sm leading-relaxed max-w-sm mb-12"
						>
							An autonomous coding environment that gives you
							control over your agents with ticket-based
							workflows, transparent execution, and secure remote
							access.
						</p>
					</div>

					<form onsubmit={handleJoinWaitlist} class="flex gap-0">
						<input
							type="email"
							bind:value={waitlistEmail}
							placeholder="Enter your email"
							required
							class="flex-1 border border-white/10 bg-black px-4 font-mono text-xs text-white placeholder:text-muted-foreground/50 h-12 outline-none focus:border-primary/50 transition-colors"
						/>
						<button
							type="submit"
							disabled={isSubmitting}
							class="bg-white text-black px-6 font-mono text-xs font-bold h-12 hover:bg-primary transition-colors disabled:opacity-50"
						>
							{isSubmitting ? "..." : "Subscribe"}
						</button>
					</form>
				</div>

				<div
					class="lg:col-span-8 grid grid-cols-1 md:grid-cols-3 gap-8 px-8 md:px-16 pt-2 mt-4"
				>
					<div class="flex flex-col gap-6">
						<h4
							class="text-xs font-bold text-white/50 uppercase tracking-widest font-mono"
						>
							Social Media
						</h4>
						<nav
							class="flex flex-col gap-4 text-sm font-mono text-muted-foreground"
						>
							<a
								href="https://instagram.com"
								target="_blank"
								rel="noopener noreferrer"
								class="hover:text-white flex items-center gap-2"
								><span class="text-xs">&lt;&gt;</span> Instagram</a
							>
							<a
								href="https://youtube.com"
								target="_blank"
								rel="noopener noreferrer"
								class="hover:text-white flex items-center gap-2"
								><span class="text-xs">&lt;&gt;</span> Youtube</a
							>
							<a
								href="https://discord.com"
								target="_blank"
								rel="noopener noreferrer"
								class="hover:text-white flex items-center gap-2"
								><span class="text-xs">&lt;&gt;</span> Discord</a
							>
							<a
								href="https://x.com"
								target="_blank"
								rel="noopener noreferrer"
								class="hover:text-white flex items-center gap-2"
								><span class="text-xs">&lt;&gt;</span> X</a
							>
						</nav>
					</div>
					<div class="flex flex-col gap-6">
						<h4
							class="text-xs font-bold text-white/50 uppercase tracking-widest font-mono"
						>
							Product
						</h4>
						<nav
							class="flex flex-col gap-4 text-sm font-mono text-muted-foreground"
						>
							<a href="/how-it-works" class="hover:text-white">How It Works</a
							>
							<a href="/use-cases" class="hover:text-white">Use Cases</a>
							<a href="/docs" class="hover:text-white"
								>Documentation</a
							>
							<a href="/start-building" class="hover:text-white"
								>Start Building</a
							>
						</nav>
					</div>
					<div class="flex flex-col gap-6">
						<h4
							class="text-xs font-bold text-white/50 uppercase tracking-widest font-mono"
						>
							Legal
						</h4>
						<nav
							class="flex flex-col gap-4 text-sm font-mono text-muted-foreground"
						>
							<a href="/terms" class="hover:text-white"
								>Terms & Conditions</a
							>
							<a href="/privacy" class="hover:text-white"
								>Privacy Policy</a
							>
							<a href="/security" class="hover:text-white"
								>Security Policy</a
							>
							<a href="/licensing" class="hover:text-white"
								>Licensing & Regulations</a
							>
						</nav>
					</div>
				</div>

				<div class="lg:col-span-12 px-8 mt-16 flex justify-end">
					<button
						onclick={() =>
							window.scrollTo({ top: 0, behavior: "smooth" })}
						class="text-xs font-mono text-muted-foreground hover:text-white flex items-center gap-2 group"
					>
						///// Back to Top <ArrowRight
							class="w-4 h-4 -rotate-45 group-hover:-translate-y-1 transition-transform"
						/>
					</button>
				</div>
			</footer>
		</main>
	</div>
{/if}

<Toaster theme="dark" position="bottom-right" />

<style>
	@keyframes grid-drift {
		0% {
			background-position: 0 0;
		}
		100% {
			background-position: 40px 40px;
		}
	}

	@keyframes float {
		0%,
		100% {
			transform: translateY(0) scale(1);
		}
		50% {
			transform: translateY(-20px) scale(1.05);
		}
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 0.2;
			transform: scale(1);
		}
		50% {
			opacity: 0.5;
			transform: scale(1.1);
		}
	}

	@keyframes reveal {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	:global(html) {
		scroll-behavior: smooth;
	}

	:global(.bg-primary) {
		--primary-rgb: 255, 255, 255; /* Fallback to white glow for primary */
	}
</style>
