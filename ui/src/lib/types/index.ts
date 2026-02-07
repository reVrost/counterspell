export interface Workspace {
  id: string;
  name: string;
  icon: string;
  color: string;
  local_path?: string;
}

export type WorkspaceCreateMode = 'import_current' | 'new_folder';

export interface WorkspaceSetupResponse {
  cwd_path: string;
  cwd_name: string;
}

export interface CreateWorkspaceRequest {
  name: string;
  mode: WorkspaceCreateMode;
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
  workspace_id?: string;
  workspace_name?: string;
  title: string;
  intent: string;
  status: TaskStatus;
  position?: number;
  failed_reason?: string;
  last_assistant_message?: string;
  created_at: number;
  updated_at: number;
  gitDiff?: string;
  git_diff?: string;
}

// Task Status Flow: draft → planning → in_progress → review → done (or failed)
export type TaskStatus = 'draft' | 'planning' | 'in_progress' | 'review' | 'done' | 'failed';

export interface Message {
  id: string;
  task_id: string;
  run_id?: string | null;
  role: 'user' | 'assistant' | 'system' | 'tool' | 'tool_result';
  parts: string; // JSON string
  model?: string | null;
  provider?: string | null;
  content: string;
  tool_id?: string | null;
  created_at: number;
  updated_at: number;
  finished_at?: number | null;
}

export interface ContentBlock {
  type: 'text' | 'thinking' | 'tool_use' | 'tool_result';
  text?: string;
  name?: string;
  input?: Record<string, unknown>;
  id?: string;
  tool_use_id?: string;
  content?: string;
}

export interface AgentStreamEvent {
  type: string;
  message_id?: string;
  role?: string;
  block_type?: string;
  delta?: string;
  block?: ContentBlock;
  session_id?: string;
  todos?: Todo[];
  error?: string;
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
  workspace?: Workspace;
  messages: Message[];
  artifacts: Artifact[];
  agent_runs?: AgentRun[];
  git_diff?: string;
  logs?: LogEntry[];
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
  agent_backend: 'native' | 'claude-code' | 'codex';
  openrouter_key?: string;
  zai_key?: string;
  anthropic_key?: string;
  openai_key?: string;
}

export interface OpenAIConnectorStatus {
  connected: boolean;
  account_id?: string;
  expires_at?: number;
  connected_at?: number;
  token_expired: boolean;
  needs_reconnect: boolean;
}

export interface Session {
  id: string;
  agent_backend: string;
  external_id?: string | null;
  backend_session_id?: string | null;
  title?: string | null;
  message_count: number;
  last_message_at?: number | null;
  created_at: number;
  updated_at: number;
}

export interface SessionMessage {
  id: string;
  session_id: string;
  sequence: number;
  role: string;
  kind: string;
  content?: string | null;
  tool_name?: string | null;
  tool_call_id?: string | null;
  raw_json: string;
  created_at: number;
}

export interface SessionResponse {
  session: Session;
  messages: SessionMessage[];
}

export interface SessionInfo {
  authenticated: boolean;
  email?: string;
  githubConnected: boolean;
  githubLogin?: string;
  needsGitHubAuth: boolean;
  authErrorCode?: string;
  authErrorMessage?: string;
}

export interface FeedData {
  active: Task[];
  reviews: Task[];
  done: Task[];
  todo: Task[];
  planning: Task[];
  projects?: Record<string, Project>;
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
  { id: 'o#openai/gpt-5.2-codex-xhigh', name: 'GPT 5.2 Codex XHigh' },
  { id: 'o#openai/gpt-5.2-codex-high', name: 'GPT 5.2 Codex High' },
  { id: 'o#openai/gpt-5.2-codex-medium', name: 'GPT 5.2 Codex Medium' },
  { id: 'o#openai/gpt-5.1-codex-max', name: 'GPT 5.1 Codex Max' },
  { id: 'zai#glm-4.7', name: 'GLM 4.7' },
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
