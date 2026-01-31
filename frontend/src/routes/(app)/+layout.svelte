<script lang="ts">
	import Header from '$lib/components/app/Header.svelte';
	import type { Snippet } from 'svelte';
	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: Snippet } = $props();
</script>

<div class="flex min-h-screen flex-col">
	<Header user={data.user} />
	<main class="flex-1">
		{@render children()}
	</main>
</div>
