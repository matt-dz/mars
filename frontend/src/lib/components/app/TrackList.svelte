<script lang="ts">
	import TrackRow from './TrackRow.svelte';
	import type { Track } from '$lib/api/types';

	let { tracks }: { tracks: Track[] } = $props();
</script>

<div class="divide-y divide-border rounded-lg border">
	{#each tracks as track (track.id)}
		<TrackRow {track} />
	{/each}

	{#if tracks.length === 0}
		<div class="p-8 text-center text-muted-foreground">No tracks in this playlist</div>
	{/if}
</div>
