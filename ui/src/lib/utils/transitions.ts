/**
 * Svelte 5 transition functions following Vercel Web Interface Guidelines
 */

import type { TransitionConfig } from 'svelte/transition';
import { prefersReducedMotion } from './animations';

// Duration constants
export const DURATIONS = {
	fast: 150,
	quick: 200,
	normal: 300,
	slow: 400,
	deliberate: 500
} as const;

// Easing functions
export const EASINGS = {
	size: 'cubic-bezier(0.16, 1, 0.3, 1)',
	position: 'cubic-bezier(0.2, 0, 0, 1)',
	appear: 'cubic-bezier(0.16, 1, 0.3, 1)',
	ui: 'cubic-bezier(0.4, 0, 0.2, 1)'
} as const;

/**
 * Fade transition - simple opacity change
 */
export function fade(node: Element, { duration = 200, delay = 0 }: { duration?: number; delay?: number } = {}): TransitionConfig {
	if (prefersReducedMotion()) {
		return {
			duration: 0,
			tick: () => {}
		};
	}

	return {
		duration,
		delay,
		css: (t) => `
			opacity: ${t};
		`
	};
}

/**
 * Slide transition - combine translate and fade
 */
export function slide(
	node: Element,
	{ direction = 'up', duration = 300, delay = 0 }: { direction?: 'up' | 'down' | 'left' | 'right'; duration?: number; delay?: number } = {}
): TransitionConfig {
	if (prefersReducedMotion()) {
		return {
			duration: 0,
			tick: () => {}
		};
	}

	const transforms = {
		up: 'translateY(0)',
		down: 'translateY(0)',
		left: 'translateX(0)',
		right: 'translateX(0)'
	};

	const startTransforms = {
		up: 'translateY(32px)',
		down: 'translateY(-32px)',
		left: 'translateX(32px)',
		right: 'translateX(-32px)'
	};

	return {
		duration,
		delay,
		easing: EASINGS.position,
		css: (t) => `
			transform: ${t === 0 ? startTransforms[direction] : transforms[direction]};
			opacity: ${t};
		`
	};
}

/**
 * Scale transition - combine scale and fade
 */
export function scale(node: Element, { duration = 200, delay = 0, startScale = 0.95 }: { duration?: number; delay?: number; startScale?: number } = {}): TransitionConfig {
	if (prefersReducedMotion()) {
		return {
			duration: 0,
			tick: () => {}
		};
	}

	return {
		duration,
		delay,
		easing: EASINGS.size,
		css: (t) => `
			transform: scale(${startScale + (1 - startScale) * t});
			opacity: ${t};
		`
	};
}

/**
 * Modal slide-up transition for dialogs
 */
export function modalSlideUp(node: Element, { duration = 300, delay = 0 }: { duration?: number; delay?: number } = {}): TransitionConfig {
	if (prefersReducedMotion()) {
		return {
			duration: 0,
			tick: () => {}
		};
	}

	return {
		duration,
		delay,
		easing: EASINGS.size,
		css: (t) => `
			transform: translateY(${32 * (1 - t)}px) scale(${0.95 + 0.05 * t});
			opacity: ${t};
		`
	};
}

/**
 * Backdrop fade transition
 */
export function backdropFade(node: Element, { duration = 300, delay = 0 }: { duration?: number; delay?: number } = {}): TransitionConfig {
	if (prefersReducedMotion()) {
		return {
			duration: 0,
			tick: () => {}
		};
	}

	return {
		duration,
		delay,
		easing: EASINGS.appear,
		css: (t) => `
			opacity: ${t};
		`
	};
}

/**
 * Dropdown pop-up transition
 */
export function dropdownPop(node: Element, { duration = 200, delay = 0 }: { duration?: number; delay?: number } = {}): TransitionConfig {
	if (prefersReducedMotion()) {
		return {
			duration: 0,
			tick: () => {}
		};
	}

	return {
		duration,
		delay,
		easing: EASINGS.size,
		css: (t) => `
			transform: translateY(${8 * (1 - t)}px) scale(${0.95 + 0.05 * t});
			opacity: ${t};
		`
	};
}

/**
 * List item stagger - slide from left with fade
 */
export function listItem(node: Element, { duration = 300, delay = 0 }: { duration?: number; delay?: number } = {}): TransitionConfig {
	if (prefersReducedMotion()) {
		return {
			duration: 0,
			tick: () => {}
		};
	}

	return {
		duration,
		delay,
		easing: EASINGS.appear,
		css: (t) => `
			transform: translateY(${16 * (1 - t)}px);
			opacity: ${t};
		`
	};
}
