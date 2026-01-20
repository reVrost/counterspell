<script lang="ts">
	import { appState } from '$lib/stores/app.svelte';
	import { cn } from '$lib/utils';
	import { modalSlideUp, backdropFade, DURATIONS } from '$lib/utils/transitions';
	import type { UserSettings } from '$lib/types';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import XIcon from '@lucide/svelte/icons/x';
	import BotIcon from '@lucide/svelte/icons/bot';
	import KeyIcon from '@lucide/svelte/icons/key';
	import HeartIcon from '@lucide/svelte/icons/heart';
	import GiftIcon from '@lucide/svelte/icons/gift';
	import LogOutIcon from '@lucide/svelte/icons/log-out';

	let agentBackend = $state(appState.settings?.agentBackend || 'native');
	let openRouterKey = $state(appState.settings?.openRouterKey || '');
	let zaiKey = $state(appState.settings?.zaiKey || '');
	let anthropicKey = $state(appState.settings?.anthropicKey || '');
	let openAiKey = $state(appState.settings?.openAiKey || '');
	let saving = $state(false);

	// Update state when settings change
	$effect(() => {
		if (appState.settings) {
			agentBackend = appState.settings.agentBackend;
			openRouterKey = appState.settings.openRouterKey || '';
			zaiKey = appState.settings.zaiKey || '';
			anthropicKey = appState.settings.anthropicKey || '';
			openAiKey = appState.settings.openAiKey || '';
		}
	});

	async function handleSubmit(event: Event) {
		event.preventDefault();
		saving = true;

		const newSettings: UserSettings = {
			agentBackend,
			openRouterKey,
			zaiKey,
			anthropicKey,
			openAiKey
		};

		try {
			await appState.saveSettings(newSettings);
		} catch (err) {
			console.error('Failed to save settings:', err);
		} finally {
			saving = false;
		}
	}
</script>

