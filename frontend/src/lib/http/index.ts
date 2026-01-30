import ky, { type KyInstance, type KyResponse, type Options } from 'ky';
import {
	ACCESS_TOKEN_COOKIE_NAME,
	CSRF_HEADER,
	CSRF_TOKEN_COOKIE_NAME,
	REFRESH_TOKEN_COOKIE_NAME
} from '$lib/auth';
import { browser } from '$app/environment';
import { env } from '$env/dynamic/public';

const retryCodes = [408, 413, 429, 500, 502, 503, 504];

function getPrefixUrl(): string {
	if (browser) return '';
	return env.PUBLIC_API_URL ?? '';
}

export function getCsrfToken(): string | null {
	if (!browser) return null;
	const cookies = document.cookie.split(';');
	for (let i = 0; i < cookies.length; i++) {
		const c = cookies[i].trim();
		const splitIdx = c.indexOf('=');
		const key = c.slice(0, splitIdx);
		const val = c.slice(splitIdx + 1);
		if (key === CSRF_TOKEN_COOKIE_NAME) {
			return val;
		}
	}
	return null;
}

/**
 * Injects CSRF token into request headers for state-changing requests.
 * Only works in browser context.
 */
function injectCSRFToken(request: Request) {
	const token = getCsrfToken();
	if (token) request.headers.set(CSRF_HEADER, token);
}

const baseOptions: Options = {
	prefixUrl: getPrefixUrl(),
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

const fetchFn = ky.create({
	...baseOptions
});

export function wrap(fetchFn: typeof fetch): KyInstance {
	return ky.create({
		...baseOptions,
		fetch: fetchFn
	});
}

export function wrapWithCredentials(
	fetchFn: typeof fetch,
	accessToken: string,
	refreshToken: string,
	csrfToken: string | null = ''
): KyInstance {
	return ky.create({
		...baseOptions,
		headers: {
			...baseOptions.headers,
			Cookie: `${ACCESS_TOKEN_COOKIE_NAME}=${accessToken}; ${REFRESH_TOKEN_COOKIE_NAME}=${refreshToken}; ${CSRF_TOKEN_COOKIE_NAME}=${csrfToken}`
		},
		fetch: fetchFn
	});
}

export function isRetryable(response: KyResponse) {
	return retryCodes.includes(response.status);
}

type FetchFn = typeof fetchFn;

export type { FetchFn };
export { baseOptions };
export default fetchFn;
