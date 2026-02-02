import { env } from '$env/dynamic/public';

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

// TODO: refactor this to pull from backend
export async function getSpotifyAuthUrl() {
	const state = createAndStoreState();
	const baseUrl = 'https://accounts.spotify.com/authorize';
	const params = new URLSearchParams({
		response_type: 'code',
		client_id: env.PUBLIC_SPOTIFY_CLIENT_ID ?? '',
		redirect_uri: env.PUBLIC_SPOTIFY_REDIRECT_URI ?? '',
		scope:
			'user-read-private user-read-email user-library-read user-top-read user-read-recently-played playlist-modify-public playlist-modify-private ugc-image-upload',
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
