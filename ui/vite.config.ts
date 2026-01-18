import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		port: 5173,
		proxy: {
			'/api/v1': {
				target: 'http://localhost:8710',
				changeOrigin: true
			}
		}
	}
});
