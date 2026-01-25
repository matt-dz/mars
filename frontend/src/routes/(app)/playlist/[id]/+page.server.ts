import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { getPlaylist } from '$lib/api/playlists';

export const load: PageServerLoad = async ({ params, fetch }) => {
	try {
		const data = await getPlaylist(params.id, fetch);
		return {
			playlist: data,
			tracks: data.tracks
		};
	} catch {
		error(404, 'Playlist not found');
	}
};
