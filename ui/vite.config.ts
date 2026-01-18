import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		port: 5173,
		allowedHosts: ['dan-severe-overly.ngrok-free.dev'],
		proxy: {
			'/api/v1': {
				target: 'http://127.0.0.1:8710',
				changeOrigin: true
			}
		}
	}
});
