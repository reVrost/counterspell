import { Q as head } from "../../../chunks/index2.js";
import { o as onDestroy } from "../../../chunks/index-server.js";
import "../../../chunks/app.svelte.js";
import "clsx";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    onDestroy(() => {
    });
    head("x1i5gj", $$renderer2, ($$renderer3) => {
      $$renderer3.title(($$renderer4) => {
        $$renderer4.push(`<title>Dashboard | Counterspell</title>`);
      });
    });
    {
      $$renderer2.push("<!--[-->");
      $$renderer2.push(`<div class="flex items-center justify-center h-64"><div class="flex flex-col items-center gap-3"><div class="w-8 h-8 rounded-lg bg-violet-500/10 border border-violet-500/20 flex items-center justify-center"><i class="fas fa-spinner fa-spin text-sm text-violet-400"></i></div> <p class="text-xs text-gray-500">Loading feed...</p></div></div>`);
    }
    $$renderer2.push(`<!--]-->`);
  });
}
export {
  _page as default
};
