import {
  MODELS,
  type Model,
  type Project,
  type UserSettings,
  type GitHubRepo,
  type ToastType,
} from "$lib/types";
import { authAPI, projectsAPI, settingsAPI, githubAPI } from "$lib/api";
import { pushState } from "$app/navigation";

// Reactive app state using Svelte 5 runes
class AppState {
  // UI State
  modalOpen = $state(false);
  modalTaskId = $state<string | null>(null);
  settingsOpen = $state(false);
  projectMenuOpen = $state(false);
  inputProjectMenuOpen = $state(false);
  showChatInput = $state(false);
  showNewTaskModal = $state(false);
  activeTab = $state<'inbox' | 'sessions' | 'projects' | 'focus' | 'layers'>('inbox');

  // Toast
  toastOpen = $state(false);
  toastMsg = $state("");
  toastType = $state<ToastType>("success");

  // Workspace State
  activeWorkspaceId = $state("");
  activeWorkspaceName = $state("");
  projects = $state<Project[]>([]);
  repos = $state<GitHubRepo[]>([]);

  // Model
  activeModelId = $state("");

  // Voice Recording
  isRecording = $state(false);
  isTranscribing = $state(false);
  audioLevels = $state<number[]>(Array(12).fill(0));
  recordedDuration = $state(0);

  // PWA
  deferredPrompt = $state<BeforeInstallPromptEvent | null>(null);
  canInstallPWA = $state(false);

  // Auth
  isAuthenticated = $state(false);
  userEmail = $state("");
  githubConnected = $state(false);
  githubLogin = $state("");
  needsGitHubAuth = $state(false);

  // Settings
  settings = $state<UserSettings | null>(null);

  constructor() {
    if (typeof window !== "undefined") {
      this.activeWorkspaceId =
        localStorage.getItem("counterspell_active_workspace_id") || "";
      this.activeWorkspaceName =
        localStorage.getItem("counterspell_active_workspace_name") || "";
      this.activeModelId =
        localStorage.getItem("counterspell_model") || MODELS[0].id;

      window.addEventListener("beforeinstallprompt", (e) => {
        // Prevent the mini-infobar from appearing on mobile
        e.preventDefault();
        // Stash the event so it can be triggered later.
        this.deferredPrompt = e as BeforeInstallPromptEvent;
        // Update UI notify the user they can install the PWA
        this.canInstallPWA = true;
      });

      window.addEventListener("appinstalled", () => {
        // Clear the deferredPrompt so it can be garbage collected
        this.deferredPrompt = null;
        this.canInstallPWA = false;
        console.log("PWA was installed");
      });
    }
  }

  // ==================== INITIALIZATION ====================

  async init() {
    // Load auth status
    await this.checkAuth();
    if (!this.isAuthenticated) {
      return;
    }
    // Load workspaces
    await this.loadProjects();
    // Load repos
    await this.loadRepos();
    // Load settings
    await this.loadSettings();
  }

  async checkAuth() {
    try {
      const session = await authAPI.checkSession();
      this.isAuthenticated = session.authenticated;
      this.userEmail = session.email || "";
      this.githubConnected = session.githubConnected;
      this.githubLogin = session.githubLogin || "";
      this.needsGitHubAuth = session.needsGitHubAuth;
    } catch (err) {
      console.error("Auth check failed:", err);
      this.isAuthenticated = false;
      this.githubConnected = false;
      this.needsGitHubAuth = false;
    }
  }

  async loadProjects() {
    try {
      this.projects = await projectsAPI.list();
    } catch (err) {
      console.error("Failed to load workspaces:", err);
    }
  }

  async loadRepos() {
    try {
      this.repos = await githubAPI.listRepos();
    } catch (err) {
      console.error("Failed to load repos:", err);
    }
  }

  async loadSettings() {
    try {
      this.settings = await settingsAPI.get();
    } catch (err) {
      console.error("Failed to load settings:", err);
    }
  }

  // ==================== GETTERS ====================

  get modelName(): string {
    const m = MODELS.find((m) => m.id === this.activeModelId);
    return m ? m.name.split(" ")[1] : this.activeModelId.split("#")[1];
  }

  // ==================== ACTIONS ====================

