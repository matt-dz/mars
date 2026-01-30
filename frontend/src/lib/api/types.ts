import * as z from 'zod';

export const UserSchema = z.object({
	id: z.string(),
	email: z.string().email(),
	role: z.enum(['admin', 'user'])
});

export type User = z.infer<typeof UserSchema>;

export const PlaylistSchema = z.object({
	id: z.string(),
	user_id: z.string(),
	playlist_type: z.enum(['weekly', 'monthly']),
	name: z.string(),
	timestamp: z.string(),
	created_at: z.string()
});

export type Playlist = z.infer<typeof PlaylistSchema>;

export const TrackSchema = z.object({
	id: z.string(),
	name: z.string(),
	artists: z.array(z.string()),
	href: z.string(),
	image_url: z.string().nullable()
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

export const ApiErrorSchema = z.object({
	code: z.string(),
	error_id: z.number(),
	message: z.string(),
	status: z.number()
});

export type ApiError = z.infer<typeof ApiErrorSchema>;
