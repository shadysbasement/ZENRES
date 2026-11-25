<script lang="ts">
	import { page } from '$app/state';
</script>

<div class="err">
	<span class="code mono">{page.status}</span>
	<h1>{page.status === 404 ? 'This page does not exist.' : 'Something went wrong.'}</h1>
	<a href="/dashboard" class="btn btn-dark">Back to dashboard</a>
</div>

<style>
	.err {
		min-height: 60vh;
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: flex-start;
		gap: 16px;
	}
	.code {
		font-size: 6rem;
		line-height: 1;
		letter-spacing: -0.06em;
		color: var(--ink-primary);
	}
	h1 {
		font-size: 1.8rem;
	}
</style>
