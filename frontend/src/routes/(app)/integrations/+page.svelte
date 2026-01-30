<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import SpotifyStatus from '$lib/components/app/SpotifyStatus.svelte';
	import { disconnectSpotify } from '$lib/api/spotify';
	import { getSpotifyAuthUrl } from '@/oauth';
	import { invalidateAll } from '$app/navigation';
	import type { PageData } from './$types';
	import { resolve } from '$app/paths';

	let { data }: { data: PageData } = $props();

	let showDisconnectDialog = $state(false);
	let isDisconnecting = $state(false);

	async function handleDisconnect() {
		isDisconnecting = true;
		try {
			await disconnectSpotify();
			showDisconnectDialog = false;
			await invalidateAll();
		} catch (err) {
			console.error('Failed to disconnect:', err);
		} finally {
			isDisconnecting = false;
		}
	}
</script>

<div class="container mx-auto px-4 py-8">
	<div class="mb-6">
		<a href={resolve('/')} class="text-sm text-muted-foreground hover:text-foreground"
			>&larr; Back to playlists</a
		>
	</div>

	<h1 class="mb-6 text-3xl font-bold">Integrations</h1>

	<Card.Root class="max-w-md">
		<Card.Header>
			<Card.Title>Spotify</Card.Title>
			<Card.Description>
				Connect your Spotify account to sync your listening history and create playlists.
			</Card.Description>
		</Card.Header>
		<Card.Content>
			<SpotifyStatus status={data.spotifyStatus} />
		</Card.Content>
		<Card.Footer>
			{#if data.spotifyStatus.connected}
				<div class="flex gap-2">
					<Button variant="outline" onclick={getSpotifyAuthUrl}>Reconnect</Button>
					<Button variant="destructive" onclick={() => (showDisconnectDialog = true)}>
						Disconnect
					</Button>
				</div>
			{:else}
				<Button onclick={getSpotifyAuthUrl}>Connect Spotify</Button>
			{/if}
		</Card.Footer>
	</Card.Root>
</div>

<Dialog.Root bind:open={showDisconnectDialog}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Disconnect Spotify?</Dialog.Title>
			<Dialog.Description>
				This will stop syncing your listening history. You can reconnect anytime.
			</Dialog.Description>
		</Dialog.Header>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (showDisconnectDialog = false)}>Cancel</Button>
			<Button variant="destructive" onclick={handleDisconnect} disabled={isDisconnecting}>
				{isDisconnecting ? 'Disconnecting...' : 'Disconnect'}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
