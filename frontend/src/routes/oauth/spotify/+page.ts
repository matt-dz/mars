import type { PageLoad } from './$types';
import { browser } from '$app/environment';
import { stateMatches } from '$lib/oauth';
import { getSpotifyTokens } from '@/api';
import { resolve } from '$app/paths';
import { wrap } from '@/http';
import { redirect } from '@sveltejs/kit';

export const load: PageLoad = async ({ url, fetch }) => {
	if (!browser) return;

	const error = url.searchParams.get('error');
	if (error) {
		console.error('user denied request', error);
		return;
	}

	const code = url.searchParams.get('code');
	if (code === null) {
		console.error('code missing');
		return;
	}

	const state = url.searchParams.get('state');
	if (state === null || !stateMatches(state)) {
		console.error('state mismatch');
		return;
	}

	try {
		await getSpotifyTokens(code, wrap(fetch));
	} catch (e) {
		console.error(e);
	} finally {
		redirect(303, resolve('/integrations'));
	}
};

export const ssr = false;
