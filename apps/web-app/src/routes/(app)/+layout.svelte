<script lang="ts">
	import AppShell from '$lib/components/AppShell.svelte';
	import Preloader from '$design/components/Preloader.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { onDestroy } from 'svelte';

	let { children } = $props();
	onDestroy(() => ws.destroy());
</script>

<Preloader
	sessionKey="zv-app-preloaded"
	minDuration={1700}
	tagline="Workspace · {ws.orgName}"
	status={['Verifying session', 'Loading entitlement manifest', 'Syncing Tally vouchers', 'Rendering factory twin']}
/>
<AppShell>
	{@render children()}
</AppShell>
