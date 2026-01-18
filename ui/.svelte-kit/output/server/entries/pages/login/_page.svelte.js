import { Q as head } from "../../../chunks/index2.js";
import "../../../chunks/app.svelte.js";
import "clsx";
import { tv } from "tailwind-variants";
import { G as Github } from "../../../chunks/github.js";
tv({
  base: "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-lg text-sm font-medium transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 active:scale-[0.98]",
  variants: {
    variant: {
      default: "bg-primary text-primary-foreground hover:bg-primary/90 shadow-lg shadow-primary/20",
      destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90 shadow-lg shadow-destructive/20",
      outline: "border border-input bg-transparent hover:bg-accent hover:text-accent-foreground",
      secondary: "bg-secondary text-secondary-foreground hover:bg-secondary/80",
      ghost: "hover:bg-accent hover:text-accent-foreground",
      link: "text-primary underline-offset-4 hover:underline",
      // Custom variants matching your theme
      white: "bg-white text-black hover:bg-gray-100 shadow-lg shadow-white/10",
      card: "bg-card border border-white/[0.08] hover:border-purple-500/30 hover:bg-card/80"
    },
    size: {
      default: "h-10 px-4 py-2",
      sm: "h-9 rounded-md px-3",
      lg: "h-12 rounded-xl px-8",
      xl: "h-14 rounded-2xl px-10 text-base",
      icon: "h-10 w-10",
      "icon-sm": "h-8 w-8",
      "icon-lg": "h-11 w-11 rounded-xl"
    }
  },
  defaultVariants: {
    variant: "default",
    size: "default"
  }
});
function _page($$renderer) {
  head("1x05zx6", $$renderer, ($$renderer2) => {
    $$renderer2.title(($$renderer3) => {
      $$renderer3.push(`<title>Counterspell - AI-Native Engineering Orchestration</title>`);
    });
  });
  $$renderer.push(`<div class="fixed inset-0 z-[100] bg-background flex flex-col items-center justify-center text-center px-6"><div class="absolute inset-0 overflow-hidden pointer-events-none"><div class="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-[100px] animate-pulse"></div> <div class="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-[100px] animate-pulse" style="animation-delay: 2s;"></div></div> <div class="relative z-10 max-w-md w-full space-y-8"><div class="space-y-4"><img src="/icon-144.png" class="w-16 h-16 rounded-2xl mx-auto border border-gray-800 shadow-lg shadow-blue-500/20" alt="Counterspell Logo"/> <h1 class="text-3xl font-bold text-white tracking-tight">Welcome to Counterspell</h1> <p class="text-gray-400 text-sm leading-relaxed">Mobile-first, hosted AI agent Kanban. <br/> Orchestrate from your pocket.</p></div> `);
  {
    $$renderer.push("<!--[-->");
    $$renderer.push(`<div><a href="/auth/oauth/github" class="w-full bg-white text-black font-bold h-12 rounded-lg hover:bg-gray-200 transition active:scale-95 flex items-center justify-center gap-2">`);
    Github($$renderer, { class: "w-5 h-5" });
    $$renderer.push(`<!----> Continue with GitHub</a> <p class="mt-4 text-[10px] text-gray-600">By continuing, you agree to the Developer Protocol v2.1</p></div>`);
  }
  $$renderer.push(`<!--]--></div> <div class="absolute bottom-8 text-center"><p class="text-xs text-gray-600"><a href="https://github.com/revrost/counterspell" target="_blank" class="hover:text-gray-400 transition flex items-center justify-center gap-1">`);
  Github($$renderer, { class: "w-3 h-3" });
  $$renderer.push(`<!----> Open Source</a></p></div></div>`);
}
export {
  _page as default
};
