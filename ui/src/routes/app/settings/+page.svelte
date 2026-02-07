<script lang="ts">
  import { appState } from '$lib/stores/app.svelte';
  import { cn } from '$lib/utils';
  import type { UserSettings } from '$lib/types';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import BotIcon from '@lucide/svelte/icons/bot';
  import KeyIcon from '@lucide/svelte/icons/key';
  import HeartIcon from '@lucide/svelte/icons/heart';
  import GiftIcon from '@lucide/svelte/icons/gift';
  import LogOutIcon from '@lucide/svelte/icons/log-out';
  import ZapIcon from '@lucide/svelte/icons/zap';
  import LinkIcon from '@lucide/svelte/icons/link';

  let agentBackend = $state(appState.settings?.agent_backend || 'native');
  let openRouterKey = $state(appState.settings?.openrouter_key || '');
  let zaiKey = $state(appState.settings?.zai_key || '');
  let anthropicKey = $state(appState.settings?.anthropic_key || '');
  let openAiKey = $state(appState.settings?.openai_key || '');
  let saving = $state(false);

  // Update state when settings change
  $effect(() => {
    if (appState.settings) {
      agentBackend = appState.settings.agent_backend;
      openRouterKey = appState.settings.openrouter_key || '';
      zaiKey = appState.settings.zai_key || '';
      anthropicKey = appState.settings.anthropic_key || '';
      openAiKey = appState.settings.openai_key || '';
    }
  });

  async function handleSubmit(event: Event) {
    event.preventDefault();
    saving = true;

    const newSettings: UserSettings = {
      agent_backend: agentBackend,
      openrouter_key: openRouterKey,
      zai_key: zaiKey,
      anthropic_key: anthropicKey,
      openai_key: openAiKey,
    };

    try {
      await appState.saveSettings(newSettings);
      // Optional: Show success toast
    } catch (err) {
      console.error('Failed to save settings:', err);
    } finally {
      saving = false;
    }
  }

  function handleConnectOpenAI() {
    console.log('Triggering OpenAI OAuth flow...');
    // TODO: hit backend to trigger oauth
  }
</script>

