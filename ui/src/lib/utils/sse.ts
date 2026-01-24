import type { Todo } from '$lib/types';
import { taskStore } from '$lib/stores/tasks.svelte';

export enum EventType {
  AgentRunStarted = 'agent_run_started',
  AgentRunUpdated = 'agent_run_updated',
  TaskUpdated = 'task_updated',
  TaskStarted = 'task_started',
  Log = 'log',
  AgentUpdate = 'agent_update',
  StatusChange = 'status_change',
}

export interface SSECallbacks {
  onAgentUpdate?: (html: string) => void;
  onDiffUpdate?: (html: string) => void;
  onLog?: (html: string) => void;
  onStatus?: (html: string) => void;
  onTodo?: (data: string | Todo[]) => void;
  onComplete?: (status: string) => void;
  onTaskChange?: () => void;
  onRunUpdate?: (data: string) => void;
  onError?: (error: Event) => void;
}

export function createTaskSSE(taskId: string, callbacks: SSECallbacks = {}): EventSource {
  const eventSource = new EventSource(`/api/v1/events?task_id=${taskId}`);

  eventSource.addEventListener(EventType.AgentUpdate, (event) => {
    callbacks.onAgentUpdate?.(event.data);
  });

  eventSource.addEventListener(EventType.AgentRunUpdated, (event) => {
    callbacks.onRunUpdate?.(event.data);
  });

  eventSource.addEventListener(EventType.TaskUpdated, (event) => {
    callbacks.onRunUpdate?.(event.data);
  });

  eventSource.addEventListener('diff_update', (event) => {
    callbacks.onDiffUpdate?.(event.data);
  });

  eventSource.addEventListener(EventType.Log, (event) => {
    callbacks.onLog?.(event.data);
  });

  eventSource.addEventListener('status', (event) => {
    callbacks.onStatus?.(event.data);
  });

  eventSource.addEventListener('todo', (event) => {
    // Try parsing as JSON first, fallback to string
    try {
      const data = JSON.parse(event.data);
      taskStore.setTodos(data);
      callbacks.onTodo?.(data);
    } catch {
      taskStore.setTodos(event.data);
      callbacks.onTodo?.(event.data);
    }
  });

  eventSource.addEventListener('complete', (event) => {
    try {
      const parsed = JSON.parse(event.data);
      callbacks.onComplete?.(parsed.status);
    } catch {
      callbacks.onComplete?.('');
    }
  });

  eventSource.onerror = (error) => {
    console.error('SSE Error:', error);
    callbacks.onError?.(error);
  };

  return eventSource;
}

export function createFeedSSE(onUpdate: () => void, onError?: (error: Event) => void): EventSource {
  const eventSource = new EventSource('/api/v1/events');

  eventSource.addEventListener(EventType.TaskStarted, (event) => {
    console.log('Feed task_created event:', event.data);
    onUpdate();
  });

  eventSource.addEventListener(EventType.TaskUpdated, (event) => {
    console.log('Feed task_created event:', event.data);
    onUpdate();
  });

  eventSource.addEventListener(EventType.AgentRunStarted, (event) => {
    console.log('Feed status_change event:', event.data);
    onUpdate();
  });

  eventSource.addEventListener(EventType.AgentRunUpdated, (event) => {
    console.log('Feed status_change event:', event.data);
    onUpdate();
  });

  eventSource.onerror = (error) => {
    console.error('Feed SSE Error:', error);
    onError?.(error);
  };

  return eventSource;
}
