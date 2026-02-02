<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import TrackList from '$lib/components/app/TrackList.svelte';
	import { addPlaylistToSpotify } from '$lib/api/playlists';
	import { resolve } from '$app/paths';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let isAddingToSpotify = $state(false);
	let message = $state<{ type: 'success' | 'error'; text: string } | null>(null);

	async function handleAddToSpotify() {
		isAddingToSpotify = true;
		message = null;

		try {
			const playlist = await addPlaylistToSpotify(data.playlist.id);
			message = { type: 'success', text: 'Playlist added to Spotify!' };
			window.open(playlist.url, '_blank', 'noopener,noreferrer');
		} catch {
			message = { type: 'error', text: 'Failed to add playlist to Spotify.' };
		} finally {
			isAddingToSpotify = false;
		}
	}

	let formattedDate = $derived(
		new Date(data.playlist.created_at).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		})
	);

	let typeLabel = $derived(
		data.playlist.type === 'weekly'
			? 'Weekly'
			: data.playlist.type === 'monthly'
				? 'Monthly'
				: 'Custom'
	);
</script>

<div class="container mx-auto max-w-4xl px-4 py-8">
	<div class="mb-8">
		<a
			href={resolve('/home')}
			class="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-primary"
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				viewBox="0 0 20 20"
				fill="currentColor"
				class="h-4 w-4"
			>
				<path
					fill-rule="evenodd"
					d="M17 10a.75.75 0 01-.75.75H5.612l4.158 3.96a.75.75 0 11-1.04 1.08l-5.5-5.25a.75.75 0 010-1.08l5.5-5.25a.75.75 0 111.04 1.08L5.612 9.25H16.25A.75.75 0 0117 10z"
					clip-rule="evenodd"
				/>
			</svg>
			Back to playlists
		</a>
	</div>

	<!-- Hero Section -->
	<div
		class="mb-8 overflow-hidden rounded-2xl bg-gradient-to-br from-primary/20 via-accent/15 to-destructive/10 p-6 sm:p-8"
	>
		<div class="flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
			<div class="space-y-3">
				<Badge variant="secondary" class="bg-primary/15 text-primary hover:bg-primary/25">
					{typeLabel}
				</Badge>
				<h1 class="text-3xl font-bold tracking-tight sm:text-4xl">{data.playlist.name}</h1>
				<div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-muted-foreground">
					<span class="flex items-center gap-1.5">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							viewBox="0 0 20 20"
							fill="currentColor"
							class="h-4 w-4"
						>
							<path
								d="M6 3a3 3 0 00-3 3v2.25a3 3 0 003 3h2.25a3 3 0 003-3V6a3 3 0 00-3-3H6zM3.25 14.5a.75.75 0 000 1.5h13.5a.75.75 0 000-1.5H3.25zM3.25 17.5a.75.75 0 000 1.5h13.5a.75.75 0 000-1.5H3.25z"
							/>
						</svg>
						{data.playlist.tracks.length} tracks
					</span>
					<span class="flex items-center gap-1.5">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							viewBox="0 0 20 20"
							fill="currentColor"
							class="h-4 w-4"
						>
							<path
								fill-rule="evenodd"
								d="M5.75 2a.75.75 0 01.75.75V4h7V2.75a.75.75 0 011.5 0V4h.25A2.75 2.75 0 0118 6.75v8.5A2.75 2.75 0 0115.25 18H4.75A2.75 2.75 0 012 15.25v-8.5A2.75 2.75 0 014.75 4H5V2.75A.75.75 0 015.75 2zm-1 5.5c-.69 0-1.25.56-1.25 1.25v6.5c0 .69.56 1.25 1.25 1.25h10.5c.69 0 1.25-.56 1.25-1.25v-6.5c0-.69-.56-1.25-1.25-1.25H4.75z"
								clip-rule="evenodd"
							/>
						</svg>
						{formattedDate}
					</span>
				</div>
			</div>
			<Button
				onclick={handleAddToSpotify}
				disabled={isAddingToSpotify}
				size="lg"
				class="cursor-pointer gap-2"
			>
				{#if isAddingToSpotify}
					<svg
						class="h-4 w-4 animate-spin"
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
					>
						<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"
						></circle>
						<path
							class="opacity-75"
							fill="currentColor"
							d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
						></path>
					</svg>
					Adding...
				{:else}
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 24 24"
						fill="currentColor"
						class="h-5 w-5"
					>
						<path
							d="M12 0C5.4 0 0 5.4 0 12s5.4 12 12 12 12-5.4 12-12S18.66 0 12 0zm5.521 17.34c-.24.359-.66.48-1.021.24-2.82-1.74-6.36-2.101-10.561-1.141-.418.122-.779-.179-.899-.539-.12-.421.18-.78.54-.9 4.56-1.021 8.52-.6 11.64 1.32.42.18.479.659.301 1.02zm1.44-3.3c-.301.42-.841.6-1.262.3-3.239-1.98-8.159-2.58-11.939-1.38-.479.12-1.02-.12-1.14-.6-.12-.48.12-1.021.6-1.141C9.6 9.9 15 10.561 18.72 12.84c.361.181.54.78.241 1.2zm.12-3.36C15.24 8.4 8.82 8.16 5.16 9.301c-.6.179-1.2-.181-1.38-.721-.18-.601.18-1.2.72-1.381 4.26-1.26 11.28-1.02 15.721 1.621.539.3.719 1.02.419 1.56-.299.421-1.02.599-1.559.3z"
						/>
					</svg>
					Add to Spotify
				{/if}
			</Button>
		</div>
	</div>

	{#if message}
		<div
			class="mb-6 flex items-center gap-2 rounded-lg p-4 text-sm {message.type === 'success'
				? 'bg-green-500/10 text-green-700 dark:text-green-400'
				: 'bg-destructive/10 text-destructive'}"
		>
			{#if message.type === 'success'}
				<svg
					xmlns="http://www.w3.org/2000/svg"
					viewBox="0 0 20 20"
					fill="currentColor"
					class="h-5 w-5"
				>
					<path
						fill-rule="evenodd"
						d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z"
						clip-rule="evenodd"
					/>
				</svg>
			{:else}
				<svg
					xmlns="http://www.w3.org/2000/svg"
					viewBox="0 0 20 20"
					fill="currentColor"
					class="h-5 w-5"
				>
					<path
						fill-rule="evenodd"
						d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-5a.75.75 0 01.75.75v4.5a.75.75 0 01-1.5 0v-4.5A.75.75 0 0110 5zm0 10a1 1 0 100-2 1 1 0 000 2z"
						clip-rule="evenodd"
					/>
				</svg>
			{/if}
			{message.text}
		</div>
	{/if}

	<TrackList tracks={data.playlist.tracks} />
</div>
