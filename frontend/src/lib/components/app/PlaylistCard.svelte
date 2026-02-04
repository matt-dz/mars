<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { resolve } from '$app/paths';
	import type { Playlist } from '$lib/api/types';

	let { playlist }: { playlist: Playlist } = $props();

	let typeLabel = $derived(
		playlist.type === 'weekly' ? 'Weekly' : playlist.type === 'monthly' ? 'Monthly' : 'Custom'
	);
	let formattedDate = $derived(
		new Date(playlist.created_at).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		})
	);

	// Dynamic gradient based on playlist type
	let gradientClass = $derived(
		playlist.type === 'weekly'
			? 'from-primary/10 via-primary/5 to-transparent'
			: playlist.type === 'monthly'
				? 'from-destructive/10 via-destructive/5 to-transparent'
				: 'from-primary/10 via-destructive/10 to-transparent'
	);

	let accentColor = $derived(
		playlist.type === 'weekly'
			? 'text-primary'
			: playlist.type === 'monthly'
				? 'text-destructive'
				: 'text-primary'
	);
</script>

<a href={resolve(`/playlist/${playlist.id}`)} class="group block">
	<Card.Root
		class="relative overflow-hidden border-border/50 transition-all duration-300 hover:-translate-y-1 hover:border-primary/30 hover:shadow-xl hover:shadow-primary/5"
	>
		<!-- Background decoration -->
		<div
			class="pointer-events-none absolute inset-0 bg-gradient-to-br opacity-50 transition-opacity group-hover:opacity-70 {gradientClass}"
		></div>

		<!-- Decorative circles -->
		<div
			class="pointer-events-none absolute -top-8 -right-8 h-24 w-24 rounded-full bg-gradient-to-br from-primary/10 to-destructive/10 blur-2xl transition-transform group-hover:scale-125"
		></div>
		<div
			class="pointer-events-none absolute -bottom-6 -left-6 h-20 w-20 rounded-full bg-gradient-to-tr from-destructive/10 to-primary/10 blur-2xl transition-transform group-hover:scale-125"
		></div>

		<Card.Header class="relative space-y-4 pt-6 pb-6">
			<!-- Top bar with type badge -->
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2">
					<!-- Music note icon -->
					<div
						class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-primary/20 to-destructive/20 transition-transform group-hover:scale-110"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							viewBox="0 0 24 24"
							fill="currentColor"
							class="h-4 w-4 {accentColor}"
						>
							<path d="M9 18V5l12-2v13" />
							<circle cx="6" cy="18" r="3" />
							<circle cx="18" cy="16" r="3" />
						</svg>
					</div>
					<Badge
						variant="secondary"
						class="border border-border/50 bg-background/80 backdrop-blur-sm"
					>
						{typeLabel}
					</Badge>
				</div>
				<!-- Arrow icon -->
				<div
					class="flex h-6 w-6 items-center justify-center rounded-full bg-primary/10 transition-all group-hover:translate-x-0.5 group-hover:bg-primary/20"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 20 20"
						fill="currentColor"
						class="h-3.5 w-3.5 {accentColor}"
					>
						<path
							fill-rule="evenodd"
							d="M3 10a.75.75 0 01.75-.75h10.638L10.23 5.29a.75.75 0 111.04-1.08l5.5 5.25a.75.75 0 010 1.08l-5.5 5.25a.75.75 0 11-1.04-1.08l4.158-3.96H3.75A.75.75 0 013 10z"
							clip-rule="evenodd"
						/>
					</svg>
				</div>
			</div>

			<!-- Playlist name -->
			<div class="space-y-2">
				<Card.Title
					class="line-clamp-2 text-xl leading-tight font-bold tracking-tight transition-colors group-hover:text-primary"
				>
					{playlist.name}
				</Card.Title>

				<!-- Date -->
				<Card.Description class="flex items-center gap-2 text-xs">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 20 20"
						fill="currentColor"
						class="h-3.5 w-3.5 opacity-50"
					>
						<path
							fill-rule="evenodd"
							d="M5.75 2a.75.75 0 01.75.75V4h7V2.75a.75.75 0 011.5 0V4h.25A2.75 2.75 0 0118 6.75v8.5A2.75 2.75 0 0115.25 18H4.75A2.75 2.75 0 012 15.25v-8.5A2.75 2.75 0 014.75 4H5V2.75A.75.75 0 015.75 2zm-1 5.5c-.69 0-1.25.56-1.25 1.25v6.5c0 .69.56 1.25 1.25 1.25h10.5c.69 0 1.25-.56 1.25-1.25v-6.5c0-.69-.56-1.25-1.25-1.25H4.75z"
							clip-rule="evenodd"
						/>
					</svg>
					<span class="opacity-80">{formattedDate}</span>
				</Card.Description>
			</div>
		</Card.Header>

		<!-- Bottom accent line -->
		<div
			class="h-1 w-full bg-gradient-to-r from-primary/50 via-primary/30 to-destructive/30 transition-all group-hover:h-1.5 group-hover:from-primary group-hover:via-primary/70 group-hover:to-destructive/50"
		></div>
	</Card.Root>
</a>
