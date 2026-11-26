<script lang="ts">
	import NavBar from '$lib/components/NavBar.svelte';
	import SiteFooter from '$lib/components/SiteFooter.svelte';
	import StealthTrigger from '$lib/components/StealthTrigger.svelte';

	let { children, data } = $props();
</script>

<NavBar />
<main id="main">
	{@render children()}
</main>
<SiteFooter legalAvailable={data.legalAvailable} />
<StealthTrigger />
