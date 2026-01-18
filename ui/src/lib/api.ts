import type { Project, Task, FeedData, UserSettings, Message, LogEntry, GitHubRepo, SessionInfo } from '$lib/types';

// API base URL - uses proxy in dev, relative path in prod
const API_BASE = import.meta.env.DEV ? '' : '';

// Helper for JSON fetch with error handling
async function fetchAPI<T>(path: string, options: RequestInit = {}): Promise<T> {
	const response = await fetch(`${API_BASE}${path}`, {
		...options,
		headers: {
			'Content-Type': 'application/json',
			...options.headers
		},
		credentials: 'include'
	});

	if (!response.ok) {
		const error = await response.text().catch(() => 'Unknown error');
		throw new Error(`API error: ${response.status} - ${error}`);
	}

	return response.json();
}

// Helper for POST with FormData
async function postForm<T>(path: string, formData: FormData): Promise<T> {
	const response = await fetch(`${API_BASE}${path}`, {
		method: 'POST',
		body: formData,
		credentials: 'include'
	});

	if (!response.ok) {
		const error = await response.text().catch(() => 'Unknown error');
		throw new Error(`API error: ${response.status} - ${error}`);
	}

	return response.json().catch(() => ({} as T));
}

// Helper for POST without response body
async function postFormNoResponse(path: string, formData: FormData): Promise<void> {
	const response = await fetch(`${API_BASE}${path}`, {
		method: 'POST',
		body: formData,
		credentials: 'include'
	});

	if (!response.ok) {
		const error = await response.text().catch(() => 'Unknown error');
		throw new Error(`API error: ${response.status} - ${error}`);
	}
}

// ==================== AUTH ====================

export const authAPI = {
	async checkSession(): Promise<SessionInfo> {
		try {
			return await fetchAPI<SessionInfo>('/api/session');
		} catch (e) {
			return {
				authenticated: false,
				githubConnected: false,
				needsGitHubAuth: false
			};
		}
	},

	async loginWithGitHub() {
		window.location.href = '/auth/oauth/github';
	},

	async connectGitHub() {
		window.location.href = '/api/github/authorize';
	},

	async logout() {
		try {
			await fetchAPI('/api/logout', { method: 'POST' });
		} catch (e) {
			console.error('Logout error (ignoring):', e);
		}
		window.location.href = '/';
	},

	async disconnect() {
		try {
			await fetchAPI('/api/disconnect', { method: 'POST' });
		} catch (e) {
			console.error('Disconnect error:', e);
			throw e;
		}
		window.location.href = '/';
	}
};

// ==================== PROJECTS ====================

export const projectsAPI = {
	async list(): Promise<Project[]> {
		const feedData = await fetchAPI<{ projects: Record<string, Project> }>('/api/feed');
		return Object.values(feedData.projects || {});
	},

	async getMap(): Promise<Record<string, Project>> {
		const feedData = await fetchAPI<{ projects: Record<string, Project> }>('/api/feed');
		return feedData.projects || {};
	},
	async activate(owner: string, repo: string): Promise<void> {
		const formData = new FormData();
		formData.append('owner', owner);
		formData.append('repo', repo);
		await postFormNoResponse('/api/project/activate', formData);
	}
};

// ==================== GITHUB ====================

export const githubAPI = {
	async listRepos(): Promise<GitHubRepo[]> {
		return fetchAPI<GitHubRepo[]>('/api/github/repos');
	}
};

// ==================== TASKS ====================

export const tasksAPI = {
	async getFeed(): Promise<FeedData> {
		return fetchAPI<FeedData>('/api/feed');
	},

	async get(id: string): Promise<{ task: Task; project: Project; messages: Message[]; logs: LogEntry[] }> {
		return fetchAPI<{ task: Task; project: Project; messages: Message[]; logs: LogEntry[] }>(
			`/api/task/${id}`
		);
	},

	async create(intent: string, projectId: string, modelId: string): Promise<void> {
		const formData = new FormData();
		formData.append('voice_input', intent);
		formData.append('project_id', projectId);
		formData.append('model_id', modelId);

		await postFormNoResponse('/api/add-task', formData);
	},

	async chat(taskId: string, message: string, modelId?: string): Promise<void> {
		const formData = new FormData();
		formData.append('message', message);
		if (modelId) {
			formData.append('model_id', modelId);
		}

		await postFormNoResponse(`/api/action/chat/${taskId}`, formData);
	},

	async retry(taskId: string): Promise<void> {
		await fetchAPI(`/api/action/retry/${taskId}`, { method: 'POST' });
	},

	async clear(taskId: string): Promise<void> {
		await fetchAPI(`/api/action/clear/${taskId}`, { method: 'POST' });
	},

	async merge(taskId: string): Promise<void> {
		await fetchAPI(`/api/action/merge/${taskId}`, { method: 'POST' });
	},

	async createPR(taskId: string): Promise<void> {
		await fetchAPI(`/api/action/pr/${taskId}`, { method: 'POST' });
	},

	async discard(taskId: string): Promise<void> {
		await fetchAPI(`/api/action/discard/${taskId}`, { method: 'POST' });
	},

	async resolveConflict(taskId: string, filePath: string, choice: 'ours' | 'theirs'): Promise<void> {
		const formData = new FormData();
		formData.append('file', filePath);
		formData.append('choice', choice);

		await postFormNoResponse(`/api/action/resolve-conflict/${taskId}`, formData);
	},

	async abortMerge(taskId: string): Promise<void> {
		await fetchAPI(`/api/action/abort-merge/${taskId}`, { method: 'POST' });
	},

	async completeMerge(taskId: string): Promise<void> {
		await fetchAPI(`/api/action/complete-merge/${taskId}`, { method: 'POST' });
	}
};

// ==================== SETTINGS ====================

export const settingsAPI = {
	async get(): Promise<UserSettings> {
		return fetchAPI<UserSettings>('/api/settings');
	},

	async save(settings: UserSettings): Promise<void> {
		const formData = new FormData();
		formData.append('agent_backend', settings.agentBackend);
		if (settings.openRouterKey) {
			formData.append('openrouter_key', settings.openRouterKey);
		}
		if (settings.zaiKey) {
			formData.append('zai_key', settings.zaiKey);
		}
		if (settings.anthropicKey) {
			formData.append('anthropic_key', settings.anthropicKey);
		}
		if (settings.openAiKey) {
			formData.append('openai_key', settings.openAiKey);
		}

		await postFormNoResponse('/api/settings', formData);
	}
};

// ==================== FILES ====================

export const filesAPI = {
	async search(projectId: string, query: string): Promise<string[]> {
		if (!query || query.length < 2) return [];
		const params = new URLSearchParams({
			project_id: projectId,
			q: query
		});
		return fetchAPI<string[]>(`/api/files/search?${params}`);
	}
};

// ==================== TRANSCRIPTION ====================

export const transcribeAPI = {
	async transcribe(audioFile: File): Promise<string> {
		const formData = new FormData();
		formData.append('audio', audioFile);

		const response = await fetch(`${API_BASE}/api/transcribe`, {
			method: 'POST',
			body: formData,
			credentials: 'include'
		});

		if (!response.ok) {
			throw new Error(`Transcription failed: ${response.status} `);
		}

		return response.text();
	}
};
