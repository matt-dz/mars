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

<svelte:head>
	<title>Integrations - Mars</title>
</svelte:head>

<div class="container mx-auto max-w-2xl px-4 py-8">
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

	<div class="mb-8 space-y-1">
		<h1 class="text-3xl font-bold tracking-tight">Integrations</h1>
		<p class="text-muted-foreground">Connect your music services to enhance your experience.</p>
	</div>

	<Card.Root class="overflow-hidden">
		<div class="h-1.5 bg-gradient-to-r from-[#1DB954] via-[#1DB954]/70 to-primary/30"></div>
		<Card.Header class="pt-6">
			<div class="flex items-start gap-4">
				<div class="flex h-12 w-12 items-center justify-center rounded-xl bg-[#1DB954]/10">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 24 24"
						fill="#1DB954"
						class="h-7 w-7"
					>
						<path
							d="M12 0C5.4 0 0 5.4 0 12s5.4 12 12 12 12-5.4 12-12S18.66 0 12 0zm5.521 17.34c-.24.359-.66.48-1.021.24-2.82-1.74-6.36-2.101-10.561-1.141-.418.122-.779-.179-.899-.539-.12-.421.18-.78.54-.9 4.56-1.021 8.52-.6 11.64 1.32.42.18.479.659.301 1.02zm1.44-3.3c-.301.42-.841.6-1.262.3-3.239-1.98-8.159-2.58-11.939-1.38-.479.12-1.02-.12-1.14-.6-.12-.48.12-1.021.6-1.141C9.6 9.9 15 10.561 18.72 12.84c.361.181.54.78.241 1.2zm.12-3.36C15.24 8.4 8.82 8.16 5.16 9.301c-.6.179-1.2-.181-1.38-.721-.18-.601.18-1.2.72-1.381 4.26-1.26 11.28-1.02 15.721 1.621.539.3.719 1.02.419 1.56-.299.421-1.02.599-1.559.3z"
						/>
					</svg>
				</div>
				<div class="flex-1 space-y-1">
					<Card.Title class="text-xl">Spotify</Card.Title>
					<Card.Description>
						Connect your Spotify account to sync your listening history and create playlists.
					</Card.Description>
				</div>
			</div>
		</Card.Header>
		<Card.Content>
			<SpotifyStatus status={data.spotifyStatus} />
		</Card.Content>
		<Card.Footer class="border-t bg-muted/30 pt-4">
			{#if data.spotifyStatus.connected}
				<div class="flex gap-2">
					<Button variant="outline" onclick={getSpotifyAuthUrl} class="gap-2">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							viewBox="0 0 20 20"
							fill="currentColor"
							class="h-4 w-4"
						>
							<path
								fill-rule="evenodd"
								d="M15.312 11.424a5.5 5.5 0 01-9.201 2.466l-.312-.311h2.433a.75.75 0 000-1.5H3.989a.75.75 0 00-.75.75v4.242a.75.75 0 001.5 0v-2.43l.31.31a7 7 0 0011.712-3.138.75.75 0 00-1.449-.39zm1.23-3.723a.75.75 0 00.219-.53V2.929a.75.75 0 00-1.5 0V5.36l-.31-.31A7 7 0 003.239 8.188a.75.75 0 101.448.389A5.5 5.5 0 0113.89 6.11l.311.31h-2.432a.75.75 0 000 1.5h4.243a.75.75 0 00.53-.219z"
								clip-rule="evenodd"
							/>
						</svg>
						Reconnect
					</Button>
					<Button variant="destructive" onclick={() => (showDisconnectDialog = true)} class="gap-2">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							viewBox="0 0 20 20"
							fill="currentColor"
							class="h-4 w-4"
						>
							<path
								fill-rule="evenodd"
								d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z"
								clip-rule="evenodd"
							/>
						</svg>
						Disconnect
					</Button>
				</div>
			{:else}
				<Button onclick={getSpotifyAuthUrl} class="gap-2 bg-[#1DB954] hover:bg-[#1DB954]/90">
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
					Connect Spotify
				</Button>
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
