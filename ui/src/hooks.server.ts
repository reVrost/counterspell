import type { HandleServerError } from '@sveltejs/kit';

export const handleError: HandleServerError = ({ error, event }) => {
    console.error('❌ Server-side error detected:');
    console.error('URL:', event.url.toString());
    console.error('Error:', error);

    return {
        message: 'Internal Error',
        code: 'INTERNAL_ERROR'
    };
};
