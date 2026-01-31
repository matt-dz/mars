import { type FetchFn } from '@/http';
import { isHTTPError } from 'ky';
import { HTTPError } from './errors';

export async function verifySession(fetch: FetchFn) {
	try {
		return await fetch.get('api/auth/verify');
	} catch (e) {
		if (isHTTPError(e)) {
			throw new HTTPError(e.response.status, await e.response.text());
		}
		throw e;
	}
}
