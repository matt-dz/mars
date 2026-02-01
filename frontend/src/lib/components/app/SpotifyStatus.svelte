<script lang="ts">
	import type { SpotifyStatus } from '$lib/api/types';

	let { status }: { status: SpotifyStatus } = $props();
</script>

<div class="flex items-center gap-3">
	{#if status.connected}
		<div
			class="flex items-center gap-2 rounded-full bg-green-500/10 px-3 py-1.5 text-sm font-medium text-green-700 dark:text-green-400"
		>
			<span class="relative flex h-2 w-2">
				<span
					class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75"
				></span>
				<span class="relative inline-flex h-2 w-2 rounded-full bg-green-500"></span>
			</span>
			Connected
		</div>
	{:else}
		<div
			class="flex items-center gap-2 rounded-full bg-muted px-3 py-1.5 text-sm font-medium text-muted-foreground"
		>
			<span class="h-2 w-2 rounded-full bg-muted-foreground/50"></span>
			Not connected
		</div>
	{/if}
</div>
