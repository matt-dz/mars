<script lang="ts">
	import TrackRow from './TrackRow.svelte';
	import type { Track } from '$lib/api/types';

	let { tracks }: { tracks: Track[] } = $props();
</script>

<div class="space-y-1 rounded-xl bg-card p-2 shadow-sm">
	{#each tracks as track, index (track.id)}
		<TrackRow {track} {index} />
	{/each}

	{#if tracks.length === 0}
		<div class="flex flex-col items-center justify-center gap-3 py-16 text-muted-foreground">
			<div class="rounded-full bg-gradient-to-br from-primary/15 to-destructive/10 p-4">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.5"
					class="h-10 w-10 text-primary/50"
				>
					<path d="M9 18V5l12-2v13" />
					<circle cx="6" cy="18" r="3" />
					<circle cx="18" cy="16" r="3" />
				</svg>
			</div>
			<p>No tracks in this playlist</p>
		</div>
	{/if}
</div>
