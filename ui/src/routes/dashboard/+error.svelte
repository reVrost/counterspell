<script lang="ts">
    import { page } from "$app/state";
    import ErrorView from "$lib/components/ErrorView.svelte";

    const status = $derived(page.status);
    const error = $derived(page.error);

    console.error("Dashboard error:", error);
</script>

<ErrorView
    title={status.toString()}
    message={status === 404 ? "Not Found" : "Something went wrong"}
    description={error?.message ||
        "An unexpected error occurred while loading this dashboard page."}
    homeLink="/dashboard"
/>
