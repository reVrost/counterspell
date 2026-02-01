<script lang="ts">
    import { onMount } from "svelte";
    import {
        Loader2,
        Copy,
        RotateCcw,
        Info,
        User,
        LayoutGrid,
        Trophy,
        ShoppingCart,
        LogOut,
        ChevronDown,
    } from "lucide-svelte";
    import * as Avatar from "$lib/components/ui/avatar";
    import * as Card from "$lib/components/ui/card";
    import { Button } from "$lib/components/ui/button";
    import { Badge } from "$lib/components/ui/badge";
    import { Separator } from "$lib/components/ui/separator";
    import * as Select from "$lib/components/ui/select";
    import * as Collapsible from "$lib/components/ui/collapsible";
    import * as Tabs from "$lib/components/ui/tabs";
    import { Input } from "$lib/components/ui/input";
    import { Checkbox } from "$lib/components/ui/checkbox";

    interface UserProfile {
        id: string;
        email: string;
        username: string;
        first_name: string;
        last_name: string;
        tier: string;
    }

    let profile = $state<UserProfile | null>({
        id: "mock-id",
        email: "revrost@gmail.com",
        username: "revrost",
        first_name: "Kenley",
        last_name: "Bastari",
        tier: "free",
    });
    let loading = $state(false);
    let error = $state<string | null>(null);
    let activeTab = $state("7d");
    let isAdvancedOpen = $state(true);
    let view = $state<"threads" | "settings">("threads");

    const mockThreads = [
        {
            id: 1,
            title: "Untitled",
            private: true,
            author: "revrost",
            time: "2d ago",
            messages: 1,
            repo: "reVrost/counterspell:main",
            stars: 0,
            snippet: "do you have access to playwright mcp?",
        },
    ];

    onMount(async () => {
        // Data fetching disabled for mock view
        /*
        try {
            const token = localStorage.getItem("access_token");
            if (!token) {
                window.location.href = "/login";
                return;
            }

            const apiBase = import.meta.env.VITE_API_URL || "";
            const res = await fetch(`${apiBase}/api/v1/auth/profile`, {
                headers: {
                    Authorization: `Bearer ${token}`,
                },
            });

            if (!res.ok) {
                if (res.status === 401) {
                    localStorage.removeItem("access_token");
                    window.location.href = "/login";
                    return;
                }
                throw new Error("Failed to fetch profile");
            }

            profile = await res.json();
        } catch (err: any) {
            error = err.message || "Something went wrong";
        } finally {
            loading = false;
        }
        */
    });

    function handleLogout() {
        localStorage.removeItem("access_token");
        window.location.href = "/";
    }

    function copyToken() {
        // Mock token copy
        navigator.clipboard.writeText("amp_live_xxxxxxxxxxxxxxxxxxxxxxxx");
        alert("Token copied to clipboard!");
    }
</script>

<div
    class="min-h-screen bg-background text-foreground font-sans selection:bg-primary/30"
