<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { Toaster } from '@/components/ui/sonner/index.js';
	import * as Tooltip from '$lib/components/ui/tooltip';

	let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>
<Tooltip.Provider>
	<Toaster position="top-center" richColors />
	{@render children()}
</Tooltip.Provider>
