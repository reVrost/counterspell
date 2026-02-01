import type { Task, FeedData, Todo } from '$lib/types';

// Task timers for elapsed time tracking
export const taskTimers: Record<string, number> = {};

class TaskStore {
	// Current task detail (when modal is open)
	currentTask = $state<Task | null>(null);
	todos = $state<Todo[]>([]);
	reviewCount = $state(0);

	get completedCount(): number {
		return this.todos.filter((t) => t.status === 'completed').length;
	}

	get inProgressTask(): string {
		const t = this.todos.find((t) => t.status === 'in_progress');
		return t ? t.activeForm || t.content : '';
	}

	setTodos(data: string | Todo[]) {
		try {
			this.todos = typeof data === 'string' ? JSON.parse(data) : data;
		} catch (err) {
			console.error('Todo parse error:', err);
		}
	}

	updateTask(task: Partial<Task>) {
		if (this.currentTask && task.id === this.currentTask.id) {
			this.currentTask = { ...this.currentTask, ...task };
		}
	}
}

export const taskStore = new TaskStore();
