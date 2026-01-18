import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		port: 5173,
		proxy: {
			'/api': {
				target: 'http://localhost:8710',
				changeOrigin: true
			},
			'/events': {
				target: 'http://localhost:8710',
				changeOrigin: true,
				ws: true
			},
			'/auth': {
				target: 'http://localhost:8710',
				changeOrigin: true
			},
			'/action': {
				target: 'http://localhost:8710',
				changeOrigin: true
			},
			'/add-task': {
				target: 'http://localhost:8710',
				changeOrigin: true
			},
			'/settings': {
				target: 'http://localhost:8710',
				changeOrigin: true
			},
			'/transcribe': {
				target: 'http://localhost:8710',
				changeOrigin: true
			},
			'/github': {
				target: 'http://localhost:8710',
				changeOrigin: true
			},
			'/disconnect': {
				target: 'http://localhost:8710',
				changeOrigin: true
			}
		}
	}
});
