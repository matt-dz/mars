import type { PageServerLoad } from './$types';
import { getPlaylists } from '$lib/api/playlists';
import { wrap } from '@/http';

export const load: PageServerLoad = async ({ fetch }) => {
	const fetchFn = wrap(fetch);
	const playlists = await getPlaylists(fetchFn);
	return { playlists };
};
