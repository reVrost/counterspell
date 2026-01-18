export interface Project {
	id: string;
	name: string;
	icon: string;
	color: string;
}

export interface GitHubRepo {
	id: number;
	name: string;
	full_name: string;
	owner: string;
	description: string;
	default_branch: string;
	private: boolean;
	language: string;
	updated_at: string;
	is_favorite: boolean;
}

export interface Task {
	id: string;
	project_id: string;
	title: string;
	intent: string;
	status: TaskStatus;
	position: number;
	current_step?: string;
	created_at: string;
	updated_at: string;
	assigned_agent_id?: string;
	assigned_user_id?: string;
	// Frontend-added fields (from API enrichment)
	gitDiff?: string;
	messages?: Message[];
	logs?: LogEntry[];
}

// Task Status Flow: pending → in_progress → review → done (or failed)
export type TaskStatus = 'pending' | 'in_progress' | 'review' | 'done' | 'failed';

export interface Message {
	role: 'user' | 'assistant';
	content: MessageContent[];
}

export interface MessageContent {
	type: 'text' | 'tool_use' | 'tool_result';
	text?: string;
	toolName?: string;
	toolInput?: string;
	toolId?: string;
}

export interface LogEntry {
	type: 'info' | 'error' | 'success' | 'plan' | 'code';
	message: string;
	timestamp: string;
}

export interface Todo {
	content: string;
	status: 'pending' | 'in_progress' | 'completed';
	activeForm?: string;
}

export interface UserSettings {
	agentBackend: 'native' | 'claude-code';
	openRouterKey?: string;
	zaiKey?: string;
	anthropicKey?: string;
	openAiKey?: string;
}

export interface SessionInfo {
	authenticated: boolean;
	email?: string;
	githubConnected: boolean;
	githubLogin?: string;
	needsGitHubAuth: boolean;
}

export interface FeedData {
	active: Task[];
	reviews: Task[];
	done: Task[];
	todo: Task[];
	projects: Record<string, Project>;
}

export interface Model {
	id: string;
	name: string;
}

export const MODELS: Model[] = [
	{ id: 'o#anthropic/claude-sonnet-4.5', name: 'Claude Sonnet 4.5' },
	{ id: 'o#anthropic/claude-opus-4.5', name: 'Claude Opus 4.5' },
	{ id: 'o#google/gemini-3-pro-preview', name: 'Gemini 3 Pro Preview' },
	{ id: 'o#google/gemini-3-flash-preview', name: 'Gemini 3 Flash Preview' },
	{ id: 'o#openai/gpt-5.2', name: 'GPT 5.2' },
	{ id: 'o#openai/gpt-5.1-codex-max', name: 'GPT 5.1 Codex Max' },
	{ id: 'zai#glm-4.7', name: 'GLM 4.7' }
];

// API Response Types
export interface APIResponse {
	status: 'success' | 'error' | 'conflict';
	message?: string;
	pr_url?: string;
}

export interface ConflictFile {
	path: string;
	content: string;
}

export interface ConflictResponse {
	status: 'conflict';
	task_id: string;
	conflicts: ConflictFile[];
}

export type ToastType = 'success' | 'error' | 'info';
