<script lang="ts">
	import PlaylistCard from '$lib/components/app/PlaylistCard.svelte';
	import PlaylistFilters from '$lib/components/app/PlaylistFilters.svelte';
	import type { PageData } from './$types';
	import type { Playlist } from '$lib/api/types';
	import { SvelteDate } from 'svelte/reactivity';

	let { data }: { data: PageData } = $props();

	let period = $state<'week' | 'month' | 'year' | 'all'>('all');
	let sortOrder = $state<'asc' | 'desc'>('desc');

	function filterByPeriod(playlists: Playlist[], period: string): Playlist[] {
		if (period === 'all') return playlists;

		const now = new Date();
		const cutoff = new SvelteDate();

		if (period === 'week') {
			cutoff.setDate(now.getDate() - 7);
		} else if (period === 'month') {
			cutoff.setMonth(now.getMonth() - 1);
		} else if (period === 'year') {
			cutoff.setFullYear(now.getFullYear() - 1);
		}

		return playlists.filter((p) => new Date(p.created_at) >= cutoff);
	}

	function sortPlaylists(playlists: Playlist[], order: string): Playlist[] {
		return [...playlists].sort((a, b) => {
			const dateA = new Date(a.created_at).getTime();
			const dateB = new Date(b.created_at).getTime();
			return order === 'desc' ? dateB - dateA : dateA - dateB;
		});
	}

	let filteredPlaylists = $derived(
		sortPlaylists(filterByPeriod(data.playlists.playlists, period), sortOrder)
	);
</script>

<div class="container mx-auto px-4 py-8">
	<div class="mb-8 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
		<div class="space-y-1">
			<h1 class="text-3xl font-bold tracking-tight">Your Playlists</h1>
			<p class="text-muted-foreground">
				{filteredPlaylists.length} playlist{filteredPlaylists.length === 1 ? '' : 's'}
			</p>
		</div>
		<PlaylistFilters bind:period bind:sortOrder />
	</div>

	{#if filteredPlaylists.length > 0}
		<div class="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
			{#each filteredPlaylists as playlist (playlist.id)}
				<PlaylistCard {playlist} />
			{/each}
		</div>
	{:else}
		<div class="flex flex-col items-center justify-center gap-4 rounded-2xl border border-dashed border-primary/20 bg-gradient-to-br from-primary/5 via-transparent to-destructive/5 p-16 text-center">
			<div class="rounded-full bg-gradient-to-br from-primary/20 to-destructive/10 p-4">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.5"
					class="h-8 w-8 text-primary/60"
				>
					<path d="M9 18V5l12-2v13" />
					<circle cx="6" cy="18" r="3" />
					<circle cx="18" cy="16" r="3" />
				</svg>
			</div>
			<div class="space-y-1">
				<p class="font-medium">No playlists found</p>
				<p class="text-sm text-muted-foreground">Try adjusting your filters to see more results.</p>
			</div>
		</div>
	{/if}
</div>
