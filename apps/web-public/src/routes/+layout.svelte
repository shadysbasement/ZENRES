<script lang="ts">
	import '$design/styles/fonts.css';
	import '$design/styles/tokens.css';
	import '../app.css';
	import { onMount } from 'svelte';
	import { onNavigate } from '$app/navigation';
	import { page } from '$app/state';
	import Cursor from '$design/components/Cursor.svelte';
	import { initSmoothScroll } from '$design/motion';

	let { children } = $props();

	onMount(() => {
		let cleanup: (() => void) | undefined;
		initSmoothScroll().then((c) => (cleanup = c));
		return () => cleanup?.();
	});

	onNavigate((navigation) => {
		if (window.matchMedia('(prefers-reduced-motion: reduce)').matches || !document.startViewTransition || navigation.from?.url.pathname === navigation.to?.url.pathname) return;
		return new Promise((resolve) => {
			document.startViewTransition(async () => {
				resolve();
				await navigation.complete;
				window.scrollTo(0, 0);
			});
		});
	});
</script>

{#if !page.data?.vault}<Cursor />{/if}
{@render children()}
