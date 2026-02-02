import type { PageServerLoad } from './$types';
import { extractAuthCookies } from '@/http';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ cookies }) => {
	const { accessToken } = extractAuthCookies(cookies);
	if (accessToken) {
		console.debug('[login] user authorized, sending them to home page');
		return redirect(302, '/home');
	}
};
