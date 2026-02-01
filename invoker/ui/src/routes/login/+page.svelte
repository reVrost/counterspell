<script lang="ts">
    import { onMount } from "svelte";
    import { Loader2, ArrowRight } from "lucide-svelte";
    import { supabase } from "$lib/supabase";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";
    import * as Card from "$lib/components/ui/card";

    let email = $state("");
    let password = $state("");
    let loading = $state(false);
    let error = $state<string | null>(null);

    // Mouse tracking for spotlight effect
    let mouseX = $state(0);
    let mouseY = $state(0);

    function handleMouseMove(e: MouseEvent) {
        mouseX = e.clientX;
        mouseY = e.clientY;
    }

    onMount(() => {
        window.addEventListener("mousemove", handleMouseMove);
        return () => window.removeEventListener("mousemove", handleMouseMove);
    });

    async function handleLogin(e: Event) {
        e.preventDefault();
        loading = true;
        error = null;

        try {
            const res = await fetch("/api/v1/auth/login", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    email,
                    password,
                }),
            });

            if (!res.ok) {
                const data = await res.json();
                throw new Error(data.error || "Invalid credentials");
            }

            const data = await res.json();
            if (data.token) {
                localStorage.setItem("access_token", data.token);
            }
            window.location.href = "/";
        } catch (err: any) {
            error = err.message;
        } finally {
            loading = false;
        }
    }

    async function handleGoogleLogin() {
        loading = true;
        error = null;
        try {
            const { data, error: signInError } =
                await supabase.auth.signInWithOAuth({
                    provider: "google",
                    options: {
                        redirectTo: `${window.location.origin}/auth/callback`,
                        skipBrowserRedirect: false,
                    },
                });

            if (signInError) {
                throw new Error(signInError.message);
            }
        } catch (err: any) {
            error = err.message || "Failed to start Google login";
            loading = false;
        }
    }
</script>

<div
    class="relative min-h-screen flex flex-col items-center justify-center bg-black overflow-hidden font-sans selection:bg-primary/30"
>
    <!-- Background Grid -->
    <div
        class="fixed inset-0 z-0 pointer-events-none opacity-30"
        style="background-image: linear-gradient(#333 1px, transparent 1px), linear-gradient(90deg, #333 1px, transparent 1px); background-size: 40px 40px; mask-image: radial-gradient(circle at center, black 40%, transparent 100%); animation: grid-drift 60s linear infinite;"
    ></div>

    <!-- Interactive Spotlight (Brand Purple) -->
    <div
        class="fixed inset-0 z-0 pointer-events-none"
        style="background: radial-gradient(600px circle at {mouseX}px {mouseY}px, rgba(177, 151, 252, 0.1), transparent 80%);"
    ></div>

    <div
        class="relative z-10 w-full max-w-md px-4 py-12 flex flex-col items-center"
    >
        <div class="mb-8 flex flex-col items-center text-center">
            <!-- Brand Icon -->
            <div
                class="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-primary shadow-2xl shadow-primary/30 relative"
                style="animation: float 6s ease-in-out infinite;"
            >
                <div
                    class="absolute inset-0 rounded-2xl bg-primary blur-xl opacity-20"
                ></div>
                <svg
                    viewBox="0 0 24 24"
                    class="h-10 w-10 text-primary-foreground relative z-10"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2.5"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                >
                    <path
                        d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"
                        fill="currentColor"
                    />
                </svg>
            </div>
            <h1
                class="text-3xl font-bold tracking-tight sm:text-4xl text-white mb-2"
                style="animation: reveal 0.8s cubic-bezier(0.16, 1, 0.3, 1) forwards;"
            >
                Welcome back
            </h1>
            <p
                class="text-muted-foreground font-mono text-xs uppercase tracking-widest"
                style="animation: reveal 1s cubic-bezier(0.16, 1, 0.3, 1) forwards;"
            >
                Continue the mission
            </p>
        </div>

        <Card.Root
            class="w-full bg-black/40 backdrop-blur-xl border-white/10 shadow-2xl relative overflow-hidden"
            style="animation: reveal 1.2s cubic-bezier(0.16, 1, 0.3, 1) forwards;"
        >
            <div
                class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-primary/50 to-transparent"
            ></div>
            <Card.Content class="p-8 space-y-8">
                <form onsubmit={handleLogin} class="space-y-6">
                    <div class="space-y-2">
                        <label for="email" class="text-sm font-medium"
                            >Email</label
                        >
                        <Input
                            id="email"
                            type="email"
                            bind:value={email}
                            placeholder="Your email address"
                            required
                        />
                    </div>

                    <div class="space-y-2">
                        <label for="password" class="text-sm font-medium"
                            >Password</label
                        >
                        <Input
                            id="password"
                            type="password"
                            bind:value={password}
                            placeholder="••••••••"
                            required
                        />
                    </div>

                    {#if error}
                        <div
                            class="rounded-lg bg-destructive/10 p-3 text-sm text-destructive border border-destructive/20"
                        >
                            {error}
                        </div>
                    {/if}

                    <Button
                        type="submit"
                        disabled={loading}
                        class="w-full font-bold py-6 bg-white text-black hover:bg-white/90 rounded-none transition-all group"
                    >
                        {#if loading}
                            <Loader2 class="mr-2 h-4 w-4 animate-spin" />
                        {:else}
                            Sign In <ArrowRight
                                class="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform"
                            />
                        {/if}
                    </Button>
                </form>

                <div class="relative flex items-center py-4">
                    <div class="flex-grow border-t border-border/50"></div>
                    <span
                        class="mx-4 flex-shrink text-xs font-medium uppercase text-muted-foreground"
                        >OR</span
                    >
                    <div class="flex-grow border-t border-border/50"></div>
                </div>

                <Button
                    onclick={handleGoogleLogin}
                    disabled={loading}
                    variant="outline"
                    class="w-full gap-3 font-medium border-border/50 hover:bg-accent/50"
                >
                    <svg class="h-4 w-4" viewBox="0 0 24 24">
                        <path
                            d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                            fill="#4285F4"
                        />
                        <path
                            d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                            fill="#34A853"
                        />
                        <path
                            d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z"
                            fill="#FBBC05"
                        />
                        <path
                            d="M12 5.38c1.62 0 3.06.56 4.21 1.66l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 12-4.53z"
                            fill="#EA4335"
                        />
                    </svg>
                    Continue with Google
                </Button>

                <p class="mt-8 text-center text-sm text-muted-foreground">
                    Don't have an account?
                    <a
                        href="/signup"
                        class="font-medium text-primary hover:underline underline-offset-4"
                        >Sign up</a
                    >
                </p>
            </Card.Content>
        </Card.Root>

        <div
            class="mt-12 max-w-xs text-center text-[10px] leading-relaxed text-muted-foreground/40 font-mono uppercase tracking-tighter"
        >
            Access your coding agents from anywhere. <br />
            By signing in, you agree to our
            <a
                href="/terms"
                class="underline underline-offset-2 hover:text-white transition-colors"
                >Terms</a
            >.
        </div>
    </div>
</div>

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
            transform: translateY(0);
        }
        50% {
            transform: translateY(-10px);
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

    :global(.bg-primary) {
        --primary-rgb: 177, 151, 252;
    }
</style>
