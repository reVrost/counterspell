<script lang="ts">
  import { appState } from '$lib/stores/app.svelte';
  import { tasksAPI, transcribeAPI } from '$lib/api';
  import { slide, modalSlideUp, backdropFade } from '$lib/utils/transitions';
  import { cn } from '$lib/utils';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import ModelSelector from './ModelSelector.svelte';
  import XIcon from '@lucide/svelte/icons/x';
  import ArrowUpIcon from '@lucide/svelte/icons/arrow-up';
  import MicIcon from '@lucide/svelte/icons/mic';
  import LoaderIcon from '@lucide/svelte/icons/loader-2';
  import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
  import PlusIcon from '@lucide/svelte/icons/plus';
  import ImageIcon from '@lucide/svelte/icons/image';
  import AtSignIcon from '@lucide/svelte/icons/at-sign';
  import ListIcon from '@lucide/svelte/icons/list';
  import CodeIcon from '@lucide/svelte/icons/code';
  import QuoteIcon from '@lucide/svelte/icons/quote';
  import WorkspaceSelectorDropdown from './WorkspaceSelectorDropdown.svelte';

  // State
  let title = $state('');
  let description = $state('');
  let isSubmitting = $state(false);
  let workspaceMenuOpen = $state(false);
  let mediaRecorder: MediaRecorder | null = null;
  let audioChunks: Blob[] = [];
  const canSubmit = $derived(
    Boolean((title.trim() || description.trim()) && appState.activeWorkspaceId && !isSubmitting)
  );

  // Actions
  function close() {
    appState.closeNewTaskModal();
  }

  async function submit() {
    if (!title.trim() && !description.trim()) return;
    if (!appState.activeWorkspaceId) {
      appState.showToast('Create or select a workspace first', 'error');
      return;
    }

    isSubmitting = true;
    try {
      const response = await tasksAPI.create(
        title,
        description,
        appState.activeWorkspaceId,
        appState.activeModelId
      );
      appState.showToast(response.message || 'Task created', 'success');
      close();
      // Reset form
      title = '';
      description = '';
    } catch (e) {
      console.error('Failed to create task:', e);
      appState.showToast(e instanceof Error ? e.message : 'Failed to create task', 'error');
    } finally {
      isSubmitting = false;
    }
  }

  // Voice Recording
  async function toggleRecording() {
    if (appState.isRecording) {
      stopRecording();
    } else {
      await startRecording();
    }
  }

  async function startRecording() {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      mediaRecorder = new MediaRecorder(stream);
      audioChunks = [];

      mediaRecorder.ondataavailable = (event) => {
        audioChunks.push(event.data);
      };

      mediaRecorder.onstop = async () => {
        appState.isTranscribing = true;
        const audioBlob = new Blob(audioChunks, { type: 'audio/wav' });
        const file = new File([audioBlob], 'recording.wav', { type: 'audio/wav' });

        try {
          const text = await transcribeAPI.transcribe(file);
          if (text) {
            // Append to description or title depending on focus?
            // Default to description for longer text
            if (!title) {
              // If title is empty and text is short, maybe title?
              // Let's just append to description for now to be safe,
              // or appending to wherever the cursor was would be ideal but hard.
              // Simple version: append to description.
              description = description ? description + ' ' + text : text;
            } else {
              description = description ? description + ' ' + text : text;
            }
          }
        } catch (e) {
          console.error('Transcription failed:', e);
          appState.showToast('Transcription failed', 'error');
        } finally {
          appState.isTranscribing = false;
          appState.isRecording = false;

          // Stop all tracks to release microphone
          stream.getTracks().forEach((track) => track.stop());
        }
      };

      mediaRecorder.start();
      appState.isRecording = true;
      appState.showToast('Recording started... speak now', 'info');

      // Setup audio analysis for visualizer if we want
      setupAudioAnalysis(stream);
    } catch (e) {
      console.error('Failed to start recording:', e);
      appState.showToast('Could not access microphone', 'error');
    }
  }

  function stopRecording() {
    if (mediaRecorder && mediaRecorder.state !== 'inactive') {
      mediaRecorder.stop();
      // appState.isRecording is set to false in onstop callback
    }
  }

  // Audio Visualizer
  let audioContext: AudioContext;
  let analyser: AnalyserNode;
  let dataArray: Uint8Array;
  let animationFrame: number;

  function setupAudioAnalysis(stream: MediaStream) {
    audioContext = new AudioContext();
    const source = audioContext.createMediaStreamSource(stream);
    analyser = audioContext.createAnalyser();
    analyser.fftSize = 32;
    source.connect(analyser);

    const bufferLength = analyser.frequencyBinCount;
    dataArray = new Uint8Array(bufferLength);

    updateAudioLevels();
  }

  function updateAudioLevels() {
    if (!appState.isRecording) return;

    analyser.getByteFrequencyData(dataArray as Uint8Array<ArrayBuffer>);

    // Calculate average volume for visualization
    let sum = 0;
    for (let i = 0; i < dataArray.length; i++) {
      sum += dataArray[i];
    }
    const average = sum / dataArray.length;

    // Normalize to 0-100 range roughly
    const level = Math.min(100, (average / 128) * 100);

    // Update simple level on appState if needed, or just animate locally
    // Since appState has audioLevels array, let's try to populate it nicely
    // This is a simplification
    appState.audioLevels = Array.from(dataArray)
      .slice(0, 12)
      .map((v) => (v / 255) * 50);

    if (appState.isRecording) {
      appState.recordedDuration = (appState.recordedDuration || 0) + 0.1; // precise timing needed?
      animationFrame = requestAnimationFrame(updateAudioLevels);
    }
  }

  // Effect to clean up
  $effect(() => {
    return () => {
      if (mediaRecorder) stopRecording();
      if (animationFrame) cancelAnimationFrame(animationFrame);
      if (audioContext) audioContext.close();
    };
  });

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      if (canSubmit) {
        submit();
      }
    }
    if (e.key === 'Escape') {
      close();
    }
  }