<div class="animate-in fade-in duration-300 space-y-8 max-w-2xl p-4">
  <!-- Connectors -->
  <section>
    <h2
      class="text-xs font-bold text-muted-foreground uppercase tracking-wider mb-3 px-1 flex items-center gap-2"
    >
      <ZapIcon class="w-3.5 h-3.5" /> Connectors
    </h2>
    <div class="bg-card border border-border/50 rounded-xl overflow-hidden shadow-sm">
      <div
        class="p-4 flex items-center justify-between group hover:bg-white/[0.02] transition-colors"
      >
        <div class="flex items-center gap-4">
          <div
            class="w-10 h-10 rounded-full bg-emerald-500/10 flex items-center justify-center text-emerald-500 border border-emerald-500/20"
          >
            <LinkIcon class="w-5 h-5" />
          </div>
          <div>
            <h3 class="font-medium text-foreground">OpenAI Subscription</h3>
            <p class="text-sm text-muted-foreground">Connect your account for premium models</p>
          </div>
        </div>
        <Button
          variant="outline"
          size="sm"
          onclick={handleConnectOpenAI}
          class="rounded-full px-5 border-emerald-500/20 text-emerald-400 hover:text-emerald-300 hover:bg-emerald-500/10"
        >
          Connect
        </Button>
      </div>
    </div>
  </section>

  <form onsubmit={handleSubmit} class="space-y-8">
    <!-- Agent Backend -->
    <section>
      <h3
        class="text-xs font-bold text-muted-foreground uppercase tracking-wider mb-3 px-1 flex items-center gap-2"
      >
        <BotIcon class="w-3.5 h-3.5" /> Agent Backend
      </h3>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <label
          class={cn(
            'relative flex items-center p-4 rounded-xl border cursor-pointer transition-all duration-200',
            agentBackend === 'native'
              ? 'border-violet-500 bg-violet-500/10 shadow-[0_0_15px_rgba(139,92,246,0.15)]'
              : 'border-border bg-card hover:border-violet-500/50 hover:bg-violet-500/5'
          )}
        >
          <input
            type="radio"
            name="agent_backend"
            value="native"
            bind:group={agentBackend}
            class="sr-only"
          />
          <div class="flex flex-col gap-1">
            <span class="text-base font-medium text-foreground">Counterspell</span>
            <span class="text-sm text-muted-foreground">Native Go agent</span>
          </div>
          <div
            class={cn(
              'absolute top-4 right-4 w-5 h-5 rounded-full border-2 flex items-center justify-center transition-colors',
              agentBackend === 'native' ? 'border-violet-500' : 'border-muted-foreground/30'
            )}
          >
            {#if agentBackend === 'native'}
              <div class="w-2.5 h-2.5 rounded-full bg-violet-500 shadow-sm"></div>
            {/if}
          </div>
        </label>
        <label
          class={cn(
            'relative flex items-center p-4 rounded-xl border cursor-pointer transition-all duration-200',
            agentBackend === 'claude-code'
              ? 'border-violet-500 bg-violet-500/10 shadow-[0_0_15px_rgba(139,92,246,0.15)]'
              : 'border-border bg-card hover:border-violet-500/50 hover:bg-violet-500/5'
          )}
        >
          <input
            type="radio"
            name="agent_backend"
            value="claude-code"
            bind:group={agentBackend}
            class="sr-only"
          />
          <div class="flex flex-col gap-1">
            <span class="text-base font-medium text-foreground">Claude Code</span>
            <span class="text-sm text-muted-foreground">Anthropic CLI</span>
          </div>
          <div
            class={cn(
              'absolute top-4 right-4 w-5 h-5 rounded-full border-2 flex items-center justify-center transition-colors',
              agentBackend === 'claude-code' ? 'border-violet-500' : 'border-muted-foreground/30'
            )}
          >
            {#if agentBackend === 'claude-code'}
              <div class="w-2.5 h-2.5 rounded-full bg-violet-500 shadow-sm"></div>
            {/if}
          </div>
        </label>
      </div>
      <p class="text-sm text-muted-foreground mt-3 px-1 flex items-start gap-2">
        <span>
          Counterspell uses your API keys directly. For Claude Code, ensure you have the
          <code class="text-violet-400 bg-violet-500/10 px-1 py-0.5 rounded font-mono text-xs"
            >claude</code
          >
          CLI installed locally.
        </span>
      </p>
    </section>

    <!-- API Keys -->
    <section>
      <h3
        class="text-xs font-bold text-muted-foreground uppercase tracking-wider mb-3 px-1 flex items-center gap-2"
      >
        <KeyIcon class="w-3.5 h-3.5" /> API Keys (BYOK)
      </h3>
      <div class="bg-card border border-border/50 rounded-xl p-5 space-y-5 shadow-sm">
        <div>
          <label for="openrouter-key" class="block text-sm font-medium text-muted-foreground mb-2">
            OpenRouter API Key
          </label>
          <Input
            id="openrouter-key"
            type="password"
            bind:value={openRouterKey}
            placeholder="sk-or-..."
            class="font-mono bg-background/50 focus:bg-background h-11 transition-all border-border/50 focus:border-violet-500/50"
          />
        </div>
        <div>
          <label for="zai-key" class="block text-sm font-medium text-muted-foreground mb-2"
            >Z.ai API Key</label
          >
          <Input
            id="zai-key"
            type="password"
            bind:value={zaiKey}
            placeholder="zai-..."
            class="font-mono bg-background/50 focus:bg-background h-11 transition-all border-border/50 focus:border-violet-500/50"
          />
        </div>
        <div>
          <label for="anthropic-key" class="block text-sm font-medium text-muted-foreground mb-2">
            Anthropic API Key
          </label>
          <Input
            id="anthropic-key"
            type="password"
            bind:value={anthropicKey}
            placeholder="sk-ant-..."
            class="font-mono bg-background/50 focus:bg-background h-11 transition-all border-border/50 focus:border-violet-500/50"
          />
        </div>
        <div>
          <label for="openai-key" class="block text-sm font-medium text-muted-foreground mb-2"
            >OpenAI API Key</label
          >
          <Input
            id="openai-key"
            type="password"
            bind:value={openAiKey}
            placeholder="sk-..."
            class="font-mono bg-background/50 focus:bg-background h-11 transition-all border-border/50 focus:border-violet-500/50"
          />
        </div>
      </div>
    </section>

    <!-- Save Button -->
    <div class="sticky bottom-4 z-20 flex justify-end">
      <Button
        type="submit"
        disabled={saving}
        size="lg"
        class="rounded-full px-8 shadow-xl shadow-violet-500/20 bg-violet-600 hover:bg-violet-500 text-white font-medium transition-all"
      >
        {saving ? 'Saving...' : 'Save Settings'}
      </Button>
    </div>

    <!-- Danger Zone -->
    <section class="pt-6 border-t border-border/40">
      <h3
        class="text-xs font-bold text-red-400 uppercase tracking-wider mb-3 px-1 flex items-center gap-2"
      >
        Danger Zone
      </h3>
      <div class="p-5 rounded-xl border border-red-500/20 bg-red-500/5 backdrop-blur-sm">
        <p class="text-sm text-muted-foreground mb-4 leading-relaxed">
          This will disconnect your GitHub account and <strong class="text-red-400"
            >permanently delete</strong
          > all workspaces and tasks from your local environment.
        </p>
        <Button
          type="button"
          variant="outline"
          class="w-full border-red-500/30 text-red-400 hover:bg-red-500/10 hover:text-red-300 h-10 transition-colors"
          onclick={() => appState.disconnect()}
        >
          <LogOutIcon class="w-4 h-4 mr-2" />
          Disconnect & Delete All Data
        </Button>
      </div>
    </section>

    <!-- Sponsor -->
    <section
      class="bg-gradient-to-br from-violet-900/10 to-blue-900/10 border border-violet-500/20 rounded-xl p-6 text-center"
    >
      <div
        class="w-12 h-12 bg-background rounded-full flex items-center justify-center mx-auto mb-4 border border-border shadow-lg"
      >
        <HeartIcon class="w-5 h-5 text-pink-500 animate-pulse" />
      </div>
      <h3 class="text-base font-bold text-foreground mb-2">Support Open Source</h3>
      <p class="text-sm text-muted-foreground mb-5 leading-relaxed max-w-sm mx-auto">
        Counterspell is free and open source. Your sponsorship helps keep the lights on and the
        agents coding.
      </p>
      <Button
        type="button"
        variant="secondary"
        size="sm"
        class="bg-white/10 hover:bg-white/20 border border-white/10 text-foreground"
      >
        <GiftIcon class="w-4 h-4 mr-2" /> Sponsor Project
      </Button>
    </section>

    <div class="text-center pb-8">
      <p class="text-[10px] text-muted-foreground font-mono opacity-50">
        Counterspell v2.1 (Build 8492)
      </p>
    </div>
  </form>
</div>
