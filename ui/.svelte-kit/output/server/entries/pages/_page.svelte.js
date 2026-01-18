import { Q as head } from "../../chunks/index2.js";
import "@sveltejs/kit/internal";
import "../../chunks/exports.js";
import "../../chunks/utils.js";
import "clsx";
import "@sveltejs/kit/internal/server";
import "../../chunks/state.svelte.js";
import "../../chunks/app.svelte.js";
import { G as Github } from "../../chunks/github.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    head("1uha8ag", $$renderer2, ($$renderer3) => {
      {
        $$renderer3.push("<!--[!-->");
      }
      $$renderer3.push(`<!--]-->`);
    });
    $$renderer2.push(`<div class="h-screen flex flex-col overflow-hidden bg-[#0C0E12]"><div class="absolute inset-0 overflow-hidden pointer-events-none"><div class="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-[100px] animate-pulse"></div> <div class="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-[100px] animate-pulse" style="animation-delay: 2s;"></div></div> `);
    {
      $$renderer2.push("<!--[!-->");
    }
    $$renderer2.push(`<!--]--> <div class="fixed inset-0 z-[100] bg-[#0C0E12] flex flex-col items-center justify-center text-center px-6"><div class="relative z-10 max-w-md w-full space-y-8">`);
    {
      $$renderer2.push("<!--[-->");
      $$renderer2.push(`<div class="space-y-4"><div class="w-16 h-16 border-4 border-purple-500 border-t-transparent rounded-full animate-spin mx-auto"></div> <p class="text-gray-400">Checking authentication...</p></div>`);
    }
    $$renderer2.push(`<!--]--></div> <div class="absolute bottom-8 text-center space-y-2"><p class="text-xs text-gray-600"><a href="https://github.com/revrost/counterspell" target="_blank" class="hover:text-gray-400 transition">`);
    Github($$renderer2, { class: "w-3 h-3 inline mr-1" });
    $$renderer2.push(`<!----> Open Source</a></p></div></div></div>`);
  });
}
export {
  _page as default
};
