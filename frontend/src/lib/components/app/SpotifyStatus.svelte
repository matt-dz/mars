<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import type { SpotifyStatus } from '$lib/api/types';

	let { status }: { status: SpotifyStatus } = $props();
</script>

<div class="flex items-center gap-3">
	{#if status.connected}
		<Badge variant="default" class="bg-green-600">Connected</Badge>
		{#if status.spotify_user_id}
			<span class="text-sm text-muted-foreground">as {status.spotify_user_id}</span>
		{/if}
	{:else}
		<Badge variant="secondary">Not connected</Badge>
	{/if}
</div>
