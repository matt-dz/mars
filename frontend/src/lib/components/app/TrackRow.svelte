<script lang="ts">
	import type { Track } from '$lib/api/types';

	let { track }: { track: Track } = $props();
</script>

<div class="flex items-center gap-4 rounded-md p-3 transition-colors hover:bg-muted/50">
	{#if track.image_url}
		<img
			src={track.image_url}
			alt="{track.name} album cover"
			class="h-12 w-12 shrink-0 rounded object-cover"
		/>
	{:else}
		<div class="h-12 w-12 shrink-0 rounded bg-muted"></div>
	{/if}

	<div class="min-w-0 flex-1">
		<p class="truncate font-medium">{track.name}</p>
		<p class="truncate text-sm text-muted-foreground">
			{track.artists.join(', ')}
		</p>
	</div>
</div>
