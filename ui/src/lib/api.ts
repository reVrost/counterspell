import { invalidate } from '$app/navigation';
import type {
  Workspace,
  WorkspaceSetupResponse,
  CreateWorkspaceRequest,
  TaskResponse,
  FeedData,
  UserSettings,
  GitHubRepo,
  SessionInfo,
  APIResponse,
  ConflictResponse,
  Session,
  SessionResponse,
} from '$lib/types';

// API base URL - uses proxy in dev, relative path in prod
const API_BASE = import.meta.env.DEV ? '' : '';

// Helper for JSON fetch with error handling
async function fetchAPI<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.text().catch(() => 'Unknown error');
    const errMsg = `API error: ${response.status} - ${error}`;
    throw new Error(errMsg);
  }

  return response.json();
}

// Helper for POST action that returns APIResponse (no form data)
async function postAction(path: string): Promise<APIResponse> {
  const response = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
  });

  const data = await response.json().catch(() => ({ status: 'error', message: 'Unknown error' }));

  if (!response.ok) {
    const errMsg = data.message || `API error: ${response.status}`;
    throw new Error(errMsg);
  }

  return data as APIResponse;
}

// Helper for POST with JSON body that returns APIResponse
async function postJsonWithResponse(path: string, body: object): Promise<APIResponse> {
  const response = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
    credentials: 'include',
  });

  const data = await response.json().catch(() => ({ status: 'error', message: 'Unknown error' }));

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
      return await fetchAPI<SessionInfo>('/api/v1/session');
    } catch (e) {
      return {
        authenticated: false,
        githubConnected: false,
        needsGitHubAuth: true,
      };
    }
  },

  async loginWithInvoker() {
    const returnTo = `${window.location.origin}/dashboard`;
    window.location.href = `/api/v1/auth/login?return_to=${encodeURIComponent(returnTo)}`;
  },

  async logout() {
    try {
      await fetchAPI('/api/v1/logout', { method: 'POST' });
    } catch (e) {
      console.error('Logout error (ignoring):', e);
    }
    window.location.href = '/';
  },

  async disconnect() {
    try {
      await fetchAPI('/api/v1/disconnect', { method: 'POST' });
    } catch (e) {
      console.error('Disconnect error:', e);
      throw e;
    }
    window.location.href = '/';
  },
};

// ==================== WORKSPACES ====================

export const workspacesAPI = {
  async list(): Promise<Workspace[]> {
    return fetchAPI<Workspace[]>('/api/v1/workspaces');
  },

  async getSetup(): Promise<WorkspaceSetupResponse> {
    return fetchAPI<WorkspaceSetupResponse>('/api/v1/workspaces/setup');
  },

  async create(payload: CreateWorkspaceRequest): Promise<Workspace> {
    const response = await fetchAPI<Workspace>('/api/v1/workspaces', {
      method: 'POST',
      body: JSON.stringify(payload),
    });

    invalidate('/api/v1/workspaces');

    return response;
  },
};

// ==================== GITHUB ====================

export const githubAPI = {
  async listRepos(): Promise<GitHubRepo[]> {
    return fetchAPI<GitHubRepo[]>('/api/v1/github/repos');
  },
};

// ==================== TASKS ====================

export const tasksAPI = {
  async getFeed(): Promise<FeedData> {
    return fetchAPI<FeedData>('/api/v1/tasks');
  },

  async get(id: string): Promise<TaskResponse> {
    return fetchAPI<TaskResponse>(`/api/v1/tasks/${id}`);
  },

  async getDiff(id: string): Promise<{ git_diff: string }> {
    return fetchAPI<{ git_diff: string }>(`/api/v1/tasks/${id}/diff`);
  },

  async create(
    title: string,
    intent: string,
    workspaceId: string,
    modelId: string
  ): Promise<APIResponse> {
    return postJsonWithResponse('/api/v1/tasks', {
      title: title,
      intent: intent,
      workspace_id: workspaceId,
      model_id: modelId,
    });
  },

  async chat(taskId: string, message: string, modelId?: string): Promise<APIResponse> {
    return postJsonWithResponse(`/api/v1/tasks/${taskId}/chat`, {
      task_id: taskId,
      intent: message,
      model_id: modelId,
    });
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

  async retry(taskId: string): Promise<APIResponse> {
    return postAction(`/api/v1/tasks/${taskId}/retry`);
  },
};

// ==================== SESSIONS ====================

export const sessionsAPI = {
  async list(): Promise<Session[]> {
    return fetchAPI<Session[]>('/api/v1/sessions');
  },

  async get(id: string): Promise<SessionResponse> {
    return fetchAPI<SessionResponse>(`/api/v1/sessions/${id}`);
  },

  async create(agentBackend?: string): Promise<Session> {
    return fetchAPI<Session>('/api/v1/sessions', {
      method: 'POST',
      body: JSON.stringify({ agent_backend: agentBackend || '' }),
    });
  },

  async chat(id: string, message: string, modelId?: string): Promise<APIResponse> {
    return postJsonWithResponse(`/api/v1/sessions/${id}/chat`, {
      message: message,
      model_id: modelId,
    });
  },

  async promote(id: string): Promise<{ task_id: string }> {
    return fetchAPI<{ task_id: string }>(`/api/v1/sessions/${id}/promote`, {
      method: 'POST',
    });
  },
};

// ==================== SETTINGS ====================

export const settingsAPI = {
  async get(): Promise<UserSettings> {
    return fetchAPI<UserSettings>('/api/v1/settings');
  },

  async save(settings: UserSettings): Promise<void> {
    await fetchAPI('/api/v1/settings', {
      method: 'POST',
      body: JSON.stringify({
        agent_backend: settings.agentBackend,
        openrouter_key: settings.openRouterKey || '',
        zai_key: settings.zaiKey || '',
        anthropic_key: settings.anthropicKey || '',
        openai_key: settings.openAiKey || '',
      }),
    });
  },
};

// ==================== FILES ====================

export const filesAPI = {
  async search(workspaceId: string, query: string): Promise<string[]> {
    if (!query || query.length < 2) return [];
    const params = new URLSearchParams({
      workspace_id: workspaceId,
      q: query,
    });
    return fetchAPI<string[]>(`/api/v1/files/search?${params}`);
  },
};

// ==================== TRANSCRIPTION ====================

export const transcribeAPI = {
  async transcribe(audioFile: File): Promise<string> {
    const formData = new FormData();
    formData.append('audio', audioFile);

    const response = await fetch(`${API_BASE}/api/v1/transcribe`, {
      method: 'POST',
      body: formData,
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Transcription failed: ${response.status} `);
    }

    return response.text();
  },
};
