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
	repository_id?: string | null;
	repository_name?: string | null;
	title: string;
	intent: string;
	status: TaskStatus;
	position?: number | null;
	created_at: number;
	updated_at: number;
}

// Task Status Flow: pending → planning → in_progress → review → done (or failed)
export type TaskStatus = 'pending' | 'planning' | 'in_progress' | 'review' | 'done' | 'failed';

export interface Message {
	id: string;
	task_id: string;
	run_id?: string | null;
	role: 'user' | 'assistant' | 'system';
	parts: string; // JSON string
	model?: string | null;
	provider?: string | null;
	content: string;
	tool_id?: string | null;
	created_at: number;
	updated_at: number;
	finished_at?: number | null;
}

export interface AgentRun {
	id: string;
	task_id: string;
	prompt: string;
	agent_backend: string;
	summary_message_id?: string | null;
	cost: number;
	message_count: number;
	prompt_tokens: number;
	completion_tokens: number;
	completed_at?: number | null;
	created_at: number;
	updated_at: number;
	messages?: Message[];
	artifacts?: Artifact[];
}

export interface Artifact {
	id: string;
	run_id: string;
	path: string;
	content: string;
	version: number;
	created_at: number;
	updated_at: number;
}

export interface TaskResponse {
	task: Task;
	messages: Message[];
	artifacts: Artifact[];
	agent_runs?: AgentRun[];
	git_diff?: string;
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
	planning: Task[];
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