</script>

<div
  class="fixed inset-0 z-50 flex items-end sm:items-center justify-center pointer-events-none"
  role="dialog"
  aria-modal="true"
>
  <!-- Backdrop -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    transition:backdropFade
    onclick={close}
    class="absolute inset-0 bg-black/80 backdrop-blur-sm pointer-events-auto"
  ></div>

  <!-- Modal content -->
  <div
    transition:modalSlideUp
    class="pointer-events-auto relative w-full h-[100dvh] bg-[#0C0E12] flex flex-col sm:h-[90vh] sm:max-w-2xl sm:rounded-2xl sm:border sm:border-white/10 shadow-2xl overflow-hidden"
  >
    <!-- Header -->
    <div class="px-4 py-3 flex items-center justify-between shrink-0">
      <button
        onclick={close}
        class="w-8 h-8 rounded-full flex items-center justify-center text-gray-500 hover:text-white hover:bg-white/10 transition"
      >
        <XIcon class="w-5 h-5" />
      </button>

      {#if appState.workspaces.length > 0}
        <WorkspaceSelectorDropdown bind:open={workspaceMenuOpen} contentClass="mt-2">
          <DropdownMenu.Trigger
            class="px-3 py-1 bg-white/5 rounded-full border border-white/5 text-sm font-medium flex items-center gap-1.5 text-zinc-100 hover:bg-white/10 transition outline-none"
          >
            {appState.activeWorkspaceName || 'Select Workspace'}
            <ChevronDownIcon class="w-3.5 h-3.5 text-zinc-400" />
          </DropdownMenu.Trigger>
        </WorkspaceSelectorDropdown>
      {:else}
        <button
          onclick={() => appState.openNewWorkspaceModal()}
          class="px-3 py-1 bg-white/5 rounded-full border border-white/5 text-sm font-medium text-zinc-100 hover:bg-white/10 transition"
        >
          Create Workspace
        </button>
      {/if}

      <div class="w-8"></div>
      <!-- Spacer -->
    </div>

    <!-- Main Form -->
    <div class="flex-1 flex flex-col px-6 pt-2 pb-6 overflow-y-auto">
      {#if appState.workspaces.length === 0}
        <div
          class="mb-6 rounded-lg border border-violet-500/10 bg-violet-500/[0.03] p-4 flex items-center justify-between"
        >
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded-lg bg-violet-500/10 flex items-center justify-center">
              <span class="text-violet-400 text-sm">+</span>
            </div>
            <div class="flex flex-col">
              <span class="text-sm font-medium text-white/90">No workspace yet</span>
              <span class="text-xs text-white/40">Create one to start working</span>
            </div>
          </div>
          <button
            onclick={() => appState.openNewWorkspaceModal()}
            class="shrink-0 rounded-md bg-violet-600/20 hover:bg-violet-600/30 px-4 py-2 text-xs font-medium text-violet-300 transition-colors"
          >
            Create workspace
          </button>
        </div>
      {/if}

      <!-- Title Input -->
      <!-- autoFocus -->
      <input
        type="text"
        bind:value={title}
        placeholder="Issue title"
        onkeydown={handleKeydown}
        class="bg-transparent border-none text-3xl font-bold text-white placeholder-gray-600 focus:ring-0 focus:outline-none p-0 w-full mb-4"
      />

      <!-- Description Input -->
      <textarea
        bind:value={description}
        placeholder="Description..."
        onkeydown={handleKeydown}
        class="flex-1 bg-transparent border-none text-lg text-gray-300 placeholder-gray-600 focus:ring-0 focus:outline-none p-0 w-full resize-none font-sans leading-relaxed min-h-[100px]"
      ></textarea>

      <!-- Recording Visualizer Overlay -->
      {#if appState.isRecording}
        <div
          class="absolute inset-x-0 bottom-24 flex items-center justify-center pointer-events-none"
          transition:slide
        >
          <div
            class="bg-red-500/10 backdrop-blur-md border border-red-500/20 rounded-full px-4 py-2 flex items-center gap-3"
          >
            <div class="w-2 h-2 rounded-full bg-red-500 animate-pulse"></div>
            <span class="text-red-400 font-mono text-xs">Recording...</span>
            <!-- Simple visualizer bars -->
            <div class="flex items-end gap-[2px] h-4">
              {#each appState.audioLevels.slice(0, 8) as level}
                <div
                  class="w-1 bg-red-500/50 rounded-full transition-all duration-75"
                  style="height: {Math.max(4, level)}px"
                ></div>
              {/each}
            </div>
          </div>
        </div>
      {/if}

      {#if appState.isTranscribing}
        <div
          class="absolute inset-x-0 bottom-24 flex items-center justify-center pointer-events-none"
          transition:slide
        >
          <div
            class="bg-violet-500/10 backdrop-blur-md border border-violet-500/20 rounded-full px-4 py-2 flex items-center gap-2"
          >
            <LoaderIcon class="w-3.5 h-3.5 text-violet-400 animate-spin" />
            <span class="text-violet-400 font-mono text-xs">Transcribing...</span>
          </div>
        </div>
      {/if}
    </div>

    <!-- Bottom Controls -->
    <div class="px-4 py-3 border-t border-white/5 bg-[#0C0E12]/50 shrink-0">
      <!-- Tags/Properties -->
      <div class="flex items-center gap-2 mb-4 overflow-x-auto no-scrollbar mask-gradient-right">
        <button
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-white/5 border border-white/5 text-xs text-gray-400 hover:text-white hover:bg-white/10 transition"
        >
          <span>Backlog</span>
        </button>
        <button
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-white/5 border border-white/5 text-xs text-gray-400 hover:text-white hover:bg-white/10 transition"
        >
          <span>Priority</span>
        </button>
        <button
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-white/5 border border-white/5 text-xs text-gray-400 hover:text-white hover:bg-white/10 transition"
        >
          <span>Assignee</span>
        </button>
        <button
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-white/5 border border-white/5 text-xs text-gray-400 hover:text-white hover:bg-white/10 transition"
        >
          <span>Label</span>
        </button>
        <button
          class="w-7 h-7 rounded-full bg-white/5 border border-white/5 flex items-center justify-center text-gray-400 hover:text-white hover:bg-white/10 transition"
        >
          <PlusIcon class="w-3.5 h-3.5" />
        </button>
      </div>

      <!-- Toolbar -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4 text-gray-500">
          <button class="hover:text-gray-300 transition">
            <ImageIcon class="w-5 h-5" />
          </button>
          <button class="hover:text-gray-300 transition">
            <AtSignIcon class="w-5 h-5" />
          </button>
          <button class="hover:text-gray-300 transition">
            <ListIcon class="w-5 h-5" />
          </button>
          <button class="hover:text-gray-300 transition">
            <CodeIcon class="w-5 h-5" />
          </button>
          <button class="hover:text-gray-300 transition">
            <QuoteIcon class="w-5 h-5" />
          </button>
        </div>

        <div class="flex items-center gap-3">
          <!-- Microphone -->
          <button
            onclick={toggleRecording}
            class={cn(
              'w-9 h-9 flex items-center justify-center rounded-full transition-all duration-200',
              appState.isRecording
                ? 'bg-red-500/20 text-red-400'
                : 'text-gray-400 hover:text-white hover:bg-white/10'
            )}
          >
            <MicIcon class={cn('w-5 h-5', appState.isRecording && 'animate-pulse')} />
          </button>

          <!-- Model Selector -->
          <ModelSelector position="right" />

          <!-- Submit -->
          <button
            onclick={submit}
            disabled={!canSubmit}
            class={cn(
              'flex items-center justify-center w-9 h-9 rounded-full transition-all duration-200 shadow-lg',
              canSubmit
                ? 'bg-violet-600 text-white hover:bg-violet-500 shadow-violet-500/20'
                : 'bg-gray-800 text-gray-500 cursor-not-allowed'
            )}
          >
            {#if isSubmitting}
              <LoaderIcon class="w-4 h-4 animate-spin" />
            {:else}
              <ArrowUpIcon class="w-5 h-5" />
            {/if}
          </button>
        </div>
      </div>
    </div>
  </div>
</div>
