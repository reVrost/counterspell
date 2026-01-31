import { sequence } from '@sveltejs/kit/hooks';
import type { Handle, HandleServerError } from '@sveltejs/kit';

const API_BASE_URL = 'http://localhost:8710';

const proxyApi: Handle = async ({ event, resolve }) => {
	if (event.url.pathname.startsWith('/api/v1')) {
		const apiUrl = new URL(event.url.pathname + event.url.search, API_BASE_URL);

		const headers = new Headers();
		event.request.headers.forEach((value, key) => {
			headers.set(key, value);
		});

		const response = await fetch(apiUrl.toString(), {
			method: event.request.method,
			headers,
			body: event.request.body,
			credentials: 'include'
		});

		return new Response(response.body, {
			status: response.status,
			headers: response.headers
		});
	}

	return resolve(event);
};

export const handleError: HandleServerError = ({ error, event }) => {
    console.error('❌ Server-side error detected:');
    console.error('URL:', event.url.toString());
    console.error('Error:', error);

    return {
        message: 'Internal Error',
        code: 'INTERNAL_ERROR'
    };
};

export const handle = sequence(proxyApi);