{#if appState.settingsOpen}
	<div
		transition:backdropFade|global={{ duration: DURATIONS.normal }}
		class="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4"
		role="presentation"
		aria-hidden="true"
		onclick={(e) => e.target === e.currentTarget && appState.closeSettings()}
		onkeydown={(e) => e.key === 'Escape' && appState.closeSettings()}
		aria-label="Close settings"
	>
		<div
			transition:modalSlideUp|global={{ duration: DURATIONS.normal }}
			class="bg-popover border border-gray-700 w-full max-w-lg rounded-2xl shadow-2xl overflow-hidden max-h-[90vh] overflow-y-auto"
			role="dialog"
			aria-modal="true"
			tabindex="-1"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && appState.closeSettings()}
		>
			<div
				class="px-6 py-4 border-b border-gray-800 flex justify-between items-center sticky top-0 bg-popover z-10"
			>
				<h2 class="text-lg font-bold text-white">Settings</h2>
				<button
					onclick={() => appState.closeSettings()}
					class="text-gray-500 hover:text-white transition"
				>
					<XIcon class="w-5 h-5" />
				</button>
			</div>

			<form onsubmit={handleSubmit} class="p-6 space-y-8">
				<!-- Agent Backend -->
				<div>
					<h3 class="text-xs font-bold text-gray-500 uppercase tracking-wider mb-4 flex items-center gap-2">
						<BotIcon class="w-4 h-4" /> Agent Backend
					</h3>
					<div class="grid grid-cols-2 gap-3">
						<label
							class={cn(
								'relative flex items-center p-3 rounded-lg border cursor-pointer transition-all',
								agentBackend === 'native'
									? 'border-purple-500 bg-purple-500/10'
									: 'border-gray-700 bg-gray-900 hover:border-gray-600'
							)}
						>
							<input
								type="radio"
								name="agent_backend"
								value="native"
								bind:group={agentBackend}
								class="sr-only"
							/>
							<div class="flex flex-col">
								<span class="text-sm font-medium text-white">Counterspell</span>
								<span class="text-xs text-gray-500">Native Go agent</span>
							</div>
							<div
								class={cn(
									'absolute top-2 right-2 w-4 h-4 rounded-full border-2 flex items-center justify-center',
									agentBackend === 'native' ? 'border-purple-500' : 'border-gray-600'
								)}
							>
								{#if agentBackend === 'native'}
									<div class="w-2 h-2 rounded-full bg-purple-500"></div>
								{/if}
							</div>
						</label>
						<label
							class={cn(
								'relative flex items-center p-3 rounded-lg border cursor-pointer transition-all',
								agentBackend === 'claude-code'
									? 'border-purple-500 bg-purple-500/10'
									: 'border-gray-700 bg-gray-900 hover:border-gray-600'
							)}
						>
							<input
								type="radio"
								name="agent_backend"
								value="claude-code"
								bind:group={agentBackend}
								class="sr-only"
							/>
							<div class="flex flex-col">
								<span class="text-sm font-medium text-white">Claude Code</span>
								<span class="text-xs text-gray-500">Anthropic CLI</span>
							</div>
							<div
								class={cn(
									'absolute top-2 right-2 w-4 h-4 rounded-full border-2 flex items-center justify-center',
									agentBackend === 'claude-code' ? 'border-purple-500' : 'border-gray-600'
								)}
							>
								{#if agentBackend === 'claude-code'}
									<div class="w-2 h-2 rounded-full bg-purple-500"></div>
								{/if}
							</div>
						</label>
					</div>
					<p class="text-xs text-gray-600 mt-2">
						<i class="fas fa-info-circle mr-1"></i>
						Counterspell uses your API keys. Claude Code requires the
						<code class="text-purple-400">claude</code> CLI installed.
					</p>
				</div>

				<!-- API Keys -->
				<div>
					<h3 class="text-xs font-bold text-gray-500 uppercase tracking-wider mb-4 flex items-center gap-2">
						<KeyIcon class="w-4 h-4" /> BYOK (Bring Your Own Keys)
					</h3>
					<div class="space-y-4">
						<div>
							<label for="openrouter-key" class="block text-xs font-medium text-gray-400 mb-1.5">
								OpenRouter API Key
							</label>
							<Input
								id="openrouter-key"
								type="password"
								bind:value={openRouterKey}
								placeholder="sk-or-..."
								class="font-mono"
							/>
						</div>
						<div>
							<label for="zai-key" class="block text-xs font-medium text-gray-400 mb-1.5">Z.ai API Key</label>
							<Input
								id="zai-key"
								type="password"
								bind:value={zaiKey}
								placeholder="zai-..."
								class="font-mono"
							/>
						</div>
						<div>
							<label for="anthropic-key" class="block text-xs font-medium text-gray-400 mb-1.5">
								Anthropic API Key
							</label>
							<Input
								id="anthropic-key"
								type="password"
								bind:value={anthropicKey}
								placeholder="sk-ant-..."
								class="font-mono"
							/>
						</div>
						<div>
							<label for="openai-key" class="block text-xs font-medium text-gray-400 mb-1.5">OpenAI API Key</label>
							<Input
								id="openai-key"
								type="password"
								bind:value={openAiKey}
								placeholder="sk-..."
								class="font-mono"
							/>
						</div>
					</div>
				</div>

				<!-- Save Button -->
				<div class="flex justify-end">
					<Button type="submit" disabled={saving}>
						{saving ? 'Saving...' : 'Save Changes'}
					</Button>
				</div>

				<!-- Danger Zone -->
				<div class="pt-6 border-t border-gray-800">
					<h3 class="text-xs font-bold text-red-500 uppercase tracking-wider mb-4 flex items-center gap-2">
						Danger Zone
					</h3>
					<div class="p-4 rounded-xl border border-red-500/20 bg-red-500/5">
						<p class="text-xs text-gray-400 mb-4 leading-relaxed">
							This will disconnect your GitHub account and <strong>permanently delete</strong> all repositories and tasks from the server.
						</p>
						<Button
							type="button"
							variant="outline"
							size="sm"
							class="w-full border-red-500/30 text-red-400 hover:bg-red-500/10 hover:text-red-300"
							onclick={() => appState.disconnect()}
						>
							<LogOutIcon class="w-3.5 h-3.5 mr-2" />
							Disconnect & Delete All Data
						</Button>
					</div>
				</div>

				<!-- Sponsor -->
				<div
					class="bg-gradient-to-br from-purple-900/20 to-blue-900/20 border border-purple-500/30 rounded-xl p-5 text-center mt-8"
				>
					<div
						class="w-12 h-12 bg-gray-800 rounded-full flex items-center justify-center mx-auto mb-3 border border-gray-700 shadow-xl"
					>
						<HeartIcon class="w-5 h-5 text-pink-500 animate-pulse" />
					</div>
					<h3 class="text-sm font-bold text-white mb-1">Support Open Source</h3>
					<p class="text-xs text-gray-400 mb-4 leading-relaxed">
						Counterspell is free and open source. Your sponsorship helps keep the lights on and
						the agents coding.
					</p>
					<Button type="button" variant="white" size="sm">
						<GiftIcon class="w-3 h-3 mr-1" /> Sponsor Project
					</Button>
				</div>

				<div class="text-center">
					<p class="text-[10px] text-gray-600 font-mono">Counterspell v2.1 (Build 8492)</p>
				</div>
			</form>
		</div>
	</div>
{/if}
