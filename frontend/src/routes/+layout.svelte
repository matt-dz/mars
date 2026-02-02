<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { page } from '$app/state';
	import { Toaster } from '@/components/ui/sonner/index.js';
	import * as Tooltip from '$lib/components/ui/tooltip';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />

	<!-- Primary Meta Tags -->
	<title>Mars - Music Archival Software</title>
	<meta name="title" content="Mars - Music Archival Software" />
	<meta
		name="description"
		content="Mars automatically creates weekly and monthly playlists from your Spotify listening history. Rediscover your favorite tracks and see how your taste evolves over time."
	/>
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<meta name="theme-color" content="#d97706" />

	<!-- Open Graph / Facebook -->
	<meta property="og:type" content="website" />
	<meta property="og:url" content={page.url.href} />
	<meta property="og:title" content="Mars - Music Archival Software" />
	<meta
		property="og:description"
		content="Mars automatically creates weekly and monthly playlists from your Spotify listening history. Rediscover your favorite tracks and see how your taste evolves over time."
	/>
	<meta property="og:site_name" content="Mars" />
	<meta property="og:image" content={`${page.url.origin}/cover.jpg`} />
	<meta property="og:image:width" content="1200" />
	<meta property="og:image:height" content="630" />

	<!-- Twitter -->
	<meta property="twitter:card" content="summary_large_image" />
	<meta property="twitter:url" content={page.url.href} />
	<meta property="twitter:title" content="Mars - Music Archival Software" />
	<meta
		property="twitter:description"
		content="Mars automatically creates weekly and monthly playlists from your Spotify listening history. Rediscover your favorite tracks and see how your taste evolves over time."
	/>
	<meta property="twitter:image" content={`${page.url.origin}/cover.jpg`} />

	<!-- Canonical URL -->
	<link rel="canonical" href={page.url.href} />

	<!-- Favicons -->
	<link rel="icon" type="image/png" href="/favicon-96x96.png" sizes="96x96" />
	<link rel="icon" type="image/svg+xml" href="/favicon.svg" />
	<link rel="shortcut icon" href="/favicon.ico" />
	<link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png" />
	<meta name="apple-mobile-web-app-title" content="Mars" />
	<link rel="manifest" href="/site.webmanifest" />
</svelte:head>
<Tooltip.Provider>
	<Toaster position="top-center" richColors />
	{@render children()}
</Tooltip.Provider>
