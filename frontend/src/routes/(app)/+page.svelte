<script lang="ts">
	import PlaylistCard from '$lib/components/app/PlaylistCard.svelte';
	import PlaylistFilters from '$lib/components/app/PlaylistFilters.svelte';
	import type { PageData } from './$types';
	import type { Playlist } from '$lib/api/types';

	let { data }: { data: PageData } = $props();

	let period = $state<'week' | 'month' | 'year' | 'all'>('all');
	let sortOrder = $state<'asc' | 'desc'>('desc');

	function filterByPeriod(playlists: Playlist[], period: string): Playlist[] {
		if (period === 'all') return playlists;

		const now = new Date();
		const cutoff = new Date();

		if (period === 'week') {
			cutoff.setDate(now.getDate() - 7);
		} else if (period === 'month') {
			cutoff.setMonth(now.getMonth() - 1);
		} else if (period === 'year') {
			cutoff.setFullYear(now.getFullYear() - 1);
		}

		return playlists.filter((p) => new Date(p.timestamp) >= cutoff);
	}

	function sortPlaylists(playlists: Playlist[], order: string): Playlist[] {
		return [...playlists].sort((a, b) => {
			const dateA = new Date(a.timestamp).getTime();
			const dateB = new Date(b.timestamp).getTime();
			return order === 'desc' ? dateB - dateA : dateA - dateB;
		});
	}

	let filteredPlaylists = $derived(
		sortPlaylists(filterByPeriod(data.playlists, period), sortOrder)
	);
</script>

<div class="container mx-auto px-4 py-8">
	<div class="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<h1 class="text-3xl font-bold">Your Playlists</h1>
		<PlaylistFilters bind:period bind:sortOrder />
	</div>

	{#if filteredPlaylists.length > 0}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each filteredPlaylists as playlist (playlist.id)}
				<PlaylistCard {playlist} />
			{/each}
		</div>
	{:else}
		<div class="rounded-lg border border-dashed p-12 text-center">
			<p class="text-muted-foreground">No playlists found for the selected period.</p>
		</div>
	{/if}
</div>