  async setActiveWorkspace(id: string, name: string) {
    // // If it's a repo ID (number as string), activate it first
    // if (id.match(/^\d+$/)) {
    // 	const repo = this.repos.find((r) => r.id.toString() === id);
    // 	if (repo) {
    // 		try {
    // 			await projectsAPI.activate(repo.owner, repo.name);
    // 			// After activation, we need to reload projects to get the actual project ID
    // 			await this.loadProjects();
    // 			const project = this.projects.find((p) => p.name === repo.full_name);
    // 			if (project) {
    // 				id = project.id;
    // 				name = project.name;
    // 			}
    // 		} catch (err) {
    // 			console.error('Failed to activate project:', err);
    // 			this.showToast('Failed to activate project', 'error');
    // 			return;
    // 		}
    // 	}
    // }

    this.activeWorkspaceId = id;
    this.activeWorkspaceName = name;
    localStorage.setItem("counterspell_active_workspace_id", id);
    localStorage.setItem("counterspell_active_workspace_name", name);
    this.inputProjectMenuOpen = false;
    this.projectMenuOpen = false;
  }

  setModel(id: string) {
    this.activeModelId = id;
    localStorage.setItem("counterspell_model", id);
  }

  showToast(msg: string, type: ToastType = "success") {
    this.toastMsg = msg;
    this.toastType = type;
    this.toastOpen = true;
    setTimeout(
      () => {
        this.toastOpen = false;
      },
      type === "error" ? 5000 : 3000,
    );
  }

  closeModal() {
    if (!this.modalOpen) return;
    this.modalOpen = false;
    this.modalTaskId = null;
    if (history.state?.modal) {
      history.back();
    }
  }

  toggleChatInput() {
    this.showChatInput = !this.showChatInput;
  }

  closeChatInput() {
    this.showChatInput = false;
  }

  toggleNewTaskModal() {
    this.showNewTaskModal = !this.showNewTaskModal;
  }

  closeNewTaskModal() {
    this.showNewTaskModal = false;
  }

  openModal(taskId: string) {
    this.modalTaskId = taskId;
    this.modalOpen = true;
    history.pushState({ modal: true }, "");
  }

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
  }

  clearState() {
    // Reset UI state
    this.modalOpen = false;
    this.modalTaskId = null;
    this.settingsOpen = false;
    this.projectMenuOpen = false;
    this.projectMenuOpen = false;
    this.inputProjectMenuOpen = false;
    this.showNewTaskModal = false;

    // Reset Workspace State
    this.activeWorkspaceId = "";
    this.activeWorkspaceName = "";
    this.projects = [];
    this.repos = [];

    // Reset Auth
    this.isAuthenticated = false;
    this.userEmail = "";
    this.githubConnected = false;
    this.githubLogin = "";
    this.needsGitHubAuth = false;

    // Reset Settings
    this.settings = null;

    // Clear local storage
    if (typeof window !== "undefined") {
      localStorage.removeItem("counterspell_active_workspace_id");
      localStorage.removeItem("counterspell_active_workspace_name");
      localStorage.removeItem("counterspell_model");
      // Clear any other app-specific keys if they exist
      sessionStorage.clear();
    }
  }

  // ==================== AUTH ACTIONS ====================

  async login() {
    await authAPI.loginWithInvoker();
  }

  async logout() {
    this.clearState();
    await authAPI.logout();
  }

  async disconnect() {
    const confirmed = confirm(
      "Are you sure you want to disconnect GitHub and DELETE all workspace data? This cannot be undone.",
    );
    if (!confirmed) return;

    this.clearState();
    try {
      await authAPI.disconnect();
    } catch (err) {
      console.error("Failed to disconnect:", err);
      this.showToast("Failed to disconnect properly", "error");
    }
  }

  // ==================== SETTINGS ACTIONS ====================

  openSettings() {
    this.settingsOpen = true;
  }

  closeSettings() {
    this.settingsOpen = false;
  }

  async saveSettings(newSettings: UserSettings) {
    try {
      await settingsAPI.save(newSettings);
      this.settings = newSettings;
      this.closeSettings();
      this.showToast("Settings saved");
    } catch (err) {
      console.error("Failed to save settings:", err);
      this.showToast("Failed to save settings", "error");
    }
  }
}

export const appState = new AppState();

// PWA event types
interface BeforeInstallPromptEvent extends Event {
  prompt(): Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
}
