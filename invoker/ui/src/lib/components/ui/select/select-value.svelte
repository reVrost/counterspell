<script lang="ts">
    import { cn } from "$lib/utils.js";

    let {
        ref = $bindable(null),
        class: className,
        children,
        ...restProps
    }: {
        ref?: any;
        class?: string;
        children?: any;
        [key: string]: any;
    } = $props();
</script>

<div
    bind:this={ref}
    data-slot="select-value"
    class={cn("inline-block", className)}
    {...restProps}
>
    {@render children?.()}
</div>
