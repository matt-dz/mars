import { redirect, type Handle } from '@sveltejs/kit';
import { UserSchema, type User } from '$lib/api/types';
import { extractAuthCookies, patchCookies, wrapWithCredentials } from '@/http';
import { refreshSession, verifySession } from '@/api';
import { HTTPError, isRecoverableAuthError, isUnrecoverableAuthError } from '@/api/errors';

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

	// Verify session with backend - refresh logic is handled by the http module
	try {
		console.debug('[hooks] Verifying session...');
		await verifySession(
			wrapWithCredentials(
				event.fetch,
				tokens.accessToken ?? '',
				tokens.refreshToken,
				tokens.csrfToken
			)
		);
		console.debug('[hooks] Session verified');
	} catch (e) {
		console.debug('[hooks] Session verification failed:', e);
		if (!(e instanceof HTTPError)) {
			throw e;
		}
		if (isUnrecoverableAuthError(e.errorCode)) {
			console.debug('[hooks] Received an unrecoverable auth error, redirecting to login');
			redirect(302, '/login');
		}
		if (!isRecoverableAuthError(e.errorCode)) {
			console.debug('[hooks] Received an non-recoverable error, redirecting to login');
			redirect(302, '/login');
		}

		try {
			console.debug('[hooks] Refreshing session');
			const res = await refreshSession(
				wrapWithCredentials(
					event.fetch,
					tokens.accessToken ?? '',
					tokens.refreshToken,
					tokens.csrfToken
				)
			);

			// Patch cookies
			console.debug(
				'[hooks] Successfully refreshed session, patching cookies:',
				res.headers.getSetCookie()
			);
			patchCookies(res, event.cookies);
		} catch (e2) {
			console.debug('[hooks] Session refresh failed:', e);
			throw e2;
		}
	}

	// Session is valid - set user in locals
	// TODO: Parse user from verify response when backend returns user data
	// For now, use mock user data
	const user: User = UserSchema.parse({
		id: 'user-1',
		email: 'user@example.com',
		role: 'user' as const
	});

	event.locals.user = user;

	return resolve(event);
};
