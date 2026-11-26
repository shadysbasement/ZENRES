<script lang="ts">
	import Logo from '$design/components/Logo.svelte';
	import SplitText from '$design/components/SplitText.svelte';

	let { children, data } = $props();

	const guarantees = [
		['Tenant isolation', 'Row-level security on every tenant-scoped table'],
		['Argon2id', 't=3 · 64 MiB · p=4 · 16-byte salt'],
		['Sessions', 'Opaque 256-bit tokens · HttpOnly · SameSite'],
		['Passkeys', 'WebAuthn — optimised for Proton Pass'],
		['Audit', 'Immutable auth event log']
	];
</script>

<div class="auth">
	<aside class="spec" aria-label="Security specification">
		<div class="spec-grid" aria-hidden="true">{#each Array(8) as _, i (i)}<span style="--i:{i}"></span>{/each}</div>
		<a href="/" class="brand"><Logo size={26} inverse /></a>
		<div class="spec-body">
			<span class="eyebrow">Security specification</span>
			<SplitText tag="p" class="spec-title" lines={[[{ text: 'Your factory data,' }], [{ text: 'isolated by design.', serif: true }]]} />
			<dl class="specs">
				{#each guarantees as [k, v], i (k)}
					<div style="--i:{i}"><dt class="mono">{k}</dt><dd>{v}</dd></div>
				{/each}
			</dl>
		</div>
		<p class="spec-foot mono">ZV-AUTH · OIDC (Google, Apple) · Proton magic link + WKD</p>
	</aside>
	<main class="pane">
		<div class="pane-top">
			<a href="/" class="mobile-brand"><Logo size={22} /></a>
			<a href="/" class="back link-u">← Back to ZENRES</a>
		</div>
		<div class="pane-body">{@render children()}</div>
		<p class="pane-foot">ZENRES provides reporting support, not legal or tax advice. {#if data.legalAvailable}<a href="/legal/terms">Terms</a> · <a href="/legal/privacy">Privacy</a>{/if}</p>
	</main>
</div>

<style>
	.auth {
		min-height: 100dvh;
		display: grid;
		grid-template-columns: 5fr 7fr;
	}
	.spec {
		position: sticky;
		top: 0;
		height: 100dvh;
		background: var(--obsidian);
		color: #a9b6b0;
		padding: 36px clamp(28px, 4vw, 56px);
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		overflow: hidden;
	}
	.spec-grid {
		position: absolute;
		inset: 0;
		display: grid;
		grid-template-columns: repeat(8, 1fr);
		pointer-events: none;
	}
	.spec-grid span {
		border-left: 1px solid var(--obsidian-line);
		opacity: 0.6;
		animation: grow 1.6s var(--ease-expo) both;
		animation-delay: calc(var(--i) * 60ms);
		transform-origin: top;
	}
	@keyframes grow {
		from {
			transform: scaleY(0);
		}
	}
	.brand,
	.spec-body,
	.spec-foot {
		position: relative;
	}
	.spec-body {
		display: flex;
		flex-direction: column;
		gap: 24px;
	}
	.spec .eyebrow {
		color: var(--vault-phosphor);
	}
	:global(.spec-title) {
		font-size: clamp(2.2rem, 3.6vw, 3.6rem);
		font-weight: 500;
		letter-spacing: -0.035em;
		line-height: 1;
		color: var(--ink-inverse);
	}
	:global(.spec-title .serif) {
		color: var(--vault-phosphor);
	}
	.specs {
		display: flex;
		flex-direction: column;
		border-top: 1px solid var(--obsidian-line);
	}
	.specs div {
		display: grid;
		grid-template-columns: 130px 1fr;
		gap: 12px;
		padding: 12px 0;
		border-bottom: 1px solid var(--obsidian-line);
		font-size: var(--fs-xs);
		animation: fade 900ms var(--ease-expo) both;
		animation-delay: calc(var(--i) * 80ms + 400ms);
	}
	@keyframes fade {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
	}
	.specs dt {
		color: var(--ink-inverse);
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}
	.spec-foot {
		font-size: 10.5px;
		color: var(--vault-dim);
		letter-spacing: 0.04em;
	}
	.pane {
		display: flex;
		flex-direction: column;
		padding: 28px clamp(20px, 5vw, 72px);
		min-height: 100dvh;
	}
	.pane-top {
		display: flex;
		justify-content: flex-end;
		align-items: center;
	}
	.mobile-brand {
		display: none;
		margin-right: auto;
	}
	.back {
		font-size: var(--fs-xs);
		color: var(--ink-muted);
	}
	.pane-body {
		flex: 1;
		display: flex;
		flex-direction: column;
		justify-content: center;
		width: 100%;
		max-width: 460px;
		margin: 40px auto;
	}
	.pane-foot {
		font-size: 11px;
		color: var(--ink-muted);
		text-align: center;
		max-width: 460px;
		margin: 0 auto;
	}
	@media (max-width: 900px) {
		.auth {
			grid-template-columns: 1fr;
		}
		.spec {
			display: none;
		}
		.mobile-brand {
			display: block;
		}
	}
</style>
