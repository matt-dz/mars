import { redirect, type Handle } from '@sveltejs/kit';
import { UserSchema, type User } from '$lib/api/types';

export const handle: Handle = async ({ event, resolve }) => {
	// Skip auth check for login page and API routes
	if (event.url.pathname === '/login' || event.url.pathname.startsWith('/api/')) {
		return resolve(event);
	}

	// Verify session with backend (event.fetch inherits cookies from original request)
	// const response = await event.fetch('/api/auth/verify');
	//
	// if (!response.ok) {
	// 	redirect(302, '/login');
	// }

	// TODO: Parse user from verify response when backend supports it
	// For now, use mock user data
	const user: User = UserSchema.parse({
		id: 'user-1',
		email: 'user@example.com',
		role: 'user' as const
	});

	event.locals.user = user;

	return resolve(event);
};
