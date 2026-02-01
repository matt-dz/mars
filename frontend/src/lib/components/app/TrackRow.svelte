<script lang="ts">
	import type { Track } from '$lib/api/types';

	let { track, index }: { track: Track; index?: number } = $props();
</script>

<div
	class="group flex items-center gap-4 rounded-lg px-4 py-3 transition-all hover:bg-accent/50"
>
	{#if index !== undefined}
		<span class="w-6 text-center text-sm tabular-nums text-muted-foreground group-hover:text-accent-foreground">
			{index + 1}
		</span>
	{/if}

	{#if track.image_url}
		<img
			src={track.image_url}
			alt="{track.name} album cover"
			class="h-14 w-14 shrink-0 rounded-md object-cover shadow-sm transition-shadow group-hover:shadow-md"
		/>
	{:else}
		<div class="flex h-14 w-14 shrink-0 items-center justify-center rounded-md bg-gradient-to-br from-primary/25 via-accent/15 to-destructive/10">
			<svg
				xmlns="http://www.w3.org/2000/svg"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="1.5"
				class="h-6 w-6 text-primary/60"
			>
				<path d="M9 18V5l12-2v13" />
				<circle cx="6" cy="18" r="3" />
				<circle cx="18" cy="16" r="3" />
			</svg>
		</div>
	{/if}

	<div class="min-w-0 flex-1 space-y-0.5">
		<p class="truncate font-medium leading-tight group-hover:text-accent-foreground">
			{track.name}
		</p>
		<p class="truncate text-sm text-muted-foreground">
			{track.artists.join(', ')}
		</p>
	</div>
</div>
