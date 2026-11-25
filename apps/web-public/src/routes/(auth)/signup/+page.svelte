<script lang="ts">
	import AuthProviders from '$lib/components/AuthProviders.svelte';
	import SplitText from '$design/components/SplitText.svelte';
	import { page } from '$app/state';
	const plan = page.url.searchParams.get('plan');
	const planName: Record<string, string> = { starter: 'MSME Starter', growth: 'MSME Growth', importer_hub: 'EU Importer Hub' };
</script>

<svelte:head><title>Create your workspace — ZENRES</title></svelte:head>

<div class="head">
	<span class="eyebrow">Create workspace {#if plan && planName[plan]}· <span class="text-emerald">{planName[plan]}</span>{/if}</span>
	<SplitText tag="h1" class="auth-title" lines={[[{ text: 'Start with a ' }, { text: '14-day sandbox.', serif: true }]]} />
	<p class="sub">No card required. Pick a plan during onboarding — or keep exploring with seeded pilot data.</p>
</div>
<AuthProviders mode="signup" />
<p class="switch">Already have an account? <a href="/login" class="link-u text-emerald">Sign in</a></p>

<style>
	.head {
		display: flex;
		flex-direction: column;
		gap: 12px;
		margin-bottom: 28px;
	}
	.sub {
		font-size: var(--fs-sm);
	}
	.switch {
		margin-top: 24px;
		font-size: var(--fs-sm);
		text-align: center;
	}
</style>
