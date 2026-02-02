import fetchFn from '@/http';
import * as z from 'zod';

function randomBase64Url(bytes: number) {
	const arr = new Uint8Array(bytes);
	crypto.getRandomValues(arr);

	let binary = '';
	for (let i = 0; i < arr.length; i++) {
		binary += String.fromCharCode(arr[i]);
	}

	return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

function createAndStoreState(): string {
	const state = randomBase64Url(32);
	localStorage.setItem('state', state);
	return state;
}

function getState(): string | null {
	return localStorage.getItem('state');
}

const OAuthSchema = z.object({
	response_type: z.literal('code'),
	client_id: z.string(),
	redirect_uri: z.string(),
	scope: z.string()
});

// TODO: refactor this to pull from backend
export async function getSpotifyAuthUrl() {
	// Fetch config
	const res = await fetchFn('/api/oauth/spotify/config.json');
	const schema = OAuthSchema.parse(await res.json());

	// Redirect to auth
	const state = createAndStoreState();
	const baseUrl = 'https://accounts.spotify.com/authorize';
	const params = new URLSearchParams({
		response_type: schema.response_type,
		client_id: schema.client_id,
		redirect_uri: schema.redirect_uri,
		scope: schema.scope,
		state
	});
	window.location.href = `${baseUrl}?${params.toString()}`;
}

export function stateMatches(state: string): boolean {
	try {
		const groundState = getState();
		if (groundState === null) return false;
		return state === groundState;
	} catch (e) {
		console.error(e);
		return false;
	}
}
