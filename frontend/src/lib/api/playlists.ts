import { isHTTPError } from 'ky';
import {
	ApiErrorSchema,
	PlaylistsSchema,
	PlaylistWithTracksSchema,
	SpotifyPlaylistSchema,
	type Playlists,
	type PlaylistWithTracks,
	type SpotifyPlaylist
} from './types';
import fetchFn, { type FetchFn } from '@/http';
import { HTTPError } from './errors';

export async function getPlaylists(fetch: FetchFn): Promise<Playlists> {
	try {
		const res = await fetch.get('api/me/playlists').json();
		return PlaylistsSchema.parse(res);
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

export async function getPlaylist(
	id: string,
	fetch: FetchFn = fetchFn
): Promise<PlaylistWithTracks> {
	try {
		const res = await fetch.get(`api/playlists/${id}`).json();
		return PlaylistWithTracksSchema.parse(res);
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

export async function addPlaylistToSpotify(
	id: string,
	fetch: FetchFn = fetchFn
): Promise<SpotifyPlaylist> {
	try {
		const res = await fetch.post(`/api/integrations/spotify/playlist/${id}`).json();
		return SpotifyPlaylistSchema.parse(res);
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
