<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/auth';
	let { data } = $props();
	let loading = $state(true);
	let signedIn = $state(true);
	let available = $state(true);
	let marketing = $state(false);
	let saving = $state(false);
	let message = $state('');
	onMount(() => {
		void (async () => {
			const result = await api<{ marketing_opt_in: boolean }>('consent', undefined, 'GET');
			loading = false;
			if (!result.ok) { signedIn = result.status !== 401; available = false; message = result.status === 401 ? '' : 'Preferences are temporarily unavailable.'; return; }
			marketing = result.data.marketing_opt_in;
		})();
	});
	async function save(next: boolean) {
		saving = true;
		message = '';
		const result = await api<{ marketing_opt_in: boolean }>('consent/marketing', { marketing_opt_in: next });
		saving = false;
		if (result.ok) { marketing = result.data.marketing_opt_in; message = 'Preference saved.'; }
		else message = 'We could not save your preference. Please try again.';
	}
</script>

<svelte:head><title>Privacy preferences — ZENRES</title>{#if !data.legalPublished}<meta name="robots" content="noindex, nofollow" />{/if}</svelte:head>

<section class="preferences">
	<p class="eyebrow">Account settings</p>
	<h1>Privacy preferences</h1>
	<p>Review the <a href="/legal/privacy">Privacy Notice</a> and <a href="/legal/cookies">Cookies and Device Storage</a> page.</p>
	{#if loading}<p role="status">Loading your preferences…</p>
	{:else if !signedIn}<p>Sign in to manage your choices. <a href="/login?next=/account/privacy">Sign in</a></p>
	{:else if !available}<p role="status">{message}</p>
	{:else}
		<div class="choice"><div><h2>Product updates by email</h2><p>Optional messages about ZENRES. Your account works without them.</p></div><button class="btn" disabled={saving} aria-pressed={marketing} onclick={() => save(!marketing)}>{marketing ? 'On · turn off' : 'Off · turn on'}</button></div>
		{#if message}<p role="status">{message}</p>{/if}
	{/if}
</section>

<style>
	.preferences { max-width: 600px; margin: auto; padding: 60px 0; }
	h1 { font-size: clamp(2rem, 4vw, 3.5rem); letter-spacing: -.045em; margin: 10px 0 16px; }
	p { line-height: 1.65; color: var(--ink-muted); }
	a { text-decoration: underline; text-underline-offset: 2px; }
	.choice { margin-top: 35px; padding: 22px; border: var(--hairline); border-radius: 18px; background: #fcfdff; display: flex; align-items: center; justify-content: space-between; gap: 16px; }
	h2 { font-size: 17px; margin-bottom: 6px; }
	.choice p { font-size: 13px; }
	.choice button { flex: none; }
	@media (max-width: 600px) { .preferences { padding: 30px 0; } .choice { flex-direction: column; align-items: flex-start; } }
</style>
