<script lang="ts">
	import PgpVaultTerminal from '$lib/components/PgpVaultTerminal.svelte';
	let { form } = $props();
</script>

<svelte:head>
	<title>·</title>
	<meta name="robots" content="noindex, nofollow, noarchive" />
</svelte:head>

<PgpVaultTerminal {form} />
