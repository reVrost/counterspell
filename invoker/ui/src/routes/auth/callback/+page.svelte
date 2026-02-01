<script lang="ts">
    import { onMount } from 'svelte';
    import { Loader2 } from 'lucide-svelte';
    import { supabase } from '$lib/supabase';

    let loading = $state(true);
    let error = $state<string | null>(null);

    onMount(async () => {
        try {
            // Get session from Supabase SDK (handles URL params automatically)
            const { data: { session }, error: sessionError } = await supabase.auth.getSession();

            if (sessionError) {
                throw new Error(sessionError.message);
            }

            if (!session) {
                throw new Error('No session found');
            }

            const accessToken = session.access_token;
            const refreshToken = session.refresh_token;

            // Store tokens
            localStorage.setItem('access_token', accessToken);
            if (refreshToken) {
                localStorage.setItem('refresh_token', refreshToken);
            }

            // Sync user to backend database
            const apiBase = import.meta.env.VITE_API_URL || '';
            const res = await fetch(`${apiBase}/api/v1/auth/profiles`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${accessToken}`
                }
            });

            if (!res.ok) {
                console.error('Failed to sync user to backend');
                // Continue anyway since we have the token
            }

            // Auto-redirect to machine if only one exists
            try {
                const machinesRes = await fetch(`${apiBase}/api/v1/machines`, {
                    headers: { 'Authorization': `Bearer ${accessToken}` }
                });
                if (machinesRes.ok) {
                    const data = await machinesRes.json();
                    if (data?.machines?.length === 1) {
                        const subdomain = data.machines[0].subdomain;
                        if (subdomain) {
                            window.location.href = `https://${subdomain}.counterspell.app`;
                            return;
                        }
                    }
                }
            } catch (_) {
                // Best-effort; fall back to default routing
            }

            const next = new URL(window.location.href).searchParams.get('next');
            window.location.href = next || '/';
        } catch (err: any) {
            error = err.message || 'Authentication failed';
            console.error('Auth callback error:', err);
        } finally {
            loading = false;
        }
    });
</script>

<div class="flex min-h-screen flex-col items-center justify-center bg-black px-4 text-white">
    {#if loading}
        <div class="flex flex-col items-center gap-4">
            <Loader2 class="h-8 w-8 animate-spin text-orange-500" />
            <p class="text-zinc-400">Completing authentication...</p>
        </div>
    {:else if error}
        <div class="max-w-md rounded-lg border border-red-500/20 bg-red-500/10 p-6 text-center">
            <p class="text-red-400">{error}</p>
            <a
                href="/login"
                class="mt-4 inline-block rounded-lg bg-orange-500 px-4 py-2 font-medium text-white hover:bg-orange-600"
            >
                Try again
            </a>
        </div>
    {/if}
</div>
