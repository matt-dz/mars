<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip';
	import type { Track } from '$lib/api/types';

	let { track, index }: { track: Track; index?: number } = $props();
</script>

<div class="group flex items-center gap-4 rounded-lg px-4 py-3 transition-all hover:bg-accent/50">
	{#if index !== undefined}
		<span
			class="w-6 text-center text-sm text-muted-foreground tabular-nums group-hover:text-accent-foreground"
		>
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
		<div
			class="flex h-14 w-14 shrink-0 items-center justify-center rounded-md bg-gradient-to-br from-primary/25 via-accent/15 to-destructive/10"
		>
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
		<p class="truncate leading-tight font-medium group-hover:text-accent-foreground">
			{track.name}
		</p>
		<p class="truncate text-sm text-muted-foreground">
			{track.artists.join(', ')}
		</p>
	</div>

	<Tooltip.Root>
		<Tooltip.Trigger>
			<div
				class="flex shrink-0 items-center gap-1.5 rounded-md px-2 py-1 text-sm text-muted-foreground tabular-nums transition-colors hover:bg-muted hover:text-foreground"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					viewBox="0 0 20 20"
					fill="currentColor"
					class="h-4 w-4"
				>
					<path
						d="M6.3 2.84A1.5 1.5 0 004 4.11v11.78a1.5 1.5 0 002.3 1.27l9.344-5.891a1.5 1.5 0 000-2.538L6.3 2.841z"
					/>
				</svg>
				{track.plays}
			</div>
		</Tooltip.Trigger>
		<Tooltip.Content>
			<p>{track.plays} {track.plays === 1 ? 'play' : 'plays'} this period</p>
		</Tooltip.Content>
	</Tooltip.Root>
</div>
