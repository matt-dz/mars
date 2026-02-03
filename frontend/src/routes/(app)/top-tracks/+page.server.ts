import type { PageServerLoad } from './$types';
import { getTopTracks } from '$lib/api/tracks';
import { extractAuthCookies, wrapWithCredentials } from '$lib/http';
import { redirect } from '@sveltejs/kit';
import { HTTPError } from '$lib/api/errors';

export const load: PageServerLoad = async ({ fetch, cookies, url }) => {
	try {
		const tokens = extractAuthCookies(cookies);
		if (!tokens.accessToken || !tokens.refreshToken || !tokens.csrfToken) {
			console.debug('[top-tracks] Missing auth tokens; redirecting to login');
			redirect(302, '/login');
		}

		// Get time period from URL params, default to past 24 hours
		const period = url.searchParams.get('period') || 'day';
		const customStart = url.searchParams.get('start');
		const customEnd = url.searchParams.get('end');

		const now = new Date();
		let startDate: Date;
		let endDate = now;

		if (period === 'custom' && customStart && customEnd) {
			startDate = new Date(customStart);
			endDate = new Date(customEnd);
		} else {
			switch (period) {
				case 'week':
					startDate = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
					break;
				case 'month-to-date':
					startDate = new Date(now.getFullYear(), now.getMonth(), 1);
					break;
				case 'year-to-date':
					startDate = new Date(now.getFullYear(), 0, 1);
					break;
				case 'day':
				default:
					startDate = new Date(now.getTime() - 24 * 60 * 60 * 1000);
					break;
			}
		}

		console.debug('[top-tracks] Fetching top tracks');
		const fetchFn = wrapWithCredentials(
			fetch,
			tokens.accessToken,
			tokens.refreshToken,
			tokens.csrfToken
		);

		const topTracks = await getTopTracks(startDate, endDate, fetchFn);

		return {
			topTracks,
			period,
			startDate: startDate.toISOString(),
			endDate: endDate.toISOString(),
			error: null
		};
	} catch (e) {
		console.error('[top-tracks] Failed to fetch top tracks:', e);
		if (e instanceof HTTPError) {
			if (e.statusCode === 401) {
				console.debug('[top-tracks] Unauthorized; redirecting to login');
				redirect(302, '/login');
			}
			if (e.statusCode === 400) {
				console.debug('[top-tracks] Bad request; returning error to client');
				const period = url.searchParams.get('period') || 'day';
				const customStart = url.searchParams.get('start');
				const customEnd = url.searchParams.get('end');
				const now = new Date();

				return {
					topTracks: { tracks: [] },
					period,
					startDate: customStart || now.toISOString(),
					endDate: customEnd || now.toISOString(),
					error: e.message || 'Invalid time frame selected. Please check your dates and try again.'
				};
			}
		}
		throw e;
	}
};
