<script lang="ts">
    import { page } from "$app/state";
    import ErrorView from "$lib/components/ErrorView.svelte";

    const status = $derived(page.status);
    const error = $derived(page.error);

    $effect(() => {
        console.error("Global error:", error);
    });
</script>

<div class="min-h-screen bg-background flex flex-col">
    <div class="flex-1 flex items-center justify-center">
        <ErrorView
            title={status.toString()}
            message={status === 404 ? "Not Found" : "Something went wrong"}
            description={error?.message ||
                "An unexpected error occurred. Please try again later."}
            homeLink="/"
        />
    </div>
</div>
