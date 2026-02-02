<script lang="ts">
  import MarkdownIt from 'markdown-it';
  import { cn } from '$lib/utils';

  interface Props {
    content: string;
    class?: string;
  }

  const allowedProtocols = new Set(['http:', 'https:', 'mailto:']);
  const markdown = new MarkdownIt({
    html: false,
    linkify: true,
    breaks: true,
  });

  function isAllowedLink(url: string): boolean {
    if (!url) return false;
    const trimmed = url.trim();
    if (trimmed.startsWith('#')) return true;
    try {
      const parsed = new URL(trimmed, 'https://counterspell.io');
      return allowedProtocols.has(parsed.protocol);
    } catch {
      return false;
    }
  }

  markdown.validateLink = (url: string) => isAllowedLink(url);

  const defaultLinkRenderer =
    markdown.renderer.rules.link_open ??
    ((tokens, idx, options, env, self) => self.renderToken(tokens, idx, options));

  markdown.renderer.rules.link_open = (tokens, idx, options, env, self) => {
    const hrefIndex = tokens[idx].attrIndex('href');
    if (hrefIndex >= 0) {
      const href = tokens[idx].attrs?.[hrefIndex]?.[1] ?? '';
      if (!isAllowedLink(href)) {
        tokens[idx].attrSet('href', '#');
      }
    }
    tokens[idx].attrSet('target', '_blank');
    tokens[idx].attrSet('rel', 'noopener noreferrer');
    return defaultLinkRenderer(tokens, idx, options, env, self);
  };

  const defaultImageRenderer =
    markdown.renderer.rules.image ??
    ((tokens, idx, options, env, self) => self.renderToken(tokens, idx, options));

  markdown.renderer.rules.image = (tokens, idx, options, env, self) => {
    const srcIndex = tokens[idx].attrIndex('src');
    if (srcIndex >= 0) {
      const src = tokens[idx].attrs?.[srcIndex]?.[1] ?? '';
      if (!isAllowedLink(src)) {
        tokens[idx].attrSet('src', '');
      }
    }
    tokens[idx].attrSet('loading', 'lazy');
    tokens[idx].attrSet('decoding', 'async');
    return defaultImageRenderer(tokens, idx, options, env, self);
  };

  let { content, class: className = '' }: Props = $props();

  let rendered = $derived.by(() => markdown.render(content || ''));
</script>

<div class={cn('markdown-renderer', className)}>
  {@html rendered}
</div>

<style>
  .markdown-renderer :global(p) {
    margin: 0 0 0.75rem;
  }

  .markdown-renderer :global(p:last-child) {
    margin-bottom: 0;
  }

  .markdown-renderer :global(ul),
  .markdown-renderer :global(ol) {
    margin: 0.5rem 0 0.75rem 1.25rem;
    padding: 0;
  }

  .markdown-renderer :global(li) {
    margin: 0.25rem 0;
  }

  .markdown-renderer :global(a) {
    color: #a78bfa;
    text-decoration: underline;
    text-underline-offset: 2px;
  }

  .markdown-renderer :global(code) {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.875em;
    background: rgba(15, 23, 42, 0.6);
    border: 1px solid rgba(148, 163, 184, 0.2);
    border-radius: 0.25rem;
    padding: 0.1rem 0.35rem;
  }

  .markdown-renderer :global(pre) {
    background: #0b0f17;
    border: 1px solid rgba(148, 163, 184, 0.2);
    border-radius: 0.5rem;
    padding: 0.75rem 0.9rem;
    overflow-x: auto;
    margin: 0.75rem 0;
  }

  .markdown-renderer :global(pre code) {
    background: transparent;
    border: 0;
    padding: 0;
    font-size: 0.85em;
    color: #d1d5db;
  }

  .markdown-renderer :global(blockquote) {
    border-left: 3px solid rgba(148, 163, 184, 0.35);
    padding-left: 0.75rem;
    color: #9ca3af;
    margin: 0.75rem 0;
  }

  .markdown-renderer :global(h1),
  .markdown-renderer :global(h2),
  .markdown-renderer :global(h3),
  .markdown-renderer :global(h4) {
    margin: 1rem 0 0.5rem;
    font-weight: 600;
    color: #f3f4f6;
  }

  .markdown-renderer :global(hr) {
    border: 0;
    border-top: 1px solid rgba(148, 163, 184, 0.2);
    margin: 1rem 0;
  }

  .markdown-renderer :global(table) {
    width: 100%;
    border-collapse: collapse;
    margin: 0.75rem 0;
  }

  .markdown-renderer :global(th),
  .markdown-renderer :global(td) {
    border: 1px solid rgba(148, 163, 184, 0.2);
    padding: 0.4rem 0.6rem;
  }

  .markdown-renderer :global(th) {
    background: rgba(15, 23, 42, 0.5);
  }
</style>
