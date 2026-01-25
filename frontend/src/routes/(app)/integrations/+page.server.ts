import type { PageServerLoad } from './$types';
import { getSpotifyStatus } from '$lib/api/spotify';

export const load: PageServerLoad = async ({ fetch }) => {
	const spotifyStatus = await getSpotifyStatus(fetch);
	return { spotifyStatus };
};
