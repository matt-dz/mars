import { type Options } from 'ky';
import { redirect, type Handle } from '@sveltejs/kit';
import { extractAuthCookies, patchCookies, wrapWithCredentials, baseOptions } from '@/http';
import { verifySession } from '@/api';
import {
	HTTPError,
	isRecoverableAuthError,
	isUnrecoverableAuthError,
	ApiErrorSchema
} from '@/api/errors';
import {
	ACCESS_TOKEN_COOKIE_NAME,
	CSRF_TOKEN_COOKIE_NAME,
	REFRESH_TOKEN_COOKIE_NAME
} from '@/auth';

// Routes that don't require authentication
const PUBLIC_ROUTES = ['/login', '/api/'];

function isPublicRoute(pathname: string): boolean {
	return PUBLIC_ROUTES.some((route) => pathname === route || pathname.startsWith(route));
}

export const handle: Handle = async ({ event, resolve }) => {
	console.debug('[hooks] Request:', event.url.pathname);

	// Skip auth check for public routes
	if (isPublicRoute(event.url.pathname)) {
		console.debug('[hooks] Public route, skipping auth');
		return resolve(event);
	}

	const tokens = extractAuthCookies(event.cookies);
	if (!tokens.refreshToken || !tokens.csrfToken) {
		console.debug('[hooks] Missing access token, refresh token, or csrf token; skipping auth');
		redirect(302, '/login');
	}

	let fetchFn = wrapWithCredentials(
		event.fetch,
		tokens.accessToken ?? '',
		tokens.refreshToken,
		tokens.csrfToken
	);
	const afterResponse = async (
		request: Request,
		options: Options,
		response: Response
	): Promise<Response | void> => {
		// Only handle 401 errors
		if (response.status !== 401) {
			return;
		}

		console.debug('[hooks][auth] Received 401 for:', request.url);
		const errorBody = ApiErrorSchema.safeParse(await response.clone().json());
		if (!errorBody.success) {
			return response;
		}
		console.debug('[hooks][auth] Error body:', errorBody);
		if (!isRecoverableAuthError(errorBody.data.code)) {
			console.debug('[hooks][auth] Non-recoverable error:', errorBody.data.code);
			return response;
		}

		console.debug('[hooks][auth] Attempting token refresh...');
		let accessToken: string = '';
		let refreshToken: string = '';
		let csrfToken: string = '';
		try {
			const res = await fetchFn.post('api/auth/refresh', {
				json: { refresh_token: tokens.refreshToken }
			});
			console.debug('[hooks][auth] Successfully refreshed tokens');
			patchCookies(res, event.cookies);
			res.headers.getSetCookie().forEach((c) => {
				const parts = c.split(';');
				if (parts.length === 0) return;

				const splitIdx = parts[0].indexOf('=');
				const value = parts[0].slice(splitIdx + 1);
				if (parts[0].startsWith(`${ACCESS_TOKEN_COOKIE_NAME}=`)) {
					accessToken = value;
				} else if (parts[0].startsWith(`${REFRESH_TOKEN_COOKIE_NAME}=`)) {
					refreshToken = value;
				} else if (parts[0].startsWith(`${CSRF_TOKEN_COOKIE_NAME}=`)) {
					csrfToken = value;
				}
			});
		} catch (e) {
			console.error('[hooks][auth] Failed to refresh token', e);
			return response;
		}

		console.debug('[hooks][auth] Retrying request...');
		const newFetch = fetchFn.extend({
			headers: {
				...options?.headers,
				Cookie: `${ACCESS_TOKEN_COOKIE_NAME}=${accessToken}; ${REFRESH_TOKEN_COOKIE_NAME}=${refreshToken}; ${CSRF_TOKEN_COOKIE_NAME}=${csrfToken}`
			}
		});
		return newFetch(request);
	};
	fetchFn = fetchFn.extend({
		hooks: {
			...baseOptions.hooks,
			afterResponse: [...(baseOptions.hooks?.afterResponse ?? []), afterResponse]
		}
	});

	// Verify session with backend - refresh logic is handled by the http module
	try {
		console.debug('[hooks] Verifying session...');
		const user = await verifySession(fetchFn);
		event.locals.user = user;
		console.debug('[hooks] Session verified');
	} catch (e1) {
		console.debug('[hooks] Session verification failed:', e1);
		if (!(e1 instanceof HTTPError)) {
			throw e1;
		}
		if (isUnrecoverableAuthError(e1.errorCode)) {
			console.debug('[hooks] Received an unrecoverable auth error, redirecting to login');
			redirect(302, '/login');
		}
		if (!isRecoverableAuthError(e1.errorCode)) {
			console.debug('[hooks] Received an non-recoverable error, redirecting to login');
			redirect(302, '/login');
		}
	}

	return resolve(event);
};
