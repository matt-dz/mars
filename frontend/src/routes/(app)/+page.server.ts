import type { PageServerLoad } from './$types';
import { getPlaylists } from '$lib/api/playlists';
import { extractAuthCookies, wrapWithCredentials } from '@/http';
import { redirect } from '@sveltejs/kit';
import { HTTPError } from '@/api/errors';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	try {
		const tokens = extractAuthCookies(cookies);
		if (!tokens.accessToken || !tokens.refreshToken || !tokens.csrfToken) {
			console.debug('[home] Missing access token, refresh token, or csrf token; skipping auth');
			redirect(302, '/login');
		}

		console.debug('[home] Fetching playlists');
		const fetchFn = wrapWithCredentials(
			fetch,
			tokens.accessToken,
			tokens.refreshToken,
			tokens.csrfToken
		);
		const playlists = await getPlaylists(fetchFn);
		return { playlists };
	} catch (e) {
		console.error('[home] Fetching integration statuses failed:', e);
		if (e instanceof HTTPError) {
			if (e.statusCode === 401) {
				console.debug('[home] redirecting user to login');
				redirect(302, '/login');
			}
		}
		throw e;
	}
};
