<script lang="ts">
	import AuthProviders from '$lib/components/AuthProviders.svelte';
	import SplitText from '$design/components/SplitText.svelte';
</script>

<svelte:head><title>Sign in — Zevra</title></svelte:head>

<div class="head">
	<span class="eyebrow">Sign in</span>
	<SplitText tag="h1" class="auth-title" lines={[[{ text: 'Welcome back to ' }, { text: 'your workspace.', serif: true }]]} />
</div>
<AuthProviders mode="login" />
<p class="switch">New to Zevra? <a href="/signup" class="link-u text-emerald">Create a free workspace</a></p>

<style>
	.head {
		display: flex;
		flex-direction: column;
		gap: 12px;
		margin-bottom: 28px;
	}
	:global(.auth-title) {
		font-size: clamp(2rem, 3.4vw, 2.75rem);
		letter-spacing: -0.03em;
		line-height: 1.05;
	}
	.switch {
		margin-top: 24px;
		font-size: var(--fs-sm);
		text-align: center;
	}
</style>
