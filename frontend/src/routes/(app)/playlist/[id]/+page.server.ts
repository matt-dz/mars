import type { PageServerLoad } from './$types';
import { extractAuthCookies, wrapWithCredentials } from '@/http';
import { getPlaylist } from '$lib/api/playlists';
import { error, redirect } from '@sveltejs/kit';
import { HTTPError } from '@/api/errors';

export const load: PageServerLoad = async ({ params, fetch, cookies }) => {
	try {
		const tokens = extractAuthCookies(cookies);
		if (!tokens.accessToken || !tokens.refreshToken || !tokens.csrfToken) {
			console.debug('[playlist] Missing access token, refresh token, or csrf token; skipping auth');
			redirect(302, '/login');
		}

		console.debug('[playlist] Fetching playlist');
		const fetchFn = wrapWithCredentials(
			fetch,
			tokens.accessToken,
			tokens.refreshToken,
			tokens.csrfToken
		);
		const playlist = await getPlaylist(params.id, fetchFn);
		return { playlist };
	} catch (e) {
		console.error('[playlist] Fetching integration statuses failed:', e);
		if (e instanceof HTTPError) {
			if (e.statusCode === 401) {
				console.debug('[playlist] redirecting user to login');
				redirect(302, '/login');
			}
			if (e.statusCode === 404) {
				console.debug('[playlist] playlist not found');
				error(404);
			}
		}
		throw e;
	}
};
