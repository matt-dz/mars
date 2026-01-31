import { redirect, type Handle } from '@sveltejs/kit';
import { UserSchema, type User } from '$lib/api/types';
import { extractAuthCookies, patchCookies, wrapWithCredentials } from '$lib/http';
import { verifySession } from '@/api';

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

	// Verify session with backend - refresh logic is handled by the http module
	try {
		console.debug('[hooks] Verifying session...');
		const tokens = extractAuthCookies(event.cookies);
		if (!tokens.refreshToken || !tokens.csrfToken) {
			console.debug('[hooks] Missing access token, refresh token, or csrf token; skipping auth');
			redirect(302, '/login');
		}
		const response = await verifySession(
			wrapWithCredentials(
				event.fetch,
				tokens.accessToken ?? '',
				tokens.refreshToken,
				tokens.csrfToken
			)
		);

		if (!response.ok) {
			console.debug('[hooks] Verify returned non-ok status:', response.status);
			redirect(302, '/login');
		}

		// Patch cookies
		patchCookies(response, event.cookies);

		console.debug('[hooks] Session verified');
	} catch (e) {
		console.debug('[hooks] Session verification failed:', e);
		redirect(302, '/login');
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
