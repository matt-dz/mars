import type { PageServerLoad } from './$types';
import { wrapWithCredentials } from '@/http';
import { getSpotifyStatus } from '$lib/api/spotify';
import { ACCESS_TOKEN_COOKIE_NAME, REFRESH_TOKEN_COOKIE_NAME } from '@/auth';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	try {
		const accessToken = cookies.get(ACCESS_TOKEN_COOKIE_NAME);
		const refreshToken = cookies.get(REFRESH_TOKEN_COOKIE_NAME);
		if (!accessToken) {
			console.error('no access token, sending to login');
			return redirect(303, '/login');
		}
		if (!refreshToken) {
			console.error('no refresh token, sending to login');
			return redirect(303, '/login');
		}
		const fetchFn = wrapWithCredentials(fetch, accessToken, refreshToken);
		const spotifyStatus = await getSpotifyStatus(fetchFn);
		return { spotifyStatus };
	} catch (e) {
		console.error('failed to get spotify status', e);
		return { spotifyStatus: { connected: false } };
	}
};
