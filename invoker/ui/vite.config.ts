import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
	plugins: [sveltekit(), tailwindcss()],
	server: {
		port: 5174,
		host: '0.0.0.0',
		allowedHosts: ['dan-severe-overly.ngrok-free.dev', 'host.docker.internal'],
		proxy: {
			'/api': {
				target: 'http://localhost:8079',
				changeOrigin: true
			}
		}
	},
	ssr: {
		noExternal: ['svelte-sonner']
	}
});


