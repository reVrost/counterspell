import type {
  Project,
  Task,
  TaskResponse,
  FeedData,
  UserSettings,
  Message,
  LogEntry,
  GitHubRepo,
  SessionInfo,
  APIResponse,
  ConflictResponse,
} from "$lib/types";

// API base URL - uses proxy in dev, relative path in prod
const API_BASE = import.meta.env.DEV ? "" : "";

// Helper for JSON fetch with error handling
async function fetchAPI<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
    credentials: "include",
  });

  if (!response.ok) {
    const error = await response.text().catch(() => "Unknown error");
    const errMsg = `API error: ${response.status} - ${error}`;
    throw new Error(errMsg);
  }

  return response.json();
}

// Helper for POST with FormData
async function postForm<T>(path: string, formData: FormData): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    body: formData,
    credentials: "include",
  });

  if (!response.ok) {
    const error = await response.text().catch(() => "Unknown error");
    const errMsg = `API error: ${response.status} - ${error}`;
    throw new Error(errMsg);
  }

  return response.json().catch(() => ({}) as T);
}

// Helper for POST without response body
async function postFormNoResponse(
  path: string,
  formData: FormData,
): Promise<void> {
  const response = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    body: formData,
    credentials: "include",
  });

  if (!response.ok) {
    const error = await response.text().catch(() => "Unknown error");
    const errMsg = `API error: ${response.status} - ${error}`;
    throw new Error(errMsg);
  }
}

// Helper for POST that returns structured APIResponse
async function postFormWithResponse(
  path: string,
  formData: FormData,
): Promise<APIResponse> {
  const response = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    body: formData,
    credentials: "include",
  });

  const data = await response
    .json()
    .catch(() => ({ status: "error", message: "Unknown error" }));

  if (!response.ok) {
    const errMsg = data.message || `API error: ${response.status}`;
    throw new Error(errMsg);
  }

  return data as APIResponse;
}

// Helper for POST action that returns APIResponse (no form data)
async function postAction(path: string): Promise<APIResponse> {
  const response = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
  });

  const data = await response
    .json()
    .catch(() => ({ status: "error", message: "Unknown error" }));

  if (!response.ok) {
    const errMsg = data.message || `API error: ${response.status}`;
    throw new Error(errMsg);
  }

  return data as APIResponse;
}

// Helper for POST with JSON body that returns APIResponse
async function postJsonWithResponse(
  path: string,
  body: object,
): Promise<APIResponse> {
  const response = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    credentials: "include",
  });

  const data = await response
    .json()
    .catch(() => ({ status: "error", message: "Unknown error" }));

  if (!response.ok) {
    const errMsg = data.message || `API error: ${response.status}`;
    throw new Error(errMsg);
  }

  return data as APIResponse;
}

// ==================== AUTH ====================

export const authAPI = {
  async checkSession(): Promise<SessionInfo> {
    try {
      return await fetchAPI<SessionInfo>("/api/v1/session");
    } catch (e) {
      return {
        authenticated: false,
        githubConnected: false,
        needsGitHubAuth: false,
      };
    }
  },

  async loginWithGitHub() {
    window.location.href =
      "/api/v1/github/authorize?redirect_url=" +
      window.location.origin +
      "/dashboard";
  },

  async connectGitHub() {
    window.location.href = "/api/v1/github/authorize";
  },

  async logout() {
    try {
      await fetchAPI("/api/v1/logout", { method: "POST" });
    } catch (e) {
      console.error("Logout error (ignoring):", e);
    }
    window.location.href = "/";
  },

  async disconnect() {
    try {
      await fetchAPI("/api/v1/disconnect", { method: "POST" });
    } catch (e) {
      console.error("Disconnect error:", e);
      throw e;
    }
    window.location.href = "/";
  },
};

// ==================== PROJECTS ====================

export const projectsAPI = {
  async list(): Promise<Project[]> {
    const feedData = await fetchAPI<{ projects: Record<string, Project> }>(
      "/api/v1/tasks",
    );
    return Object.values(feedData.projects || {});
  },

  async getMap(): Promise<Record<string, Project>> {
    const feedData = await fetchAPI<{ projects: Record<string, Project> }>(
      "/api/v1/tasks",
    );
    return feedData.projects || {};
  },
  async activate(owner: string, repo: string): Promise<void> {
    const formData = new FormData();
    formData.append("owner", owner);
    formData.append("repo", repo);
    await postFormNoResponse("/api/v1/project/activate", formData);
  },
};

// ==================== GITHUB ====================

export const githubAPI = {
  async listRepos(): Promise<GitHubRepo[]> {
    return fetchAPI<GitHubRepo[]>("/api/v1/github/repos");
  },
};

// ==================== TASKS ====================

export const tasksAPI = {
  async getFeed(): Promise<FeedData> {
    return fetchAPI<FeedData>("/api/v1/tasks");
  },

  async get(id: string): Promise<TaskResponse> {
    return fetchAPI<TaskResponse>(`/api/v1/task/${id}`);
  },

  async create(
    intent: string,
    projectId: string,
    modelId: string,
  ): Promise<APIResponse> {
    return postJsonWithResponse("/api/v1/tasks", {
      intent: intent,
      project_id: projectId,
      model_id: modelId,
    });
  },

  async chat(
    taskId: string,
    message: string,
    modelId?: string,
  ): Promise<APIResponse> {
    return postJsonWithResponse(`/api/v1/action/chat/${taskId}`, {
      message,
      model_id: modelId,
    });
  },

  async retry(taskId: string): Promise<APIResponse> {
    return postAction(`/api/v1/tasks/${taskId}/retry`);
  },

  async clear(taskId: string): Promise<APIResponse> {
    return postAction(`/api/v1/tasks/${taskId}/clear`);
  },

  async merge(taskId: string): Promise<APIResponse | ConflictResponse> {
    return postAction(`/api/v1/tasks/${taskId}/merge`);
  },

  async createPR(taskId: string): Promise<APIResponse> {
    return postAction(`/api/v1/tasks/${taskId}/pr`);
  },

  async discard(taskId: string): Promise<APIResponse> {
    return postAction(`/api/v1/tasks/${taskId}/discard`);
  },
};

// ==================== SETTINGS ====================

export const settingsAPI = {
  async get(): Promise<UserSettings> {
    return fetchAPI<UserSettings>("/api/v1/settings");
  },

  async save(settings: UserSettings): Promise<void> {
    await fetchAPI("/api/v1/settings", {
      method: "POST",
      body: JSON.stringify({
        agent_backend: settings.agentBackend,
        openrouter_key: settings.openRouterKey || "",
        zai_key: settings.zaiKey || "",
        anthropic_key: settings.anthropicKey || "",
        openai_key: settings.openAiKey || "",
      }),
    });
  },
};

// ==================== FILES ====================

export const filesAPI = {
  async search(projectId: string, query: string): Promise<string[]> {
    if (!query || query.length < 2) return [];
    const params = new URLSearchParams({
      project_id: projectId,
      q: query,
    });
    return fetchAPI<string[]>(`/api/v1/files/search?${params}`);
  },
};

// ==================== TRANSCRIPTION ====================

export const transcribeAPI = {
  async transcribe(audioFile: File): Promise<string> {
    const formData = new FormData();
    formData.append("audio", audioFile);

    const response = await fetch(`${API_BASE}/api/v1/transcribe`, {
      method: "POST",
      body: formData,
      credentials: "include",
    });

    if (!response.ok) {
      throw new Error(`Transcription failed: ${response.status} `);
    }

    return response.text();
  },
};
