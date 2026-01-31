import ky, { type KyInstance, type KyResponse, type Options } from 'ky';
import {
	ACCESS_TOKEN_COOKIE_NAME,
	CSRF_HEADER,
	CSRF_TOKEN_COOKIE_NAME,
	REFRESH_TOKEN_COOKIE_NAME
} from '$lib/auth';
import { browser } from '$app/environment';
import { env } from '$env/dynamic/public';
import {
	ErrorCode,
	isUnrecoverableAuthError,
	isRecoverableAuthError,
	AuthenticationError,
	LoginResponseSchema
} from '$lib/api/errors';
import { goto } from '$app/navigation';
import { resolve } from '$app/paths';
import type { Cookies } from '@sveltejs/kit';
import * as setCookie from 'set-cookie-parser';

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

export function patchCookies(response: KyResponse, cookies: Cookies) {
	setCookie.parse(response.headers.getSetCookie()).map(({ name, value, ...opts }) => {
		cookies.set(name, value, {
			...opts,
			httpOnly: opts.httpOnly ?? false,
			secure: opts.secure ?? false,
			sameSite: opts.sameSite as boolean | 'lax' | 'strict' | 'none' | undefined,
			path: opts.path ?? '/'
		});
	});
}

/**
 * Injects CSRF token into request headers for state-changing requests.
 * Only works in browser context.
 */
function injectCSRFToken(request: Request) {
	const token = getCsrfToken();
	if (token) request.headers.set(CSRF_HEADER, token);
}

// Track if a refresh is in progress to prevent concurrent refresh attempts
let refreshPromise: Promise<Response> | null = null;

/**
 * Attempts to refresh the access token using the refresh token.
 * Uses the provided ky instance.
 */
async function refreshAccessToken(kyInstance: KyInstance): Promise<Response> {
	// If a refresh is already in progress, wait for it
	if (refreshPromise) {
		return refreshPromise;
	}

	// Send empty JSON body - oapi-codegen always tries to decode the body
	refreshPromise = kyInstance.post('api/auth/refresh', { json: {} }).then((r) => r);

	try {
		const response = await refreshPromise;
		return response;
	} finally {
		refreshPromise = null;
	}
}

/**
 * Creates afterResponse hooks for automatic token refresh on 401.
 * The kyInstance parameter is used for making the refresh and retry requests.
 */
