import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		port: 5173,
		host: '0.0.0.0',
		allowedHosts: ['dan-severe-overly.ngrok-free.dev', 'host.docker.internal'],
		proxy: {
			'/api/v1': {
				target: 'http://localhost:8710',
				changeOrigin: true
			}
		}
	}
});
