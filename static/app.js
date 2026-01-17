// Global task timer storage (persists across SSE swaps)
window.taskTimers = window.taskTimers || {};

/**
 * Main Alpine.js application state
 * Registered as Alpine.data('appState') before Alpine initializes
 */
document.addEventListener("alpine:init", () => {
  Alpine.data("appState", () => ({
    // UI State
    modalOpen: false,
    activeTab: "diff",
    projectMenuOpen: false,
    inputProjectMenuOpen: false,
    toastMsg: "",
    toastOpen: false,
    settingsOpen: false,
    userMenuOpen: false,

    // Feed Section Collapse State
    reviewsExpanded: localStorage.getItem("feed_reviews_expanded") !== "false",
    completedExpanded:
      localStorage.getItem("feed_completed_expanded") !== "false",

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
    activeProjectId:
      localStorage.getItem("counterspell_active_project_id") || "",
    activeProjectName:
      localStorage.getItem("counterspell_active_project_name") || "",

    // Onboarding State
    showOnboarding: !localStorage.getItem("counterspell_v1_onboarded"),
    onboardingStep: 0,

    // Auth State (read from data attribute)
    get isAuthenticated() {
      return document.body.dataset.authenticated === "true";
    },

    // Lifecycle
    init() {
      console.log(
        "[APP] Alpine init, isAuthenticated:",
        this.isAuthenticated,
        "showOnboarding:",
        this.showOnboarding,
      );
      if (this.showOnboarding && this.isAuthenticated) {
        this.runPostAuthSequence();
      }
      this.setupPWAInstall();
      this.setupModalHistory();
    },

    // Feed Section Toggles
    toggleReviews() {
      this.reviewsExpanded = !this.reviewsExpanded;
      localStorage.setItem("feed_reviews_expanded", this.reviewsExpanded);
    },
    toggleCompleted() {
      this.completedExpanded = !this.completedExpanded;
      localStorage.setItem("feed_completed_expanded", this.completedExpanded);
    },

    // Project Management
    setActiveProject(id, name) {
      console.log("[setActiveProject] called with:", id, name);
      this.activeProjectId = id;
      this.activeProjectName = name;
      localStorage.setItem("counterspell_active_project_id", id);
      localStorage.setItem("counterspell_active_project_name", name);
      this.inputProjectMenuOpen = false;
      this.projectMenuOpen = false;
    },

    // Onboarding
    startOnboarding() {
      this.onboardingStep = 1;
      setTimeout(() => {
        window.location.href = "/github/authorize?type=user";
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
    async startVoiceRecording(inputRef) {
      // Store inputRef for use in onstop callback
      this._inputRef = inputRef;

      try {
        const stream = await navigator.mediaDevices.getUserMedia({
          audio: true,
        });
        this.audioContext = new (
          window.AudioContext || window.webkitAudioContext
        )();
        const source = this.audioContext.createMediaStreamSource(stream);
        this.analyser = this.audioContext.createAnalyser();
        this.analyser.fftSize = 64;
        source.connect(this.analyser);

        this.mediaRecorder = new MediaRecorder(stream);
        this.audioChunks = [];
        this.recordedDuration = 0;
        this.isRecording = true;
        this._stream = stream;

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
          const blob = new Blob(this.audioChunks, { type: "audio/webm" });
          this.recordedAudio = {
            blob: blob,
            url: URL.createObjectURL(blob),
            duration: this.recordedDuration,
          };
          stream.getTracks().forEach((t) => t.stop());
        };

        this.mediaRecorder.start();
      } catch (err) {
        this.showToast("Microphone access denied");
        this.isRecording = false;
      }
    },

    stopVoiceRecording() {
      if (!this.isRecording) return;

      this.isRecording = false;
      if (this.animationFrame) cancelAnimationFrame(this.animationFrame);
      if (this.audioContext) this.audioContext.close();

      // Set up onstop to auto-transcribe
      if (this.mediaRecorder && this.mediaRecorder.state !== "inactive") {
        const inputRef = this._inputRef; // Capture from stored ref
        console.log("setting up onstop, inputRef:", inputRef);
        this.mediaRecorder.onstop = async () => {
          console.log("mediaRecorder.onstop fired");
          const blob = new Blob(this.audioChunks, { type: "audio/webm" });
          console.log("blob size:", blob.size);
          if (this._stream) {
            this._stream.getTracks().forEach((t) => t.stop());
          }

          // Auto-transcribe to input
          if (blob.size > 0 && inputRef) {
            console.log("calling transcribeToInput");
            await this.transcribeToInput(blob, inputRef);
          } else {
            console.log(
              "skipping transcribe, blob.size:",
              blob.size,
              "inputRef:",
              inputRef,
            );
          }
        };
        this.mediaRecorder.stop();
      }
      this.audioLevels = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0];
    },

    cancelRecording() {
      if (this._stream) {
        this._stream.getTracks().forEach((t) => t.stop());
      }
      this.isRecording = false;
      if (this.animationFrame) cancelAnimationFrame(this.animationFrame);
      if (this.audioContext) this.audioContext.close();
      if (this.mediaRecorder && this.mediaRecorder.state !== "inactive") {
        this.mediaRecorder.onstop = () => {}; // Don't transcribe on cancel
        this.mediaRecorder.stop();
      }
      this.audioLevels = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0];
      this.recordedAudio = null;
    },

    // Transcribe audio blob and put text in input (no auto-submit)
    async transcribeToInput(blob, inputRef) {
      console.log("transcribeToInput called, inputRef:", inputRef);
      this.isTranscribing = true;

      try {
        const formData = new FormData();
        formData.append("audio", blob, "recording.webm");

        console.log("fetching /transcribe");
        const response = await fetch("/transcribe", {
          method: "POST",
          body: formData,
        });

        console.log("response status:", response.status);
        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(errorText || "Transcription failed");
        }

        const transcription = await response.text();
        console.log("transcription result:", transcription);

        if (!transcription || transcription.trim() === "") {
          this.showToast("Could not understand audio");
          this.isTranscribing = false;
          return;
        }

        // Set the transcribed text in the input (don't submit)
        console.log("setting inputRef.value, inputRef:", inputRef);
        if (inputRef) {
          inputRef.value = transcription;
          // Trigger Alpine to update x-model
          inputRef.dispatchEvent(new Event("input", { bubbles: true }));
          console.log("inputRef.value set to:", inputRef.value);
        }
        this.isTranscribing = false;
      } catch (err) {
        console.error("Transcription error:", err);
        this.showToast("Transcription failed: " + err.message);
        this.isTranscribing = false;
      }
    },

    formatDuration(secs) {
      const m = Math.floor(secs / 60);
      const s = secs % 60;
      return m + ":" + (s < 10 ? "0" : "") + s;
    },

    // Modal
    closeModal() {
      if (!this.modalOpen) return;
      this.modalOpen = false;
      if (history.state?.modal) {
        history.back();
      }
      setTimeout(() => {
        const el = document.getElementById("modal-content");
        if (el) el.innerHTML = "";
      }, 300);
    },

    setupModalHistory() {
      window.addEventListener("popstate", (e) => {
        if (this.modalOpen && !e.state?.modal) {
          this.modalOpen = false;
          setTimeout(() => {
            document.getElementById("modal-content").innerHTML = "";
          }, 300);
        }
      });
    },

    // PWA Installation
    setupPWAInstall() {
      window.addEventListener("beforeinstallprompt", (e) => {
        e.preventDefault();
        this.deferredPrompt = e;
        this.canInstallPWA = true;
      });
      window.addEventListener("appinstalled", () => {
        this.deferredPrompt = null;
        this.canInstallPWA = false;
        this.showToast("App installed successfully!");
      });
    },

    installPWA() {
      if (!this.deferredPrompt) return;
      this.deferredPrompt.prompt();
      this.deferredPrompt.userChoice.then((choiceResult) => {
        if (choiceResult.outcome === "accepted") {
          this.showToast("Installing app...");
        }
        this.deferredPrompt = null;
        this.canInstallPWA = false;
      });
    },
  }));
});

// NOTE: Using htmx-ext-alpine-morph extension for HTMX + Alpine integration.
