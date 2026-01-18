import "clsx";
import "../../chunks/app.svelte.js";
import "@sveltejs/kit/internal";
import "../../chunks/exports.js";
import "../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../chunks/state.svelte.js";
import { s as slot, e as bind_props } from "../../chunks/index2.js";
import { o as onDestroy } from "../../chunks/index-server.js";
import { QueryClient } from "@tanstack/query-core";
import { s as setContext } from "../../chunks/context.js";
import { f as fallback } from "../../chunks/utils2.js";
const _contextKey = "$$_queryClient";
const setQueryClientContext = (client) => {
  setContext(_contextKey, client);
};
function QueryClientProvider($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let client = fallback($$props["client"], () => new QueryClient(), true);
    setQueryClientContext(client);
    onDestroy(() => {
      client.unmount();
    });
    $$renderer2.push(`<!--[-->`);
    slot($$renderer2, $$props, "default", {});
    $$renderer2.push(`<!--]-->`);
    bind_props($$props, { client });
  });
}
function _layout($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          staleTime: 1e3 * 60 * 5,
          // 5 minutes
          refetchOnWindowFocus: false,
          retry: false
          // Don't retry on 401 auth errors
        }
      }
    });
    let { children } = $$props;
    QueryClientProvider($$renderer2, {
      client: (
        // Auth guard - handle authentication and GitHub OAuth flow
        // Dashboard requires authentication AND GitHub connection
        // User is authenticated via Supabase but needs GitHub OAuth
        // Landing page - if fully authenticated (with GitHub), redirect to dashboard
        queryClient
      ),
      children: ($$renderer3) => {
        children($$renderer3);
        $$renderer3.push(`<!---->`);
      },
      $$slots: { default: true }
    });
  });
}
export {
  _layout as default
};
