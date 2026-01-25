import type { PageServerLoad } from './$types';
import { getPlaylists } from '$lib/api/playlists';

export const load: PageServerLoad = async ({ fetch }) => {
	const playlists = await getPlaylists(fetch);
	return { playlists };
};
