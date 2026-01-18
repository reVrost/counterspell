import type { Message, LogEntry, Todo, Task } from '$lib/types';
import { taskStore } from '$lib/stores/tasks.svelte';

export interface SSECallbacks {
	onAgentUpdate?: (html: string) => void;
	onDiffUpdate?: (html: string) => void;
	onLog?: (html: string) => void;
	onStatus?: (html: string) => void;
	onTodo?: (data: string | Todo[]) => void;
	onComplete?: (status: string) => void;
	onTaskChange?: () => void;
	onError?: (error: Event) => void;
}

export function createTaskSSE(taskId: string, callbacks: SSECallbacks = {}): EventSource {
	const eventSource = new EventSource(`/api/v1/events?task_id=${taskId}`);

	eventSource.addEventListener('agent_update', (event) => {
		callbacks.onAgentUpdate?.(event.data);
	});

	eventSource.addEventListener('diff_update', (event) => {
		callbacks.onDiffUpdate?.(event.data);
	});

	eventSource.addEventListener('log', (event) => {
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

	eventSource.addEventListener('task', (event) => {
		console.log('Feed task event:', event.data);
		onUpdate();
	});

	eventSource.addEventListener('task_created', (event) => {
		console.log('Feed task_created event:', event.data);
		onUpdate();
	});

	eventSource.addEventListener('status_change', (event) => {
		console.log('Feed status_change event:', event.data);
		onUpdate();
	});

	eventSource.onerror = (error) => {
		console.error('Feed SSE Error:', error);
		onError?.(error);
	};

	return eventSource;
}
