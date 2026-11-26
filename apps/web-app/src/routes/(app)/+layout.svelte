<script lang="ts">
	import AppShell from '$lib/components/AppShell.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { onDestroy } from 'svelte';

	let { children } = $props();
	onDestroy(() => ws.destroy());
</script>

<AppShell>
	{@render children()}
</AppShell>
