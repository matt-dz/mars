import { SpotifyStatusSchema, type SpotifyStatus } from './types';

type FetchFn = typeof fetch;

export async function getSpotifyStatus(fetchFn: FetchFn = fetch): Promise<SpotifyStatus> {
	// TODO: Replace with real API call when backend is ready
	// const response = await fetchFn('/api/spotify/status');
	// if (!response.ok) throw await response.json();
	// const data = await response.json();
	// return SpotifyStatusSchema.parse(data);

	void fetchFn; // Suppress unused variable warning for now
	return SpotifyStatusSchema.parse({
		connected: false,
		spotify_user_id: null,
		token_expires: null
	});
}

export function getSpotifyAuthUrl(): string {
	return '/api/spotify/auth';
}

export async function disconnectSpotify(fetchFn: FetchFn = fetch): Promise<void> {
	// TODO: Replace with real API call when backend is ready
	// const response = await fetchFn('/api/spotify/disconnect', { method: 'POST' });
	// if (!response.ok) throw await response.json();

	void fetchFn; // Suppress unused variable warning for now
	console.log('Disconnecting Spotify (mock)');
}
