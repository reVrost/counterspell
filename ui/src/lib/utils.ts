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
