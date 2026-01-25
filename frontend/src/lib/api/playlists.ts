import { PlaylistSchema, PlaylistWithTracksSchema, type Playlist, type PlaylistWithTracks } from './types';

const MOCK_PLAYLISTS: Playlist[] = [
	{
		id: '1',
		user_id: 'user-1',
		playlist_type: 'weekly',
		name: 'Week of Jan 20, 2026',
		timestamp: '2026-01-20T00:00:00Z',
		created_at: '2026-01-20T00:00:00Z'
	},
	{
		id: '2',
		user_id: 'user-1',
		playlist_type: 'weekly',
		name: 'Week of Jan 13, 2026',
		timestamp: '2026-01-13T00:00:00Z',
		created_at: '2026-01-13T00:00:00Z'
	},
	{
		id: '3',
		user_id: 'user-1',
		playlist_type: 'monthly',
		name: 'December 2025',
		timestamp: '2025-12-01T00:00:00Z',
		created_at: '2025-12-31T00:00:00Z'
	},
	{
		id: '4',
		user_id: 'user-1',
		playlist_type: 'monthly',
		name: 'November 2025',
		timestamp: '2025-11-01T00:00:00Z',
		created_at: '2025-11-30T00:00:00Z'
	},
	{
		id: '5',
		user_id: 'user-1',
		playlist_type: 'weekly',
		name: 'Week of Jan 6, 2026',
		timestamp: '2026-01-06T00:00:00Z',
		created_at: '2026-01-06T00:00:00Z'
	}
];

const MOCK_TRACKS = [
	{
		id: 'track-1',
		name: 'Blinding Lights',
		artists: ['The Weeknd'],
		href: 'https://open.spotify.com/track/0VjIjW4GlUZAMYd2vXMi3b',
		image_url: 'https://i.scdn.co/image/ab67616d0000b2738863bc11d2aa12b54f5aeb36'
	},
	{
		id: 'track-2',
		name: 'Levitating',
		artists: ['Dua Lipa', 'DaBaby'],
		href: 'https://open.spotify.com/track/5nujrmhLynf4yMoMtj8AQF',
		image_url: 'https://i.scdn.co/image/ab67616d0000b273bd26ede1ae69327010d49946'
	},
	{
		id: 'track-3',
		name: 'Stay',
		artists: ['The Kid LAROI', 'Justin Bieber'],
		href: 'https://open.spotify.com/track/5PjdY0CKGZdEuoNab3yDmX',
		image_url: 'https://i.scdn.co/image/ab67616d0000b273a91c10fe9472d9bd89802e5a'
	},
	{
		id: 'track-4',
		name: 'Heat Waves',
		artists: ['Glass Animals'],
		href: 'https://open.spotify.com/track/02MWAaffLxlfxAUY7c5dvx',
		image_url: 'https://i.scdn.co/image/ab67616d0000b273712701c5e263efc8726b1464'
	},
	{
		id: 'track-5',
		name: 'good 4 u',
		artists: ['Olivia Rodrigo'],
		href: 'https://open.spotify.com/track/4ZtFanR9U6ndgddUvNcjcG',
		image_url: 'https://i.scdn.co/image/ab67616d0000b273a91c10fe9472d9bd89802e5a'
	}
];

type FetchFn = typeof fetch;

export async function getPlaylists(fetchFn: FetchFn = fetch): Promise<Playlist[]> {
	// TODO: Replace with real API call when backend is ready
	// const response = await fetchFn('/api/playlists');
	// if (!response.ok) throw await response.json();
	// const data = await response.json();
	// return z.array(PlaylistSchema).parse(data);

	void fetchFn; // Suppress unused variable warning for now
	return MOCK_PLAYLISTS.map((p) => PlaylistSchema.parse(p));
}

export async function getPlaylist(id: string, fetchFn: FetchFn = fetch): Promise<PlaylistWithTracks> {
	// TODO: Replace with real API call when backend is ready
	// const response = await fetchFn(`/api/playlists/${id}`);
	// if (!response.ok) throw await response.json();
	// const data = await response.json();
	// return PlaylistWithTracksSchema.parse(data);

	void fetchFn; // Suppress unused variable warning for now
	const playlist = MOCK_PLAYLISTS.find((p) => p.id === id);
	if (!playlist) {
		throw { code: 'not_found', message: 'Playlist not found', status: 404 };
	}

	return PlaylistWithTracksSchema.parse({
		...playlist,
		tracks: MOCK_TRACKS
	});
}

export async function addPlaylistToSpotify(id: string, fetchFn: FetchFn = fetch): Promise<void> {
	// TODO: Replace with real API call when backend is ready
	// const response = await fetchFn(`/api/playlists/${id}/spotify`, { method: 'POST' });
	// if (!response.ok) throw await response.json();

	void fetchFn; // Suppress unused variable warning for now
	console.log(`Adding playlist ${id} to Spotify (mock)`);
}
