import { isHTTPError } from 'ky';
import { ApiErrorSchema, TopTracksSchema, type TopTracks } from './types';
import fetchFn, { type FetchFn } from '$lib/http';
import { HTTPError } from './errors';

export async function getTopTracks(
	startDate: Date,
	endDate: Date,
	fetch: FetchFn = fetchFn
): Promise<TopTracks> {
	try {
		const params = new URLSearchParams({
			start: `${Math.floor(startDate.getTime() / 1000)}`,
			end: `${Math.floor(endDate.getTime() / 1000)}`
		});
		const res = await fetch.get(`api/me/tracks/top?${params.toString()}`).json();
		return TopTracksSchema.parse(res);
	} catch (e) {
		if (isHTTPError(e)) {
			const err = ApiErrorSchema.safeParse(await e.response.clone().json());
			if (err.success) {
				throw new HTTPError(err.data.status, err.data.message, err.data.code, err.data.error_id);
			}
			throw new HTTPError(e.response.status, await e.response.text());
		}
		throw e;
	}
}
