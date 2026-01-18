import "clsx";
const MODELS = [
  { id: "o#anthropic/claude-sonnet-4.5", name: "Claude Sonnet 4.5" },
  { id: "o#anthropic/claude-opus-4.5", name: "Claude Opus 4.5" },
  { id: "o#google/gemini-3-pro-preview", name: "Gemini 3 Pro Preview" },
  { id: "o#google/gemini-3-flash-preview", name: "Gemini 3 Flash Preview" },
  { id: "o#openai/gpt-5.2", name: "GPT 5.2" },
  { id: "o#openai/gpt-5.1-codex-max", name: "GPT 5.1 Codex Max" },
  { id: "zai#glm-4.7", name: "GLM 4.7" }
];
const API_BASE = "";
async function fetchAPI(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers
    },
    credentials: "include"
  });
  if (!response.ok) {
    const error = await response.text().catch(() => "Unknown error");
    throw new Error(`API error: ${response.status} - ${error}`);
  }
  return response.json();
}
async function postFormNoResponse(path, formData) {
  const response = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    body: formData,
    credentials: "include"
  });
  if (!response.ok) {
    const error = await response.text().catch(() => "Unknown error");
    throw new Error(`API error: ${response.status} - ${error}`);
  }
}
const authAPI = {
  async checkSession() {
    try {
      return await fetchAPI("/api/session");
    } catch (e) {
      return {
        authenticated: false,
        githubConnected: false,
        needsGitHubAuth: false
      };
    }
  },
  async loginWithGitHub() {
    window.location.href = "/auth/oauth/github";
  },
  async connectGitHub() {
    window.location.href = "/api/github/authorize";
  },
  async logout() {
    try {
      await fetchAPI("/api/logout", { method: "POST" });
    } catch (e) {
      console.error("Logout error (ignoring):", e);
    }
    window.location.href = "/";
  },
  async disconnect() {
    try {
      await fetchAPI("/api/disconnect", { method: "POST" });
    } catch (e) {
      console.error("Disconnect error:", e);
      throw e;
    }
    window.location.href = "/";
  }
};
const projectsAPI = {
  async list() {
    const feedData = await fetchAPI("/api/feed");
    return Object.values(feedData.projects || {});
  },
  async getMap() {
    const feedData = await fetchAPI("/api/feed");
    return feedData.projects || {};
  },
  async activate(owner, repo) {
    const formData = new FormData();
    formData.append("owner", owner);
    formData.append("repo", repo);
    await postFormNoResponse("/api/project/activate", formData);
  }
};
const githubAPI = {
  async listRepos() {
    return fetchAPI("/api/github/repos");
  }
};
const settingsAPI = {
  async get() {
    return fetchAPI("/api/settings");
  },
  async save(settings) {
    const formData = new FormData();
    formData.append("agent_backend", settings.agentBackend);
    if (settings.openRouterKey) {
      formData.append("openrouter_key", settings.openRouterKey);
    }
    if (settings.zaiKey) {
      formData.append("zai_key", settings.zaiKey);
    }
    if (settings.anthropicKey) {
      formData.append("anthropic_key", settings.anthropicKey);
    }
    if (settings.openAiKey) {
      formData.append("openai_key", settings.openAiKey);
    }
    await postFormNoResponse("/api/settings", formData);
  }
};
class AppState {
  // UI State
  modalOpen = false;
  modalTaskId = null;
  settingsOpen = false;
  projectMenuOpen = false;
  inputProjectMenuOpen = false;
  // Toast
  toastOpen = false;
  toastMsg = "";
  // Project State
  activeProjectId = "";
  activeProjectName = "";
  projects = [];
  repos = [];
  // Model
  activeModelId = "";
  // Voice Recording
  isRecording = false;
  isTranscribing = false;
  audioLevels = Array(12).fill(0);
  recordedDuration = 0;
  // PWA
  deferredPrompt = null;
  canInstallPWA = false;
  // Auth
  isAuthenticated = false;
  userEmail = "";
  githubConnected = false;
  githubLogin = "";
  needsGitHubAuth = false;
  // Settings
  settings = null;
  constructor() {
    if (typeof window !== "undefined") {
      this.activeProjectId = localStorage.getItem("counterspell_active_project_id") || "";
      this.activeProjectName = localStorage.getItem("counterspell_active_project_name") || "";
      this.activeModelId = localStorage.getItem("counterspell_model") || MODELS[0].id;
    }
  }
  // ==================== INITIALIZATION ====================
  async init() {
    await this.checkAuth();
    await this.loadProjects();
    await this.loadRepos();
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
      console.error("Failed to load projects:", err);
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
  get modelName() {
    const m = MODELS.find((m2) => m2.id === this.activeModelId);
    return m ? m.name.split(" ")[1] : this.activeModelId.split("#")[1];
  }
  // ==================== ACTIONS ====================
  async setActiveProject(id, name) {
    if (id.match(/^\d+$/)) {
      const repo = this.repos.find((r) => r.id.toString() === id);
      if (repo) {
        try {
          await projectsAPI.activate(repo.owner, repo.name);
          await this.loadProjects();
          const project = this.projects.find((p) => p.name === repo.full_name);
          if (project) {
            id = project.id;
            name = project.name;
          }
        } catch (err) {
          console.error("Failed to activate project:", err);
          this.showToast("Failed to activate project");
          return;
        }
      }
    }
    this.activeProjectId = id;
    this.activeProjectName = name;
    localStorage.setItem("counterspell_active_project_id", id);
    localStorage.setItem("counterspell_active_project_name", name);
    this.inputProjectMenuOpen = false;
    this.projectMenuOpen = false;
  }
  setModel(id) {
    this.activeModelId = id;
    localStorage.setItem("counterspell_model", id);
  }
  showToast(msg) {
    this.toastMsg = msg;
    this.toastOpen = true;
    setTimeout(
      () => {
        this.toastOpen = false;
      },
      3e3
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
  openModal(taskId) {
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
    this.modalOpen = false;
    this.modalTaskId = null;
    this.settingsOpen = false;
    this.projectMenuOpen = false;
    this.inputProjectMenuOpen = false;
    this.activeProjectId = "";
    this.activeProjectName = "";
    this.projects = [];
    this.repos = [];
    this.isAuthenticated = false;
    this.userEmail = "";
    this.githubConnected = false;
    this.githubLogin = "";
    this.needsGitHubAuth = false;
    this.settings = null;
    if (typeof window !== "undefined") {
      localStorage.removeItem("counterspell_active_project_id");
      localStorage.removeItem("counterspell_active_project_name");
      localStorage.removeItem("counterspell_model");
      sessionStorage.clear();
    }
  }
  // ==================== AUTH ACTIONS ====================
  async login() {
    await authAPI.loginWithGitHub();
  }
  async connectGitHub() {
    await authAPI.connectGitHub();
  }
  async logout() {
    this.clearState();
    await authAPI.logout();
  }
  async disconnect() {
    const confirmed = confirm("Are you sure you want to disconnect GitHub and DELETE all project data? This cannot be undone.");
    if (!confirmed) return;
    this.clearState();
    try {
      await authAPI.disconnect();
    } catch (err) {
      console.error("Failed to disconnect:", err);
      this.showToast("Failed to disconnect properly");
    }
  }
  // ==================== SETTINGS ACTIONS ====================
  openSettings() {
    this.settingsOpen = true;
  }
  closeSettings() {
    this.settingsOpen = false;
  }
  async saveSettings(newSettings) {
    try {
      await settingsAPI.save(newSettings);
      this.settings = newSettings;
      this.closeSettings();
      this.showToast("Settings saved");
    } catch (err) {
      console.error("Failed to save settings:", err);
      this.showToast("Failed to save settings");
    }
  }
}
const appState = new AppState();
export {
  appState as a
};
