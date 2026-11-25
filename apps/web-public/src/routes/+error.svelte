<script lang="ts">
	import { page } from '$app/state';
	import Logo from '$design/components/Logo.svelte';
</script>

<svelte:head><title>{page.status === 404 ? 'Page not found' : 'Error'} — ZENRES</title></svelte:head>

<main class="err container">
	<a href="/" class="brand"><Logo size={24} /></a>
	<div class="body">
		<span class="code mono">{page.status}</span>
		<h1>{page.status === 404 ? 'This page does not exist.' : 'Something went wrong.'}</h1>
		<p>{page.status === 404 ? 'The link may be broken or the page may have moved.' : 'Please try again in a moment.'}</p>
		<div class="row">
			<a href="/" class="btn btn-dark">Go to homepage</a>
			<a href="/pricing" class="btn btn-ghost">View pricing</a>
		</div>
	</div>
</main>

<style>
	.err {
		min-height: 100dvh;
		display: flex;
		flex-direction: column;
		padding-block: 32px;
	}
	.body {
		flex: 1;
		display: flex;
		flex-direction: column;
		justify-content: center;
		gap: 18px;
		max-width: 640px;
	}
	.code {
		font-size: clamp(5rem, 16vw, 11rem);
		line-height: 0.9;
		letter-spacing: -0.06em;
		color: var(--ink-primary);
	}
	h1 {
		font-size: var(--fs-h2);
	}
	.row {
		display: flex;
		gap: 8px;
	}
</style>
