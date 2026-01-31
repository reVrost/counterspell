import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		port: 5173,
		host: '0.0.0.0',
		allowedHosts: ['dan-severe-overly.ngrok-free.dev', 'host.docker.internal']
		// Note: Proxy is handled in hooks.server.ts for SvelteKit
	}
});
