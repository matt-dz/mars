<script lang="ts">
	import ProfileWidget from './ProfileWidget.svelte';
	import type { User } from '$lib/api/types';
	import { resolve } from '$app/paths';

	let { user }: { user: User } = $props();
</script>

<header class="border-b border-border bg-background">
	<div class="container mx-auto flex h-16 items-center justify-between px-4">
		<a href={resolve('/')} class="text-xl font-bold tracking-tight">mars</a>
		<ProfileWidget {user} />
	</div>
</header>
