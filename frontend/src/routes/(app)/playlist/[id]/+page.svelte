<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import TrackList from '$lib/components/app/TrackList.svelte';
	import { addPlaylistToSpotify } from '$lib/api/playlists';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let isAddingToSpotify = $state(false);
	let message = $state<{ type: 'success' | 'error'; text: string } | null>(null);

	async function handleAddToSpotify() {
		isAddingToSpotify = true;
		message = null;

		try {
			await addPlaylistToSpotify(data.playlist.id);
			message = { type: 'success', text: 'Playlist added to Spotify!' };
		} catch {
			message = { type: 'error', text: 'Failed to add playlist to Spotify.' };
		} finally {
			isAddingToSpotify = false;
		}
	}

	let formattedDate = $derived(
		new Date(data.playlist.timestamp).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		})
	);
</script>

<div class="container mx-auto px-4 py-8">
	<div class="mb-6">
		<a href="/" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to playlists</a>
	</div>

	<div class="mb-6 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
		<div>
			<h1 class="text-3xl font-bold">{data.playlist.name}</h1>
			<p class="mt-1 text-muted-foreground">
				{data.tracks.length} tracks &middot; {formattedDate}
			</p>
		</div>
		<Button onclick={handleAddToSpotify} disabled={isAddingToSpotify}>
			{isAddingToSpotify ? 'Adding...' : 'Add to Spotify'}
		</Button>
	</div>

	{#if message}
		<div
			class="mb-6 rounded-md p-3 text-sm {message.type === 'success'
				? 'bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400'
				: 'bg-destructive/10 text-destructive'}"
		>
			{message.text}
		</div>
	{/if}

	<TrackList tracks={data.tracks} />
</div>
