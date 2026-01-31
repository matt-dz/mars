import type { PageServerLoad } from './$types';
import { extractAuthCookies, wrapWithCredentials } from '@/http';
import { getSpotifyStatus } from '$lib/api/spotify';
import { redirect } from '@sveltejs/kit';
import { HTTPError } from '@/api/errors';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	try {
		const { accessToken, refreshToken, csrfToken } = extractAuthCookies(cookies);
		if (!accessToken || !refreshToken || !csrfToken) {
			console.debug(
				'[integrations] access token, refresh token, and csrf tokens are required - missing one of them, skipping auth'
			);
			return redirect(303, '/login');
		}
		const fetchFn = wrapWithCredentials(fetch, accessToken, refreshToken, csrfToken);

		console.debug('[integrations] getting spotify status');
		const spotifyStatus = await getSpotifyStatus(fetchFn);
		return { spotifyStatus };
	} catch (e) {
		console.error('[integrations] failed to get spotify status', e);
		if (e instanceof HTTPError) {
			if (e.statusCode === 401) {
				console.debug('[home] redirecting user to login');
				redirect(302, '/login');
			}
		}
		throw e;
	}
};
