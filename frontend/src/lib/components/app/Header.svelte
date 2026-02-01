<script lang="ts">
	import ProfileWidget from './ProfileWidget.svelte';
	import type { User } from '$lib/api/types';
	import { resolve } from '$app/paths';

	let { user }: { user: User } = $props();
</script>

<header class="sticky top-0 z-50 border-b border-border/50 bg-background/90 backdrop-blur-lg">
	<div class="container mx-auto flex h-16 items-center justify-between px-4">
		<a
			href={resolve('/')}
			class="group flex items-center gap-2 transition-opacity hover:opacity-80"
		>
			<div
				class="flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-br from-primary via-primary/80 to-destructive/60 shadow-sm shadow-primary/20"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					viewBox="0 0 24 24"
					fill="currentColor"
					class="h-4 w-4 text-primary-foreground"
				>
					<circle cx="12" cy="12" r="10" />
				</svg>
			</div>
			<span
				class="bg-gradient-to-r from-primary to-destructive/80 bg-clip-text text-xl font-bold tracking-tight text-transparent"
			>
				mars
			</span>
		</a>
		<ProfileWidget {user} />
	</div>
</header>
