import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { getPlaylist } from '$lib/api/playlists';
import { wrapWithRefreshHook } from '@/http';

export const load: PageServerLoad = async ({ params, fetch }) => {
	try {
		const fetchFn = wrapWithRefreshHook(fetch);
		const data = await getPlaylist(params.id, fetchFn);
		return {
			playlist: data,
			tracks: data.tracks
		};
	} catch {
		error(404, 'Playlist not found');
	}
};
