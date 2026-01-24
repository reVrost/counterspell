import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export function formatDuration(secs: number): string {
	const m = Math.floor(secs / 60);
	const s = secs % 60;
	return `${m}:${s < 10 ? '0' : ''}${s}`;
}

export function getInitial(text: string | null | undefined): string {
	if (!text) return '?';
	return text[0].toUpperCase();
}

export function formatRelativeTime(timestamp: number): string {
	if (!timestamp) return '';

	const now = Date.now();
	// Handle both seconds and milliseconds
	const ts = timestamp < 1e12 ? timestamp * 1000 : timestamp;
	const diff = Math.max(0, now - ts);

	const seconds = Math.floor(diff / 1000);
	const minutes = Math.floor(seconds / 60);
	const hours = Math.floor(minutes / 60);
	const days = Math.floor(hours / 24);

	if (days > 0) return `${days}d ago`;
	if (hours > 0) return `${hours}h ago`;
	if (minutes > 0) return `${minutes}m ago`;
	if (seconds > 0) return `${seconds}s ago`;
	return 'Just now';
}
