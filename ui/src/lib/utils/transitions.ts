/**
 * Svelte 5 transition utilities
 * Inspired by Vercel UI motion guidelines
 */

import type { TransitionConfig } from 'svelte/transition';
import { cubicOut, cubicInOut, quintOut } from 'svelte/easing';
import { prefersReducedMotion } from './animations';

// Re-export for convenience
export { prefersReducedMotion };

// Duration constants
export const DURATIONS = {
  fast: 150,
  quick: 200,
  normal: 300,
  slow: 400,
  deliberate: 500,
} as const;

// Easing functions (must be FUNCTIONS, not strings)
export const EASINGS = {
  size: quintOut,
  position: cubicOut,
  appear: cubicOut,
  ui: cubicInOut,
} as const;

/**
 * Fade transition
 */
export function fade(
  node: Element,
  { duration = 200, delay = 0 }: { duration?: number; delay?: number } = {}
): TransitionConfig {
  if (prefersReducedMotion()) {
    return { duration: 0, tick: () => {} };
  }

  return {
    duration,
    delay,
    easing: EASINGS.appear,
    css: (t) => `opacity: ${t};`,
  };
}

/**
 * Slide + fade transition
 */
export function slide(
  node: Element,
  {
    direction = 'up',
    duration = 300,
    delay = 0,
    opacity = true,
  }: {
    direction?: 'up' | 'down' | 'left' | 'right';
    duration?: number;
    delay?: number;
    opacity?: boolean;
  } = {}
): TransitionConfig {
  if (prefersReducedMotion()) {
    return { duration: 0, tick: () => {} };
  }

  const distance = 32;

  return {
    duration,
    delay,
    easing: EASINGS.position,
    css: (t) => {
      const move = (1 - t) * distance;

      const transform =
        direction === 'up'
          ? `translateY(${move}px)`
          : direction === 'down'
            ? `translateY(${-move}px)`
            : direction === 'left'
              ? `translateX(${move}px)`
              : `translateX(${-move}px)`;

      return `
				transform: ${transform};
				${opacity ? `opacity: ${t};` : ''}
			`;
    },
  };
}

/**
 * Scale + fade transition
 */
export function scale(
  node: Element,
  {
    duration = 200,
    delay = 0,
    startScale = 0.95,
  }: {
    duration?: number;
    delay?: number;
    startScale?: number;
  } = {}
): TransitionConfig {
  if (prefersReducedMotion()) {
    return { duration: 0, tick: () => {} };
  }

  return {
    duration,
    delay,
    easing: EASINGS.size,
    css: (t) => `
			transform: scale(${startScale + (1 - startScale) * t});
			opacity: ${t};
		`,
  };
}

/**
 * Modal slide-up transition
 */
export function modalSlideUp(
  node: Element,
  { duration = 300, delay = 0 }: { duration?: number; delay?: number } = {}
): TransitionConfig {
  if (prefersReducedMotion()) {
    return { duration: 0, tick: () => {} };
  }

  return {
    duration,
    delay,
    easing: EASINGS.size,
    css: (t) => `
			transform: translateY(${32 * (1 - t)}px) scale(${0.95 + 0.05 * t});
			opacity: ${t};
		`,
  };
}

/**
 * Backdrop fade transition
 */
export function backdropFade(
  node: Element,
  { duration = 300, delay = 0 }: { duration?: number; delay?: number } = {}
): TransitionConfig {
  if (prefersReducedMotion()) {
    return { duration: 0, tick: () => {} };
  }

  return {
    duration,
    delay,
    easing: EASINGS.appear,
    css: (t) => `opacity: ${t};`,
  };
}

/**
 * Dropdown pop transition
 */
export function dropdownPop(
  node: Element,
  { duration = 200, delay = 0 }: { duration?: number; delay?: number } = {}
): TransitionConfig {
  if (prefersReducedMotion()) {
    return { duration: 0, tick: () => {} };
  }

  return {
    duration,
    delay,
    easing: EASINGS.size,
    css: (t) => `
			transform: translateY(${8 * (1 - t)}px) scale(${0.95 + 0.05 * t});
			opacity: ${t};
		`,
  };
}

/**
 * List item transition (stagger friendly)
 */
export function listItem(
  node: Element,
  { duration = 300, delay = 0 }: { duration?: number; delay?: number } = {}
): TransitionConfig {
  if (prefersReducedMotion()) {
    return { duration: 0, tick: () => {} };
  }

  return {
    duration,
    delay,
    easing: EASINGS.appear,
    css: (t) => `
			transform: translateY(${16 * (1 - t)}px);
			opacity: ${t};
		`,
  };
}

export function spawn(
  node: Element,
  {
    delay = 0,
    duration = 180,
    y = 12,
    startScale = 0.98,
  }: {
    delay?: number;
    duration?: number;
    y?: number;
    startScale?: number;
  } = {}
): TransitionConfig {
  if (prefersReducedMotion()) {
    return { duration: 0, tick: () => {} };
  }

  return {
    delay,
    duration,
    easing: cubicOut,
    css: (t) => `
			transform:
				translateY(${(1 - t) * y}px)
				scale(${startScale + (1 - startScale) * t});
			opacity: ${t};
		`,
  };
}
