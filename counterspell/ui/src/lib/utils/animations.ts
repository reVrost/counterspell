/**
 * Animation utilities following Vercel Web Interface Guidelines
 * - Honor prefers-reduced-motion
 * - Prefer CSS over JS animations
 * - Use GPU-accelerated properties (transform, opacity)
 * - Never transition: all - list only specific properties
 */

// Check if user prefers reduced motion
export const prefersReducedMotion = () => {
	if (typeof window === 'undefined') return false;
	return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
};

// Easing functions based on what changes
export const easings = {
	// Size changes (expanding/collapsing)
	size: 'cubic-bezier(0.16, 1, 0.3, 1)',

	// Position changes (sliding, moving)
	position: 'cubic-bezier(0.2, 0, 0, 1)',

	// Appear/disappear (fading)
	appear: 'cubic-bezier(0.16, 1, 0.3, 1)',

	// Quick UI feedback (buttons, toggles)
	ui: 'cubic-bezier(0.4, 0, 0.2, 1)',

	// Spring-like motion
	spring: 'cubic-bezier(0.175, 0.885, 0.32, 1.275)'
};

// Duration based on distance and what changes
export const durations = {
	// Micro-interactions (buttons, hover)
	fast: 150,

	// Quick transitions (dropdowns, small modals)
	quick: 200,

	// Standard transitions (modals, panels)
	normal: 300,

	// Slower transitions (page transitions, large modals)
	slow: 400,

	// Deliberate animations (complex flows)
	deliberate: 500
};

// Animation classes for common patterns
export const animations = {
	// Modal slide up + fade in
	modalEnter: {
		from: 'opacity-0 translate-y-8 scale-95',
		to: 'opacity-100 translate-y-0 scale-100',
		duration: durations.normal,
		easing: easings.size
	},

	// Modal slide down + fade out
	modalExit: {
		from: 'opacity-100 translate-y-0 scale-100',
		to: 'opacity-0 translate-y-4 scale-95',
		duration: durations.quick,
		easing: easings.ui
	},

	// Drawer slide in from side
	drawerEnter: {
		from: 'translate-x-full',
		to: 'translate-x-0',
		duration: durations.normal,
		easing: easings.position
	},

	// Dropdown expand
	dropdownEnter: {
		from: 'opacity-0 scale-95 -translate-y-2',
		to: 'opacity-100 scale-100 translate-y-0',
		duration: durations.quick,
		easing: easings.size
	},

	// Toast slide down + fade in
	toastEnter: {
		from: 'opacity-0 -translate-y-8',
		to: 'opacity-100 translate-y-0',
		duration: durations.quick,
		easing: easings.position
	},

	// Toast fade out
	toastExit: {
		from: 'opacity-100 translate-y-0',
		to: 'opacity-0 -translate-y-4',
		duration: durations.fast,
		easing: easings.ui
	},

	// List item stagger
	listItemEnter: {
		from: 'opacity-0 translate-y-4',
		to: 'opacity-100 translate-y-0',
		duration: durations.normal,
		easing: easings.appear
	},

	// Button press scale
	buttonPress: {
		from: 'scale-100',
		to: 'scale-95',
		duration: durations.fast,
		easing: easings.ui
	},

	// Panel expand
	panelExpand: {
		from: 'opacity-0 max-h-0',
		to: 'opacity-100 max-h-[500px]',
		duration: durations.normal,
		easing: easings.size
	}
};

// CSS animation string builder
export const buildTransition = (
	properties: string[],
	duration: number = durations.normal,
	easing: string = easings.appear,
	delay: number = 0
) => {
	const transition = properties.map((prop) => `${prop} ${duration}ms ${easing} ${delay}ms`).join(', ');
	return transition;
};

// Animation keyframes for custom animations
export const keyframes = {
	// Pulse for loading states
	pulse: {
		'0%, 100%': { opacity: '1' },
		'50%': { opacity: '0.5' }
	},

	// Spin for loaders
	spin: {
		from: { transform: 'rotate(0deg)' },
		to: { transform: 'rotate(360deg)' }
	},

	// Bounce
	bounce: {
		'0%, 100%': { transform: 'translateY(0)' },
		'50%': { transform: 'translateY(-25%)' }
	},

	// Shimmer for skeleton screens
	shimmer: {
		from: { backgroundPosition: '-1000px 0' },
		to: { backgroundPosition: '1000px 0' }
	}
};

// Stagger delay calculator for list animations
export const getStaggerDelay = (index: number, baseDelay: number = 50) => {
	return index * baseDelay;
};

// Animation state for Svelte components
export interface AnimationState {
	isAnimating: boolean;
	direction: 'enter' | 'exit';
}

// Transition config for Svelte transitions
export const createTransition = (config: {
	duration?: number;
	easing?: string;
	delay?: number;
}) => ({
	duration: config.duration ?? durations.normal,
	easing: config.easing ?? easings.appear,
	delay: config.delay ?? 0
});
