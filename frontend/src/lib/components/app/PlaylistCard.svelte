<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { resolve } from '$app/paths';
	import type { Playlist } from '$lib/api/types';

	let { playlist }: { playlist: Playlist } = $props();

	let typeLabel = $derived(playlist.type === 'weekly' ? 'Weekly' : 'Monthly');
	let formattedDate = $derived(
		new Date(playlist.created_at).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		})
	);
</script>

<a href={resolve(`/playlist/${playlist.id}`)} class="block">
	<Card.Root class="transition-shadow hover:shadow-md">
		<Card.Header>
			<div class="flex items-start justify-between gap-2">
				<Card.Title class="line-clamp-1">{playlist.name}</Card.Title>
				<Badge variant="secondary">{typeLabel}</Badge>
			</div>
			<Card.Description>{formattedDate}</Card.Description>
		</Card.Header>
	</Card.Root>
</a>