>
    <!-- Navigation Bar -->
    <nav
        class="h-14 border-b border-border/10 flex items-center justify-between px-4 sm:px-6 bg-background/50 backdrop-blur-md sticky top-0 z-50"
    >
        <div class="flex items-center gap-4 sm:gap-6 overflow-hidden">
            <div class="flex items-center gap-2 flex-shrink-0">
                <span
                    class="font-bold tracking-tight text-base sm:text-lg hidden xs:block"
                    >Counterspell</span
                >
            </div>
            <div
                class="flex items-center gap-2 sm:gap-4 text-sm font-medium overflow-x-auto no-scrollbar py-1"
            >
                <Button
                    variant={view === "threads" ? "secondary" : "ghost"}
                    size="sm"
                    onclick={() => (view = "threads")}
                    class="gap-1.5 h-8 {view === 'threads'
                        ? 'bg-white/5 text-white'
                        : 'text-zinc-400'}"
                >
                    <LayoutGrid class="w-3.5 h-3.5" />
                    Threads
                </Button>
            </div>
        </div>

        <div class="flex items-center gap-1 sm:gap-3 flex-shrink-0">
            <Button
                variant="ghost"
                size="sm"
                href="/leaderboard"
                class="text-zinc-400 hover:text-white gap-1.5 hidden sm:flex"
            >
                <Trophy class="w-3.5 h-3.5" />
                Leaderboard
            </Button>
            <Button
                variant="ghost"
                size="sm"
                href="/marketpal"
                class="text-zinc-400 hover:text-white gap-1.5 hidden sm:flex"
            >
                <ShoppingCart class="w-3.5 h-3.5" />
                marketpal
            </Button>
            <Button
                variant="ghost"
                size="sm"
                onclick={handleLogout}
                class="text-zinc-400 hover:text-red-400 gap-1.5"
            >
                <LogOut class="w-3.5 h-3.5" />
                <span class="hidden sm:inline">Logout</span>
            </Button>

            <Button
                variant="ghost"
                size="icon"
                onclick={() => (view = "settings")}
                class="h-8 w-8 rounded-full border transition-colors {view ===
                'settings'
                    ? 'border-primary ring-2 ring-primary/20'
                    : 'border-white/20 hover:border-white/40'}"
            >
                <Avatar.Root class="h-full w-full">
                    <Avatar.Image
                        src="https://api.dicebear.com/7.x/avataaars/svg?seed={profile?.username ||
                            'user'}"
                    />
                    <Avatar.Fallback
                        >{profile?.username?.slice(0, 2).toUpperCase() ||
                            "AM"}</Avatar.Fallback
                    >
                </Avatar.Root>
            </Button>
        </div>
    </nav>

    <main class="max-w-5xl mx-auto px-4 sm:px-6 py-8 sm:py-12">
        {#if loading}
            <div class="flex items-center justify-center py-20">
                <Loader2 class="w-8 h-8 animate-spin text-primary" />
            </div>
        {:else if error}
            <div
                class="bg-red-500/10 border border-red-500/20 p-6 rounded-xl text-center"
            >
                <p class="text-red-400 mb-4">{error}</p>
                <button
                    onclick={() => window.location.reload()}
                    class="px-4 py-2 bg-white text-black text-sm font-semibold rounded-lg hover:bg-zinc-200 transition-colors"
                >
                    Retry
                </button>
            </div>
        {:else if profile}
            {#if view === "threads"}
                <!-- Threads View -->
                <div class="space-y-8">
                    <!-- Filters -->
                    <div
                        class="flex flex-col md:flex-row items-stretch md:items-center gap-4"
                    >
                        <div class="relative flex-1">
                            <span
                                class="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500 z-10 pointer-events-none"
                            >
                                <svg
                                    class="w-4 h-4"
                                    fill="none"
                                    stroke="currentColor"
                                    stroke-width="2.5"
                                    viewBox="0 0 24 24"
                                >
                                    <path
                                        d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                                    />
                                </svg>
                            </span>
                            <Input
                                placeholder="Search threads..."
                                class="pl-10 bg-zinc-900/50 border-white/5 focus-visible:ring-white/10"
                            />
                        </div>
                        <div class="flex flex-wrap gap-2">
                            <Select.Root type="single">
                                <Select.Trigger
                                    class="w-full md:w-[140px] bg-zinc-900/50 border-white/5"
                                >
                                    <span data-slot="select-value"
                                        >All users</span
                                    >
                                </Select.Trigger>
                                <Select.Content
                                    class="bg-zinc-900 border-white/10 text-white"
                                >
                                    <Select.Item value="all" label="All users"
                                        >All users</Select.Item
                                    >
                                </Select.Content>
                            </Select.Root>

                            <Select.Root type="single">
                                <Select.Trigger
                                    class="w-full md:w-[160px] bg-zinc-900/50 border-white/5"
                                >
                                    <span data-slot="select-value"
                                        >All repositories</span
                                    >
                                </Select.Trigger>
                                <Select.Content
                                    class="bg-zinc-900 border-white/10 text-white"
                                >
                                    <Select.Item
                                        value="all"
                                        label="All repositories"
                                        >All repositories</Select.Item
                                    >
                                </Select.Content>
                            </Select.Root>

                            <Select.Root type="single">
                                <Select.Trigger
                                    class="w-full md:w-[160px] bg-zinc-900/50 border-white/5"
                                >
                                    <span data-slot="select-value"
                                        >All thread types</span
                                    >
                                </Select.Trigger>
                                <Select.Content
                                    class="bg-zinc-900 border-white/10 text-white"
                                >
                                    <Select.Item
                                        value="all"
                                        label="All thread types"
                                        >All thread types</Select.Item
                                    >
                                </Select.Content>
                            </Select.Root>
                        </div>
                    </div>

                    <!-- Thread List -->
                    <div class="space-y-4">
                        {#each mockThreads as thread}
                            <div
                                class="group py-6 border-b border-white/5 last:border-0 overflow-hidden"
                            >
                                <div class="flex items-start gap-3 sm:gap-4">
                                    <Avatar.Root
                                        class="h-8 w-8 sm:h-10 sm:w-10 border border-white/10"
                                    >
                                        <Avatar.Image
                                            src="https://api.dicebear.com/7.x/avataaars/svg?seed={thread.author}"
                                        />
                                        <Avatar.Fallback
                                            >{thread.author
                                                .slice(0, 2)
                                                .toUpperCase()}</Avatar.Fallback
                                        >
                                    </Avatar.Root>
                                    <div class="flex-1 min-w-0 space-y-2.5">
                                        <div
                                            class="flex flex-wrap items-center gap-x-3 gap-y-1.5"
                                        >
                                            <h3
                                                class="text-base sm:text-lg font-bold group-hover:text-primary transition-colors cursor-pointer truncate"
                                            >
                                                {thread.title}
                                            </h3>
                                            {#if thread.private}
                                                <Badge
                                                    variant="outline"
                                                    class="bg-zinc-900 border-white/5 text-[9px] sm:text-[10px] text-zinc-500 font-bold uppercase tracking-wider gap-1.5"
                                                >
                                                    <svg
                                                        class="w-2.5 h-2.5"
                                                        fill="currentColor"
                                                        viewBox="0 0 24 24"
                                                        ><path
                                                            d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"
                                                        /></svg
                                                    >
                                                    Private
                                                </Badge>
                                            {/if}
                                        </div>
                                        <div
                                            class="flex flex-wrap items-center gap-x-2 gap-y-1 text-[10px] sm:text-[11px] text-zinc-500 font-medium"
                                        >
                                            <span class="text-zinc-400"
                                                >{thread.author}</span
                                            >
                                            <span>{thread.time}</span>
                                            <span
                                                class="hidden xs:inline text-zinc-700"
                                                >—</span
                                            >
                                            <span
                                                >{thread.messages} message</span
                                            >
                                            <Badge
                                                variant="outline"
                                                class="bg-zinc-900/50 border-white/5 text-zinc-400 gap-1 font-medium py-0 h-5"
                                            >
                                                <svg
                                                    class="w-3 h-3 flex-shrink-0"
                                                    fill="none"
                                                    stroke="currentColor"
                                                    viewBox="0 0 24 24"
                                                    ><path
                                                        d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
                                                    /></svg
                                                >
                                                {thread.repo}
                                            </Badge>
                                            <span
                                                class="flex items-center gap-1"
                                            >
                                                <svg
                                                    class="w-3 h-3"
                                                    fill="currentColor"
                                                    viewBox="0 0 24 24"
                                                    ><path
                                                        d="M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"
                                                    /></svg
                                                >
                                                {thread.stars}
                                            </span>
                                        </div>
                                        <div
                                            class="bg-zinc-900/30 border border-white/5 rounded-lg p-3 group-hover:bg-zinc-900/50 transition-colors"
                                        >
                                            <p
                                                class="text-xs sm:text-sm text-zinc-400 italic"
                                            >
                                                "{thread.snippet}"
                                            </p>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>
                </div>
            {:else}
                <!-- Settings View -->
                <div
                    class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8"
                >
                    <h1
                        class="text-2xl sm:text-3xl font-bold tracking-tight text-center sm:text-left"
                    >
                        Personal Settings
                    </h1>
                    <Button
                        variant="ghost"
                        size="sm"
                        onclick={handleLogout}
                        class="text-zinc-500 hover:text-red-400 gap-1.5 uppercase font-bold tracking-widest"
                    >
                        <LogOut class="w-3.5 h-3.5" />
                        Logout
                    </Button>
                </div>

                <div class="space-y-6">
                    <!-- Profile Card -->
                    <Card.Root
                        class="bg-zinc-900/40 border-white/5 overflow-hidden group hover:bg-zinc-900/60 transition-colors"
                    >
                        <Card.Content
                            class="p-5 sm:p-6 flex flex-col md:flex-row md:items-center justify-between gap-6"
                        >
                            <div
                                class="flex flex-col sm:flex-row items-center gap-6"
                            >
                                <Avatar.Root
                                    class="h-16 w-16 sm:h-14 sm:w-14 border border-white/10"
                                >
                                    <Avatar.Image
                                        src="https://api.dicebear.com/7.x/avataaars/svg?seed={profile.username}"
                                        alt="Avatar"
                                    />
                                    <Avatar.Fallback
                                        >{profile.username
                                            .slice(0, 2)
                                            .toUpperCase()}</Avatar.Fallback
                                    >
                                </Avatar.Root>
                                <div
                                    class="grid grid-cols-1 xs:grid-cols-2 md:grid-cols-3 gap-x-8 gap-y-4 text-center sm:text-left"
                                >
                                    <div>
                                        <p
                                            class="text-[9px] sm:text-[10px] text-zinc-500 uppercase font-bold tracking-widest mb-1 sm:mb-1.5"
                                        >
                                            Username
                                        </p>
                                        <p class="text-sm font-semibold">
                                            {profile.username}
                                        </p>
                                    </div>
                                    <div>
                                        <p
                                            class="text-[9px] sm:text-[10px] text-zinc-500 uppercase font-bold tracking-widest mb-1 sm:mb-1.5"
                                        >
                                            Name
                                        </p>
                                        <p class="text-sm font-semibold">
                                            {profile.first_name}
                                            {profile.last_name}
                                        </p>
                                    </div>
                                    <div class="xs:col-span-2 md:col-span-1">
                                        <p
                                            class="text-[9px] sm:text-[10px] text-zinc-500 uppercase font-bold tracking-widest mb-1 sm:mb-1.5"
                                        >
                                            Email
                                        </p>
                                        <p
                                            class="text-sm font-semibold break-all"
                                        >
                                            {profile.email}
                                        </p>
                                    </div>
                                </div>
                            </div>
                            <Button
                                variant="secondary"
                                size="sm"
                                class="w-full md:w-auto font-bold uppercase tracking-tight"
                            >
                                Edit Details
                            </Button>
                        </Card.Content>
                    </Card.Root>

                    <!-- Stats card -->
                    <Card.Root
                        class="bg-zinc-900/20 border-white/5 overflow-hidden"
                    >
                        <Card.Header class="p-0">
                            <div
                                class="p-5 sm:p-6 border-b border-white/5 flex flex-col sm:flex-row sm:items-center justify-between gap-6 bg-zinc-900/10"
                            >
                                <div
                                    class="flex items-center justify-around sm:justify-start gap-6 sm:gap-10 text-center sm:text-left"
                                >
                                    <div>
                                        <p
                                            class="text-[9px] sm:text-[10px] text-zinc-500 uppercase font-bold tracking-widest mb-1 sm:mb-1.5"
                                        >
                                            Lines of Code
                                        </p>
                                        <p
                                            class="text-sm font-semibold text-zinc-400"
                                        >
                                            —
                                        </p>
                                    </div>
                                    <div>
                                        <p
                                            class="text-[9px] sm:text-[10px] text-zinc-500 uppercase font-bold tracking-widest mb-1 sm:mb-1.5"
                                        >
                                            Threads
                                        </p>
                                        <p class="text-sm font-semibold">1</p>
                                    </div>
                                    <div>
                                        <p
                                            class="text-[9px] sm:text-[10px] text-zinc-500 uppercase font-bold tracking-widest mb-1 sm:mb-1.5"
                                        >
                                            Messages
                                        </p>
                                        <p class="text-sm font-semibold">1</p>
                                    </div>
                                </div>
                                <Button
                                    variant="secondary"
                                    size="sm"
                                    class="w-full sm:w-auto font-bold uppercase tracking-tight"
                                >
                                    View Profile
                                </Button>
                            </div>
                        </Card.Header>
                        <Card.Content class="p-8">
                            <Tabs.Root
                                value={activeTab}
                                onValueChange={(v) => (activeTab = v as any)}
                                class="w-full"
                            >
                                <div
                                    class="flex justify-between items-center mb-8"
                                >
                                    <p
                                        class="text-zinc-500 text-xs font-medium"
                                    >
                                        No usage in the last {activeTab}.
                                    </p>
                                    <Tabs.List class="bg-transparent gap-1 p-0">
                                        {#each ["7d", "14d", "30d"] as tab}
                                            <Tabs.Trigger
                                                value={tab}
                                                class="h-6 text-[10px] px-2 rounded font-bold bg-transparent data-[state=active]:bg-zinc-100 data-[state=active]:text-black text-zinc-500 hover:text-white transition-all"
                                            >
                                                {tab}
                                            </Tabs.Trigger>
                                        {/each}
                                    </Tabs.List>
                                </div>
                                <div
                                    class="h-40 sm:h-48 bg-gradient-to-br from-zinc-800/40 to-zinc-900/40 rounded-xl border border-white/5 flex items-center justify-center relative overflow-hidden group"
                                >
                                    <div
                                        class="absolute inset-0 bg-white/5 opacity-0 group-hover:opacity-100 transition-opacity"
                                    ></div>
                                    <p
                                        class="text-[10px] text-zinc-500 font-bold uppercase tracking-widest relative z-10"
                                    >
                                        Usage Chart Placeholder
                                    </p>
                                </div>
                            </Tabs.Root>
                        </Card.Content>
                    </Card.Root>

                    <!-- Access Token card -->
                    <Card.Root class="bg-zinc-900/40 border-white/5">
                        <Card.Content class="p-6 sm:p-8 space-y-6">
                            <div>
                                <h3 class="text-lg font-bold mb-1">
                                    Access Token
                                </h3>
                                <p
                                    class="text-xs text-zinc-500 font-medium leading-relaxed"
                                >
                                    Used to authenticate Counterspell extensions
                                    and the CLI (see <a
                                        href="/docs"
                                        class="text-zinc-400 hover:text-white underline underline-offset-4"
                                        >install instructions</a
                                    >)
                                </p>
                            </div>
                            <div
                                class="flex flex-col sm:flex-row items-center gap-3"
                            >
                                <Button
                                    onclick={copyToken}
                                    variant="secondary"
                                    size="sm"
                                    class="w-full sm:w-auto gap-2 font-bold uppercase tracking-tight"
                                >
                                    <Copy class="w-3.5 h-3.5" />
                                    Copy Token
                                </Button>
                                <Button
                                    variant="secondary"
                                    size="sm"
                                    class="w-full sm:w-auto gap-2 font-bold uppercase tracking-tight"
                                >
                                    <RotateCcw class="w-3.5 h-3.5" />
                                    Rotate Token
                                </Button>
                            </div>
                        </Card.Content>
                    </Card.Root>

                    <!-- Counterspell Free card -->
                    <Card.Root class="bg-zinc-900/40 border-white/5">
                        <Card.Content class="p-6 sm:p-8 space-y-6">
                            <div>
                                <h3 class="text-lg font-bold mb-1">
                                    Counterspell Free
                                </h3>
                                <p
                                    class="text-xs text-zinc-500 font-medium leading-relaxed"
                                >
                                    Ad-supported free daily allowance of
                                    Counterspell usage (<a
                                        href="/docs"
                                        class="text-zinc-400 hover:text-white underline underline-offset-4 font-semibold"
                                        >learn more</a
                                    >)
                                </p>
                            </div>

                            <div class="flex items-center gap-2 text-zinc-500">
                                <Info class="w-4 h-4 flex-shrink-0" />
                                <span class="text-xs font-semibold"
                                    >Counterspell Free is not available for your
                                    account right now.</span
                                >
                            </div>

                            <div class="space-y-4">
                                <p
                                    class="text-[10px] font-bold italic text-zinc-500 uppercase tracking-wider"
                                >
                                    Note from the Counterspell Team
                                    (2025-01-22):
                                </p>
                                <p
                                    class="text-xs text-zinc-400 leading-relaxed font-medium"
                                >
                                    We're seeing a lot of both demand and <span
                                        class="text-zinc-200">abuse</span
                                    >
                                    now with Counterspell's new
                                    <span
                                        class="text-zinc-200 underline underline-offset-4 decoration-primary/50"
                                        >$10 daily grant</span
                                    >, which lets you use our frontier
                                    <span class="text-zinc-100 font-bold"
                                        >snart</span
                                    > agent for free, up to $10/day.
                                </p>
                                <p
                                    class="text-xs text-zinc-400 leading-relaxed font-medium"
                                >
                                    We need to apply some filters to ensure
                                    you're a real person. Your account didn't
                                    pass those filters, so we haven't been able
                                    to grant you access to Counterspell's free
                                    daily grant yet. Sorry!
                                </p>
                                <Button
                                    variant="secondary"
                                    class="w-full font-bold uppercase tracking-tight"
                                >
                                    Try to convince us that you're human
                                </Button>
                                <p
                                    class="text-[10px] text-zinc-500 font-medium italic"
                                >
                                    We want to get you access to Counterspell
                                    Free soon. Sorry for the delay.
                                </p>
                            </div>
                        </Card.Content>
                    </Card.Root>

                    <!-- Billing card -->
                    <Card.Root
                        class="bg-zinc-900/20 border-white/5 overflow-hidden"
                    >
                        <Card.Header class="p-0">
                            <div
                                class="p-6 sm:p-8 border-b border-white/5 flex flex-col sm:flex-row sm:items-center justify-between gap-6 bg-zinc-900/10"
                            >
                                <h3 class="text-lg font-bold">Billing</h3>
                                <div class="flex flex-wrap gap-2">
                                    <Button
                                        variant="secondary"
                                        size="sm"
                                        class="flex-1 sm:flex-none font-bold uppercase tracking-tight"
                                    >
                                        Workspace Billing
                                    </Button>
                                    <Button
                                        variant="secondary"
                                        size="sm"
                                        class="flex-1 sm:flex-none font-bold uppercase tracking-tight"
                                    >
                                        Invoices
                                    </Button>
                                </div>
                            </div>
                        </Card.Header>
                        <Card.Content class="p-8">
                            <div
                                class="w-full h-40 sm:h-48 bg-gradient-to-br from-zinc-800/40 to-zinc-900/40 rounded-xl border border-white/5 flex items-center justify-center relative overflow-hidden group"
                            >
                                <div
                                    class="absolute inset-0 bg-white/5 opacity-0 group-hover:opacity-100 transition-opacity"
                                ></div>
                                <p
                                    class="text-[10px] text-zinc-500 font-bold uppercase tracking-widest relative z-10"
                                >
                                    Usage Chart Placeholder
                                </p>
                            </div>
                        </Card.Content>
                    </Card.Root>

                    <!-- Code Host Connections -->
                    <Card.Root class="bg-zinc-900/40 border-white/5">
                        <Card.Content class="p-6 sm:p-8 space-y-6">
                            <div>
                                <h3 class="text-lg font-bold mb-1">
                                    Code Host Connections
                                </h3>
                                <p
                                    class="text-xs text-zinc-500 font-medium leading-relaxed"
                                >
                                    Connect your accounts for repository access
                                    and agent tools.
                                </p>
                            </div>

                            <div
                                class="flex flex-col sm:flex-row sm:items-center justify-between bg-black/40 border border-white/5 p-4 rounded-lg group hover:border-white/10 transition-all gap-6"
                            >
                                <div class="flex items-center gap-4">
                                    <div
                                        class="h-10 w-10 bg-zinc-900 rounded-lg flex items-center justify-center border border-white/5 flex-shrink-0"
                                    >
                                        <svg
                                            class="w-6 h-6 text-white"
                                            viewBox="0 0 24 24"
                                            fill="currentColor"
                                        >
                                            <path
                                                d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.041-1.416-4.041-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"
                                            />
                                        </svg>
                                    </div>
                                    <div class="flex flex-col">
                                        <div
                                            class="flex items-center gap-1.5 text-sm font-bold"
                                        >
                                            GitHub.com
                                            <Info
                                                class="w-3.5 h-3.5 text-zinc-500"
                                            />
                                        </div>
                                        <p
                                            class="text-[11px] text-zinc-400 font-medium"
                                        >
                                            @{profile.username}
                                        </p>
                                        <p
                                            class="text-[10px] text-zinc-500 font-medium"
                                        >
                                            {profile.first_name}
                                            {profile.last_name}
                                        </p>
                                    </div>
                                </div>
                                <div
                                    class="flex flex-col sm:flex-row items-stretch gap-2"
                                >
                                    <Button
                                        variant="secondary"
                                        size="sm"
                                        class="font-bold uppercase tracking-tight"
                                    >
                                        Manage Permissions
                                    </Button>
                                    <Button
                                        variant="secondary"
                                        size="sm"
                                        class="font-bold uppercase tracking-tight"
                                    >
                                        Disconnect
                                    </Button>
                                </div>
                            </div>
                        </Card.Content>
                    </Card.Root>

                    <!-- Advanced Section -->
                    <Card.Root
                        class="bg-zinc-900/40 border-white/5 overflow-hidden"
                    >
                        <Collapsible.Root bind:open={isAdvancedOpen}>
                            <Collapsible.Trigger
                                class="w-full p-6 sm:p-8 flex items-center justify-between text-left hover:bg-white/5 transition-colors rounded-none"
                            >
                                <h3 class="text-lg font-bold">Advanced</h3>
                                <ChevronDown
                                    class="w-5 h-5 text-zinc-500 transition-all {isAdvancedOpen
                                        ? 'rotate-180'
                                        : ''}"
                                />
                            </Collapsible.Trigger>

                            <Collapsible.Content>
                                <Card.Content
                                    class="px-6 sm:px-8 pb-8 space-y-10"
                                >
                                    <!-- Training -->
                                    <div class="space-y-3 pt-4">
                                        <h4 class="text-sm font-bold">
                                            Training
                                        </h4>
                                        <p
                                            class="text-xs text-zinc-500 font-medium leading-relaxed"
                                        >
                                            Opt in to sharing data with
                                            Counterspell to help train models (<a
                                                href="/docs"
                                                class="text-zinc-400 hover:text-white underline underline-offset-4"
                                                >Learn more</a
                                            >)
                                        </p>
                                        <div
                                            class="flex items-start gap-3 text-zinc-500 bg-black/20 p-3 rounded-lg border border-white/5"
                                        >
                                            <Info
                                                class="w-4 h-4 flex-shrink-0 mt-0.5"
                                            />
                                            <span
                                                class="text-[11px] font-semibold leading-normal"
                                                >Training must be enabled by
                                                your workspace administrator</span
                                            >
                                        </div>
                                    </div>

                                    <!-- Profile -->
                                    <div class="space-y-3">
                                        <h4 class="text-sm font-bold">
                                            Profile
                                        </h4>
                                        <p
                                            class="text-xs text-zinc-500 font-medium"
                                        >
                                            Control what's visible on your
                                            profile
                                        </p>
                                        <div class="flex items-center gap-3">
                                            <Checkbox
                                                id="show-calendar"
                                                checked={true}
                                                class="border-white/20 data-[state=checked]:bg-primary data-[state=checked]:border-primary"
                                            />
                                            <label
                                                for="show-calendar"
                                                class="text-xs font-bold cursor-pointer"
                                                >Show Activity Calendar on
                                                Profile</label
                                            >
                                        </div>
                                    </div>

                                    <!-- Slack Integration -->
                                    <div class="space-y-3">
                                        <h4 class="text-sm font-bold">
                                            Slack Integration
                                        </h4>
                                        <p
                                            class="text-xs text-zinc-500 font-medium leading-relaxed"
                                        >
                                            Show rich previews when you share
                                            Counterspell thread links in Slack
                                        </p>
                                        <Button
                                            variant="secondary"
                                            size="sm"
                                            class="w-full sm:w-auto gap-2 font-bold uppercase tracking-tight"
                                        >
                                            <svg
                                                class="w-3.5 h-3.5"
                                                viewBox="0 0 24 24"
                                                fill="currentColor"
                                            >
                                                <path
                                                    d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zM6.313 15.165a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313zM8.834 5.042a2.528 2.528 0 0 1-2.521-2.52A2.528 2.528 0 0 1 8.834 0a2.527 2.527 0 0 1 2.521 2.522v2.52H8.834zM8.834 6.313a2.527 2.527 0 0 1 2.521 2.521 2.527 2.527 0 0 1-2.521 2.521H2.522A2.528 2.528 0 0 1 0 8.834a2.528 2.528 0 0 1 2.522-2.521h6.312zM18.958 8.834a2.528 2.528 0 0 1 2.522-2.521A2.528 2.528 0 0 1 24 8.834a2.527 2.527 0 0 1-2.522 2.521h-2.52V8.834zM17.687 8.834a2.527 2.527 0 0 1-2.521 2.521 2.527 2.527 0 0 1-2.521-2.521V2.522A2.528 2.528 0 0 1 15.166 0a2.528 2.528 0 0 1 2.521 2.522v6.312zM15.166 18.958a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521 2.527 2.527 0 0 1-2.521-2.522v-2.52h2.521zM15.166 17.687a2.527 2.527 0 0 1-2.521-2.521 2.527 2.527 0 0 1 2.521-2.521h6.312A2.528 2.528 0 0 1 24 15.166a2.528 2.528 0 0 1-2.522 2.521h-6.312z"
                                                />
                                            </svg>
                                            Connect to Slack
                                        </Button>
                                    </div>

                                    <!-- Archive Threads -->
                                    <div class="space-y-3">
                                        <h4 class="text-sm font-bold">
                                            Archive Threads
                                        </h4>
                                        <p
                                            class="text-xs text-zinc-500 font-medium leading-relaxed"
                                        >
                                            <a
                                                href="/docs"
                                                class="text-zinc-400 hover:text-white underline underline-offset-4"
                                                >Archive</a
                                            > all threads that haven't been updated
                                            in 72 hours
                                        </p>
                                        <Button
                                            variant="secondary"
                                            size="sm"
                                            class="w-full sm:w-auto font-bold uppercase tracking-tight"
                                        >
                                            Archive Old Threads
                                        </Button>
                                    </div>

                                    <Separator class="bg-white/5" />

                                    <!-- Delete Account -->
                                    <div
                                        class="space-y-4 text-center sm:text-left"
                                    >
                                        <h4
                                            class="text-sm font-bold flex items-center gap-2 justify-center sm:justify-start"
                                        >
                                            <span
                                                class="w-1.5 h-1.5 rounded-full bg-red-500"
                                            ></span>
                                            Delete Account
                                        </h4>
                                        <p
                                            class="text-xs text-zinc-500 font-medium leading-relaxed"
                                        >
                                            Contact <a
                                                href="mailto:amp-devs@ampcode.com"
                                                class="text-zinc-400 hover:text-white underline underline-offset-4 font-semibold"
                                                >amp-devs@ampcode.com</a
                                            > for help with other account operations
                                        </p>
                                        <Button
                                            variant="destructive"
                                            class="w-full sm:w-auto font-bold uppercase tracking-widest shadow-lg shadow-red-900/20"
                                        >
                                            Delete Account
                                        </Button>
                                    </div>
                                </Card.Content>
                            </Collapsible.Content>
                        </Collapsible.Root>
                    </Card.Root>
                </div>
            {/if}
        {/if}
    </main>
</div>

<style>
    :global(body) {
        background-color: black;
        cursor: default;
    }

    /* Custom scrollbar for premium feel */
    :global(::-webkit-scrollbar) {
        width: 8px;
    }
    :global(::-webkit-scrollbar-track) {
        background: transparent;
    }
    :global(::-webkit-scrollbar-thumb) {
        background: #27272a;
        border-radius: 4px;
    }
    :global(::-webkit-scrollbar-thumb:hover) {
        background: #3f3f46;
    }
</style>
