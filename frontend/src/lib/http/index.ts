import ky, { type KyResponse, type Options } from 'ky';
import { CSRF_HEADER, CSRF_TOKEN_COOKIE_NAME } from '$lib/auth';
import { browser } from '$app/environment'

const retryCodes = [408, 413, 429, 500, 502, 503, 504];

export function getCsrfToken(): string | null {
	if (!browser) return null;
	const cookies = document.cookie.split(';');
	for (let i = 0; i < cookies.length; i++) {
		const c = cookies[i].trim();
		const splitIdx = c.indexOf('=');
		const key = c.slice(0, splitIdx);
		const val = c.slice(splitIdx + 1);
		if (key === CSRF_TOKEN_COOKIE_NAME) {
			return val
		}
	}
	return null;
}

/**
 * Injects CSRF token into request headers for state-changing requests.
 * Only works in browser context.
 */
function injectCSRFToken(request: Request) {
	const token = getCsrfToken()
	if (token) request.headers.set(CSRF_HEADER, token)
}

const baseOptions: Options = {
	timeout: 15 * 1000,
	retry: {
		retryOnTimeout: true,
		limit: 4,
		backoffLimit: 10 * 1000, // 10 seconds,
		statusCodes: retryCodes
	},
	credentials: 'include',
	hooks: {
		beforeRequest: [
			(request) => {
				injectCSRFToken(request);
			}
		]
	}
};

const fetch = ky.create({
	...baseOptions,
});

export function isRetryable(response: KyResponse) {
	return retryCodes.includes(response.status);
}

type FetchType = typeof fetch;

export type { FetchType };
export { baseOptions };
export default fetch;
