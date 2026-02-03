import * as z from 'zod';

export const UserSchema = z.object({
	id: z.string(),
	email: z.email(),
	role: z.enum(['admin', 'user'])
});

export type User = z.infer<typeof UserSchema>;

export const PlaylistSchema = z.object({
	id: z.string(),
	type: z.enum(['weekly', 'monthly', 'custom']),
	name: z.string(),
	created_at: z.iso.datetime()
});

export type Playlist = z.infer<typeof PlaylistSchema>;

export const PlaylistsSchema = z.object({
	playlists: z.array(PlaylistSchema)
});

export type Playlists = z.infer<typeof PlaylistsSchema>;

export const TrackSchema = z.object({
	id: z.string(),
	name: z.string(),
	artists: z.array(z.string()),
	href: z.string(),
	image_url: z.string().optional(),
	plays: z.int()
});

export type Track = z.infer<typeof TrackSchema>;

export const PlaylistWithTracksSchema = PlaylistSchema.extend({
	tracks: z.array(TrackSchema)
});

export type PlaylistWithTracks = z.infer<typeof PlaylistWithTracksSchema>;

export const SpotifyStatusSchema = z.object({
	connected: z.boolean()
});

export type SpotifyStatus = z.infer<typeof SpotifyStatusSchema>;

// Re-export error types from errors.ts for backwards compatibility
export { ApiErrorSchema, type ApiError } from './errors';

export const SpotifyPlaylistSchema = z.object({
	id: z.string(),
	url: z.string()
});

export type SpotifyPlaylist = z.infer<typeof SpotifyPlaylistSchema>;

export const TopTracksSchema = z.object({
	tracks: z.array(TrackSchema)
});

export type TopTracks = z.infer<typeof TopTracksSchema>;
