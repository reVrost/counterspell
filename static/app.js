/**
 * Main Alpine.js application state
 * Registered as Alpine.data('appState') before Alpine initializes
 */
document.addEventListener('alpine:init', () => {
  Alpine.data('appState', () => ({
    // UI State
    modalOpen: false,
    activeTab: 'diff',
    projectMenuOpen: false,
    inputProjectMenuOpen: false,
    toastMsg: '',
    toastOpen: false,
    settingsOpen: false,
    userMenuOpen: false,

    // Voice Recording State
    listening: false,
    isRecording: false,
    isTranscribing: false,
    audioLevel: 0,
    audioLevels: [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0],
    recordedAudio: null,
    recordedDuration: 0,
    mediaRecorder: null,
    audioChunks: [],
    audioContext: null,
    analyser: null,
    animationFrame: null,

    // PWA State
    deferredPrompt: null,
    canInstallPWA: false,

    // Project State
    activeProjectId: localStorage.getItem('counterspell_active_project_id') || '',
    activeProjectName: localStorage.getItem('counterspell_active_project_name') || '',

    // Onboarding State
    showOnboarding: !localStorage.getItem('counterspell_v1_onboarded'),
    onboardingStep: 0,

    // Auth State (read from data attribute)
    get isAuthenticated() {
      return document.body.dataset.authenticated === 'true';
    },

    // Lifecycle
    init() {
      console.log('[APP] Alpine init, isAuthenticated:', this.isAuthenticated, 'showOnboarding:', this.showOnboarding);
      if (this.showOnboarding && this.isAuthenticated) {
        this.runPostAuthSequence();
      }
      this.connectSSE();
      this.setupPWAInstall();
    },

    // Project Management
    setActiveProject(id, name) {
      console.log('[setActiveProject] called with:', id, name);
      this.activeProjectId = id;
      this.activeProjectName = name;
      localStorage.setItem('counterspell_active_project_id', id);
      localStorage.setItem('counterspell_active_project_name', name);
      this.inputProjectMenuOpen = false;
      this.projectMenuOpen = false;
    },

    // SSE Connection
    connectSSE() {
      if (this._eventSource) {
        this._eventSource.close();
      }
      const eventSource = new EventSource('/events');
      this._eventSource = eventSource;
      eventSource.addEventListener('task', (e) => {
        const data = JSON.parse(e.data);
        if (data.type === 'status_change') {
          htmx.trigger('#reviews-container', 'refresh');
        }
      });
      eventSource.onerror = () => {
        eventSource.close();
        setTimeout(() => this.connectSSE(), 5000);
      };
    },

    // Onboarding
    startOnboarding() {
      this.onboardingStep = 1;
      setTimeout(() => {
        window.location.href = '/github/authorize?type=user';
      }, 1000);
    },

    runPostAuthSequence() {
      this.onboardingStep = 2;
      setTimeout(() => {
        this.onboardingStep = 3;
      }, 2000);
    },

    // Toast Notifications
    showToast(msg) {
      this.toastMsg = msg;
      this.toastOpen = true;
      setTimeout(() => (this.toastOpen = false), 3000);
    },

    // Legacy Voice (simulated)
    simulateVoice() {
      this.listening = true;
      setTimeout(() => {
        this.listening = false;
        this.$refs.voiceForm.requestSubmit();
      }, 1200);
    },

    // Voice Recording
    async startVoiceRecording() {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
        this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
        const source = this.audioContext.createMediaStreamSource(stream);
        this.analyser = this.audioContext.createAnalyser();
        this.analyser.fftSize = 64;
        source.connect(this.analyser);

        this.mediaRecorder = new MediaRecorder(stream);
        this.audioChunks = [];
        this.recordedDuration = 0;
        this.isRecording = true;

        const startTime = Date.now();
        const updateLevels = () => {
          if (!this.isRecording) return;
          const dataArray = new Uint8Array(this.analyser.frequencyBinCount);
          this.analyser.getByteFrequencyData(dataArray);
          const levels = [];
          const step = Math.floor(dataArray.length / 12);
          for (let i = 0; i < 12; i++) {
            const val = dataArray[i * step] || 0;
            levels.push(Math.min(100, (val / 255) * 100));
          }
          this.audioLevels = levels;
          this.recordedDuration = Math.floor((Date.now() - startTime) / 1000);
          this.animationFrame = requestAnimationFrame(updateLevels);
        };
        updateLevels();

        this.mediaRecorder.ondataavailable = (e) => {
          if (e.data.size > 0) this.audioChunks.push(e.data);
        };

        this.mediaRecorder.onstop = () => {
          const blob = new Blob(this.audioChunks, { type: 'audio/webm' });
          this.recordedAudio = {
            blob: blob,
            url: URL.createObjectURL(blob),
            duration: this.recordedDuration,
          };
          stream.getTracks().forEach((t) => t.stop());
        };

        this.mediaRecorder.start();
      } catch (err) {
        this.showToast('Microphone access denied');
        this.isRecording = false;
      }
    },

    stopVoiceRecording() {
      if (!this.isRecording) return;
      this.isRecording = false;
      if (this.animationFrame) cancelAnimationFrame(this.animationFrame);
      if (this.audioContext) this.audioContext.close();
      if (this.mediaRecorder && this.mediaRecorder.state !== 'inactive') {
        this.mediaRecorder.stop();
      }
      this.audioLevels = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0];
    },

    cancelRecording() {
      this.stopVoiceRecording();
      this.recordedAudio = null;
    },

    // Transcribe and submit voice recording
    async transcribeAndSubmit(formRef, inputRef) {
      if (!this.recordedAudio || !this.recordedAudio.blob) {
        this.showToast('No recording available');
        return;
      }

      if (!this.activeProjectId) {
        this.showToast('Select a project first');
        return;
      }

      this.isTranscribing = true;
      this.showToast('Transcribing...');

      try {
        const formData = new FormData();
        formData.append('audio', this.recordedAudio.blob, 'recording.webm');

        const response = await fetch('/transcribe', {
          method: 'POST',
          body: formData,
        });

        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(errorText || 'Transcription failed');
        }

        const transcription = await response.text();

        if (!transcription || transcription.trim() === '') {
          this.showToast('Could not understand audio');
          this.recordedAudio = null;
          this.isTranscribing = false;
          return;
        }

        // Set the transcribed text in the input and submit
        inputRef.value = transcription;
        this.recordedAudio = null;
        this.isTranscribing = false;

        // Trigger form submission via HTMX
        formRef.requestSubmit();
      } catch (err) {
        console.error('Transcription error:', err);
        this.showToast('Transcription failed: ' + err.message);
        this.isTranscribing = false;
      }
    },

    formatDuration(secs) {
      const m = Math.floor(secs / 60);
      const s = secs % 60;
      return m + ':' + (s < 10 ? '0' : '') + s;
    },

    // Modal
    closeModal() {
      this.modalOpen = false;
      setTimeout(() => {
        document.getElementById('modal-content').innerHTML = '';
      }, 300);
    },

    // PWA Installation
    setupPWAInstall() {
      window.addEventListener('beforeinstallprompt', (e) => {
        e.preventDefault();
        this.deferredPrompt = e;
        this.canInstallPWA = true;
      });
      window.addEventListener('appinstalled', () => {
        this.deferredPrompt = null;
        this.canInstallPWA = false;
        this.showToast('App installed successfully!');
      });
    },

    installPWA() {
      if (!this.deferredPrompt) return;
      this.deferredPrompt.prompt();
      this.deferredPrompt.userChoice.then((choiceResult) => {
        if (choiceResult.outcome === 'accepted') {
          this.showToast('Installing app...');
        }
        this.deferredPrompt = null;
        this.canInstallPWA = false;
      });
    },
  }));
});