function createRefreshHooks(kyInstance: KyInstance) {
	return [
		async (request: Request, _options: Options, response: Response): Promise<Response | void> => {
			// Only handle 401 errors
			if (response.status !== 401) {
				return;
			}

			console.debug('[auth] Received 401 for:', request.url);

			// Don't try to refresh on auth endpoints to avoid infinite loops
			const url = new URL(request.url);
			if (url.pathname === '/api/auth/refresh' || url.pathname === '/api/login') {
				console.debug('[auth] Skipping refresh for auth endpoint:', url.pathname);
				return;
			}

			// Try to parse the error to check if it's unrecoverable
			try {
				const errorBody = await response.clone().json();
				console.debug('[auth] Error body:', errorBody);

				if (errorBody.code && isUnrecoverableAuthError(errorBody.code)) {
					console.debug('[auth] Unrecoverable error:', errorBody.code);
					if (browser) {
						window.location.href = '/login';
					}
					throw new AuthenticationError(errorBody.message, errorBody.code);
				}

				// If it's not a recoverable auth error, don't try to refresh
				if (errorBody.code && !isRecoverableAuthError(errorBody.code)) {
					console.debug('[auth] Non-recoverable error, skipping refresh:', errorBody.code);
					return;
				}
			} catch (e) {
				if (e instanceof AuthenticationError) throw e;
				console.debug('[auth] Could not parse error body, attempting refresh');
			}

			// Attempt to refresh the token
			console.debug('[auth] Attempting token refresh...');
			try {
				const refreshResponse = await refreshAccessToken(kyInstance);
				console.debug('[auth] Refresh response status:', refreshResponse.status);

				if (!refreshResponse.ok) {
					console.debug('[auth] Refresh failed');
					// Refresh failed
					try {
						const errorBody = await refreshResponse.json();
						console.debug('[auth] Refresh error:', errorBody);
						if (browser) {
							goto(resolve('/login'));
						}
						throw new AuthenticationError(
							errorBody.message || 'Session expired',
							errorBody.code || ErrorCode.ExpiredRefreshToken
						);
					} catch (e) {
						if (e instanceof AuthenticationError) throw e;
						if (browser) {
							goto(resolve('/login'));
						}
						throw new AuthenticationError('Session expired', ErrorCode.ExpiredRefreshToken);
					}
				}

				// Validate refresh response
				const refreshData = await refreshResponse.json();
				const parsed = LoginResponseSchema.safeParse(refreshData);
				if (!parsed.success) {
					console.debug('[auth] Invalid refresh response schema:', parsed.error);
					if (browser) {
						goto(resolve('/login'));
					}
					throw new AuthenticationError('Invalid refresh response', ErrorCode.InternalServerError);
				}

				console.debug('[auth] Refresh successful, retrying request');

				// Refresh succeeded - retry the original request
				if (browser) {
					// In browser, cookies are automatically set - just retry
					return ky(request);
				} else {
					// On server, extract cookies from refresh response
					const setCookieHeaders = refreshResponse.headers.getSetCookie();
					console.debug('[auth] Extracted', setCookieHeaders.length, 'Set-Cookie headers');

					let accessToken = '';
					let refreshToken = '';
					let csrfToken = '';

					for (const setCookie of setCookieHeaders) {
						const cookieValue = setCookie.split(';')[0];
						if (cookieValue.startsWith(`${ACCESS_TOKEN_COOKIE_NAME}=`)) {
							accessToken = cookieValue.split('=')[1];
						} else if (cookieValue.startsWith(`${REFRESH_TOKEN_COOKIE_NAME}=`)) {
							refreshToken = cookieValue.split('=')[1];
						} else if (cookieValue.startsWith(`${CSRF_TOKEN_COOKIE_NAME}=`)) {
							csrfToken = cookieValue.split('=')[1];
						}
					}

					console.debug('[auth] Retrying with fresh credentials');
					// Create a new ky instance with the fresh credentials and retry
					const retryInstance = wrapWithCredentials(fetch, accessToken, refreshToken, csrfToken);
					return retryInstance(request);
				}
			} catch (e) {
				if (e instanceof AuthenticationError) throw e;
				console.debug('[auth] Refresh error:', e);
				if (browser) {
					goto(resolve('/login'));
				}
				throw new AuthenticationError('Failed to refresh session', ErrorCode.ExpiredRefreshToken);
			}
		}
	];
}

const baseOptions: Options = {
	prefixUrl: getPrefixUrl(),
	timeout: 15 * 1000,
	retry: {
		retryOnTimeout: true,
		limit: 4,
		backoffLimit: 10 * 1000,
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

// Create base instance without refresh hooks first
const fetchFn = ky.create(baseOptions);

export function wrap(customFetch: typeof fetch): KyInstance {
	// Create base instance first
	if (browser) {
		return fetchFn.extend({
			hooks: {
				...baseOptions.hooks,
				afterResponse: createRefreshHooks(fetchFn)
			}
		});
	}

	return fetchFn.extend({
		fetch: customFetch
	});
}

export function wrapWithCredentials(
	customFetch: typeof fetch,
	accessToken: string,
	refreshToken: string,
	csrfToken: string | null = ''
): KyInstance {
	const instance = wrap(customFetch).extend({
		headers: {
			...baseOptions.headers,
			Cookie: `${ACCESS_TOKEN_COOKIE_NAME}=${accessToken}; ${REFRESH_TOKEN_COOKIE_NAME}=${refreshToken}; ${CSRF_TOKEN_COOKIE_NAME}=${csrfToken}`
		}
	});

	return instance.extend({
		hooks: {
			...baseOptions.hooks,
			afterResponse: createRefreshHooks(instance)
		}
	});
}

export function isRetryable(response: KyResponse) {
	return retryCodes.includes(response.status);
}

export function extractAuthCookies(cookies: Cookies) {
	return {
		accessToken: cookies.get(ACCESS_TOKEN_COOKIE_NAME),
		refreshToken: cookies.get(REFRESH_TOKEN_COOKIE_NAME),
		csrfToken: cookies.get(CSRF_TOKEN_COOKIE_NAME)
	};
}

// Re-export for convenience
export { AuthenticationError, ErrorCode, isUnrecoverableAuthError, isRecoverableAuthError };

type FetchFn = typeof fetchFn;

export type { FetchFn };
export { baseOptions };
export default fetchFn;
