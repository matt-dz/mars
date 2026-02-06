<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import * as Alert from '$lib/components/ui/alert';
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

	let accentColor = $derived(
		data.playlist.type === 'weekly'
			? 'text-primary'
			: data.playlist.type === 'monthly'
				? 'text-destructive'
				: 'text-primary'
	);

	let gradientClass = $derived(
		data.playlist.type === 'weekly'
			? 'from-primary/20 via-primary/10 to-transparent'
			: data.playlist.type === 'monthly'
				? 'from-destructive/20 via-destructive/10 to-transparent'
				: 'from-primary/20 via-destructive/15 to-transparent'
	);
</script>

<svelte:head>
	<title>{data.playlist.name} - Mars</title>
</svelte:head>

<div class="container mx-auto max-w-4xl px-4 py-8">
	<!-- Back button -->
	<div class="mb-8">
		<a
			href={resolve('/home')}
			class="group inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-primary"
		>
			<div
				class="flex h-7 w-7 items-center justify-center rounded-full bg-muted/50 transition-colors group-hover:bg-primary/10"
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
			</div>
			<span class="font-medium">Back to playlists</span>
		</a>
	</div>

	<!-- Hero Section -->
	<div class="relative mb-10 overflow-hidden rounded-3xl border border-border/50 shadow-xl">
		<!-- Background decoration -->
		<div
			class="pointer-events-none absolute inset-0 bg-gradient-to-br opacity-50 {gradientClass}"
		></div>

		<!-- Decorative blurred circles -->
		<div
			class="pointer-events-none absolute -top-20 -right-20 h-40 w-40 rounded-full bg-gradient-to-br from-primary/30 to-destructive/20 blur-3xl"
		></div>
		<div
			class="pointer-events-none absolute -bottom-16 -left-16 h-32 w-32 rounded-full bg-gradient-to-tr from-destructive/30 to-primary/20 blur-3xl"
		></div>

		<!-- Content -->
		<div class="relative p-8 sm:p-10">
			<div class="mb-8 flex flex-col gap-6 sm:flex-row sm:items-start sm:justify-between">
				<!-- Left side - Info -->
				<div class="flex-1 space-y-6">
					<!-- Badge and type icon -->
					<div class="flex items-center gap-3">
						<div
							class="flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-primary/30 to-destructive/25 shadow-lg shadow-primary/10"
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								viewBox="0 0 24 24"
								fill="currentColor"
								class="h-6 w-6 {accentColor}"
							>
								<path d="M9 18V5l12-2v13" />
								<circle cx="6" cy="18" r="3" />
								<circle cx="18" cy="16" r="3" />
							</svg>
						</div>
						<Badge
							variant="secondary"
							class="border border-border/50 bg-background/80 px-4 py-1 text-sm backdrop-blur-sm"
						>
							{typeLabel}
						</Badge>
					</div>

					<!-- Title -->
					<div class="space-y-3">
						<h1
							class="bg-gradient-to-br from-foreground to-foreground/70 bg-clip-text text-4xl leading-tight font-bold tracking-tight text-transparent sm:text-5xl"
						>
							{data.playlist.name}
						</h1>

						<!-- Meta info -->
						<div class="flex flex-wrap items-center gap-x-6 gap-y-2 text-sm text-muted-foreground">
							<span class="flex items-center gap-2">
								<div class="rounded-lg bg-primary/10 p-1.5">
									<svg
										xmlns="http://www.w3.org/2000/svg"
										viewBox="0 0 20 20"
										fill="currentColor"
										class="h-4 w-4 {accentColor}"
									>
										<path
											d="M6 3a3 3 0 00-3 3v2.25a3 3 0 003 3h2.25a3 3 0 003-3V6a3 3 0 00-3-3H6zM3.25 14.5a.75.75 0 000 1.5h13.5a.75.75 0 000-1.5H3.25zM3.25 17.5a.75.75 0 000 1.5h13.5a.75.75 0 000-1.5H3.25z"
										/>
									</svg>
								</div>
								<span class="font-medium"
									>{data.playlist.tracks.length} track{data.playlist.tracks.length === 1
										? ''
										: 's'}</span
								>
							</span>
							<span class="flex items-center gap-2">
								<div class="rounded-lg bg-primary/10 p-1.5">
									<svg
										xmlns="http://www.w3.org/2000/svg"
										viewBox="0 0 20 20"
										fill="currentColor"
										class="h-4 w-4 {accentColor}"
									>
										<path
											fill-rule="evenodd"
											d="M5.75 2a.75.75 0 01.75.75V4h7V2.75a.75.75 0 011.5 0V4h.25A2.75 2.75 0 0118 6.75v8.5A2.75 2.75 0 0115.25 18H4.75A2.75 2.75 0 012 15.25v-8.5A2.75 2.75 0 014.75 4H5V2.75A.75.75 0 015.75 2zm-1 5.5c-.69 0-1.25.56-1.25 1.25v6.5c0 .69.56 1.25 1.25 1.25h10.5c.69 0 1.25-.56 1.25-1.25v-6.5c0-.69-.56-1.25-1.25-1.25H4.75z"
											clip-rule="evenodd"
										/>
									</svg>
								</div>
								<span>{formattedDate}</span>
							</span>
						</div>
					</div>
				</div>

				<!-- Right side - Actions -->
				<div class="flex items-start sm:flex-col sm:items-end">
					<Button
						onclick={handleAddToSpotify}
						disabled={isAddingToSpotify}
						size="lg"
						class="group relative cursor-pointer gap-2 overflow-hidden shadow-lg transition-all hover:shadow-xl hover:shadow-primary/20"
					>
						<!-- Shimmer effect on hover -->
						<div
							class="pointer-events-none absolute inset-0 -translate-x-full bg-gradient-to-r from-transparent via-white/20 to-transparent transition-transform duration-500 group-hover:translate-x-full"
						></div>

						{#if isAddingToSpotify}
							<svg
								class="h-5 w-5 animate-spin"
								xmlns="http://www.w3.org/2000/svg"
								fill="none"
								viewBox="0 0 24 24"
							>
								<circle
									class="opacity-25"
									cx="12"
									cy="12"
									r="10"
									stroke="currentColor"
									stroke-width="4"
								></circle>
								<path
									class="opacity-75"
									fill="currentColor"
									d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
								></path>
							</svg>
							<span>Adding...</span>
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
							<span>Add to Spotify</span>
						{/if}
					</Button>
				</div>
			</div>

			<!-- Accent line -->
			<div
				class="h-1 w-full rounded-full bg-gradient-to-r from-primary/50 via-primary/30 to-destructive/30"
			></div>
		</div>
	</div>

	<!-- Success/Error Messages -->
	{#if message}
		<div class="mb-8">
			{#if message.type === 'success'}
				<Alert.Root class="border-green-500/50 bg-green-500/10">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 20 20"
						fill="currentColor"
						class="h-5 w-5 text-green-600 dark:text-green-400"
					>
						<path
							fill-rule="evenodd"
							d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z"
							clip-rule="evenodd"
						/>
					</svg>
					<Alert.Title class="text-green-700 dark:text-green-400">Success!</Alert.Title>
					<Alert.Description class="text-green-700 dark:text-green-400"
						>{message.text}</Alert.Description
					>
				</Alert.Root>
			{:else}
				<Alert.Root variant="destructive">
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
					<Alert.Title>Error</Alert.Title>
					<Alert.Description>{message.text}</Alert.Description>
				</Alert.Root>
			{/if}
		</div>
	{/if}

	<!-- Track List -->
	<div class="space-y-4">
		<div class="flex items-center gap-3">
			<div class="h-px flex-1 bg-gradient-to-r from-transparent via-border to-transparent"></div>
			<h2 class="text-sm font-medium text-muted-foreground">Tracks</h2>
			<div class="h-px flex-1 bg-gradient-to-r from-transparent via-border to-transparent"></div>
		</div>
		<TrackList tracks={data.playlist.tracks} />
	</div>
</div>
