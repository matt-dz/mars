import { getCsrfToken } from '@/http';
import { SpotifyStatusSchema, type SpotifyStatus } from './types';
import { CSRF_HEADER } from '@/auth';
import fetchFn, { type FetchFn } from '@/http';

export async function getSpotifyStatus(fetch: FetchFn = fetchFn): Promise<SpotifyStatus> {
	const res = await fetch.get('api/spotify/status').json();
	return SpotifyStatusSchema.parse(res);
}

export function getSpotifyAuthUrl(): string {
	return '/api/spotify/auth';
}

export async function disconnectSpotify(fetch: FetchFn = fetchFn): Promise<void> {
	// TODO: Replace with real API call when backend is ready
	// const response = await fetchFn('/api/spotify/disconnect', { method: 'POST' });
	// if (!response.ok) throw await response.json();

	void fetch; // Suppress unused variable warning for now
	console.log('Disconnecting Spotify (mock)');
}

export async function getSpotifyTokens(code: string, fetch: FetchFn = fetchFn): Promise<void> {
	const token = getCsrfToken();
	await fetch.post('/api/oauth/spotify/token', {
		headers: {
			'Content-Type': 'application/json',
			[CSRF_HEADER]: token ?? ''
		},
		json: { code }
	});
}
