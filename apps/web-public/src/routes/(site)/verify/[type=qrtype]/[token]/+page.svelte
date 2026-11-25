<script lang="ts">
	import { reveal } from '$design/motion';
	import { DISCLAIMERS } from '$lib/content';
	import type { PublicBatch, PublicContainer } from '$lib/server/passports';

	let { data } = $props();
	const passport = $derived(data.state === 'ok' || data.state === 'revoked' ? (data.passport as PublicBatch | PublicContainer) : null);
	const fmtDate = (d: string | null | undefined) =>
		d ? new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' }) : '—';
</script>

<svelte:head>
	<title>Verify QR passport — ZENRES</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<section class="verify container">
	<div class="v-head" use:reveal>
		<span class="eyebrow">Public QR passport resolver · {data.type === 'c' ? 'Container / Shipment' : 'Product / Batch'}</span>
		{#if data.source === 'demo'}<span class="pill pill-outline">Sample passport</span>{/if}
	</div>

	{#if data.state === 'not_found'}
		<div class="notice card" use:reveal>
			<span class="pill pill-crimson">Not found</span>
			<h1>This passport could not be resolved.</h1>
			<p>The QR token is unknown or malformed. ZENRES passport tokens are 26-character identifiers — check the label for damage, or ask the manufacturer for a reprint.</p>
		</div>
	{:else if data.state === 'unavailable'}
		<div class="notice card" use:reveal>
			<span class="pill pill-terracotta">Temporarily unavailable</span>
			<h1>The verification service is not reachable.</h1>
			<p>Please try again in a moment.</p>
		</div>
	{:else if passport}
		{#if data.state === 'revoked'}
			<div class="revoked" use:reveal>
				<span class="pill pill-crimson">Revoked label</span>
				<p>This label was revoked by the manufacturer (damaged or reprinted). Do not rely on it — scan the replacement label on the pack.</p>
			</div>
		{/if}

		<div class="v-grid" class:dim={data.state === 'revoked'}>
			<div class="main card" use:reveal>
				<div class="mfr">
					<div>
						<span class="eyebrow">Manufacturer</span>
						<h1>{passport.manufacturer.name}</h1>
						<p class="muted">{passport.manufacturer.city} · {passport.manufacturer.country}</p>
					</div>
					{#if passport.manufacturer.verified}<span class="pill pill-emerald"><span class="dot"></span>Verified MSME</span>{/if}
				</div>

				{#if passport.type === 'p'}
					<dl class="facts">
						<div><dt class="eyebrow">Product</dt><dd>{passport.product.name}</dd></div>
						<div><dt class="eyebrow">CN code</dt><dd class="num">{passport.product.cn_code}</dd></div>
						<div><dt class="eyebrow">Batch</dt><dd class="num">{passport.batch.code}</dd></div>
						<div><dt class="eyebrow">Produced</dt><dd>{fmtDate(passport.batch.produced_on)}</dd></div>
						<div><dt class="eyebrow">Sector</dt><dd>{passport.product.sector}</dd></div>
						<div><dt class="eyebrow">Stage</dt><dd>{passport.batch.stage}</dd></div>
					</dl>
					<div class="carbon">
						<div>
							<span class="eyebrow">Embedded emissions intensity</span>
							<div class="big num">{passport.carbon.intensity_tco2e_per_t ?? '—'}<span class="unit">tCO₂e / t</span></div>
						</div>
						<div class="badges">
							<span class="pill" class:pill-emerald={passport.carbon.data_quality === 'Measured'} class:pill-cobalt={passport.carbon.data_quality !== 'Measured'}>{passport.carbon.data_quality}</span>
							<span class="pill pill-outline">{passport.verification_tier}</span>
						</div>
					</div>
					{#if passport.container}
						<a class="link-row" href="/verify/c/{passport.container.token}">Packed in container <span class="mono">{passport.container.code}</span> <span class="arrow">→</span></a>
					{/if}
				{:else}
					<dl class="facts">
						<div><dt class="eyebrow">Container</dt><dd class="num">{passport.container.code}</dd></div>
						<div><dt class="eyebrow">Status</dt><dd>{passport.container.dispatch_status}</dd></div>
						<div class="wide"><dt class="eyebrow">Route</dt><dd>{passport.container.route.join(' → ')}</dd></div>
					</dl>
					<ol class="milestones" role="list">
						{#each passport.container.milestones as m, i (m.label)}
							<li class:done={m.done} style="--i:{i}">
								<span class="node"></span>
								<span>{m.label}</span>
								<span class="mono muted">{fmtDate(m.at)}</span>
							</li>
						{/each}
					</ol>
					<div class="gate">
						<div>
							<span class="eyebrow">Linked batches & documents</span>
							<p>{passport.batch_count} production batches and transport documents are linked to this container. Partner authentication required.</p>
						</div>
						<a href="/login?next={encodeURIComponent(`/verify/c/${passport.token}`)}" class="btn btn-dark btn-sm">Partner sign in</a>
					</div>
				{/if}
			</div>

			<aside class="side" use:reveal={{ delay: 100 }}>
				<div class="card anchor">
					<span class="eyebrow">Hash anchoring</span>
					{#if passport.anchor.anchored}
						<p class="ok"><span class="dot"></span>Anchored {fmtDate(passport.anchor.anchored_at)}</p>
						<code class="hash">{passport.anchor.merkle_root}</code>
					{:else}
						<p>Not yet anchored.</p>
					{/if}
					<p class="fine">{DISCLAIMERS.anchoring}</p>
				</div>
				<div class="card token">
					<span class="eyebrow">Public token</span>
					<code>{passport.token}</code>
					<p class="fine">Non-guessable identifier. No confidential data is encoded in the QR payload.</p>
				</div>
			</aside>
		</div>
	{/if}
	<p class="footnote">{DISCLAIMERS.reporting}</p>
</section>

<style>
	.verify {
		padding-top: 128px;
		padding-bottom: 40px;
		min-height: 80vh;
	}
	.v-head {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 12px;
		margin-bottom: 24px;
	}
	.notice {
		padding: 40px;
		display: flex;
		flex-direction: column;
		gap: 16px;
		align-items: flex-start;
		max-width: 720px;
	}
	.notice h1 {
		font-size: var(--fs-h2);
	}
	.revoked {
		display: flex;
		gap: 14px;
		align-items: center;
		padding: 14px 18px;
		margin-bottom: 18px;
		border: 1px solid var(--status-crimson);
		background: var(--status-crimson-tint);
		border-radius: var(--radius-md);
		font-size: var(--fs-sm);
		color: var(--status-crimson);
	}
	.v-grid {
		display: grid;
		grid-template-columns: 8fr 4fr;
		gap: var(--gap);
	}
	.v-grid.dim {
		opacity: 0.55;
		filter: grayscale(0.6);
	}
	.main {
		padding: 32px;
		display: flex;
		flex-direction: column;
		gap: 28px;
	}
	.mfr {
		display: flex;
		justify-content: space-between;
		gap: 16px;
		align-items: flex-start;
	}
	.mfr h1 {
		margin-top: 8px;
		font-size: var(--fs-h2);
		letter-spacing: -0.03em;
	}
	.facts {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 1px;
		background: var(--border-hairline);
		border: var(--hairline);
		border-radius: var(--radius-sm);
		overflow: hidden;
	}
	.facts > div {
		background: var(--surface-card);
		padding: 16px;
	}
	.facts .wide {
		grid-column: span 1;
	}
	.facts dd {
		margin-top: 6px;
		color: var(--ink-primary);
		font-size: var(--fs-sm);
	}
	.carbon {
		display: flex;
		justify-content: space-between;
		align-items: flex-end;
		gap: 16px;
		padding: 20px;
		border-radius: var(--radius-sm);
		background: var(--surface-canvas);
		border: var(--hairline);
	}
	.big {
		margin-top: 8px;
		font-size: 2.6rem;
		color: var(--ink-primary);
		letter-spacing: -0.04em;
		line-height: 1;
	}
	.unit {
		font-size: 0.35em;
		color: var(--ink-muted);
		margin-left: 8px;
		letter-spacing: 0;
	}
	.badges {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.link-row {
		display: flex;
		gap: 8px;
		align-items: center;
		font-size: var(--fs-sm);
		color: var(--brand-emerald);
		font-weight: 500;
	}
	.milestones {
		display: flex;
		flex-direction: column;
		list-style: none;
		padding: 0;
	}
	.milestones li {
		position: relative;
		display: grid;
		grid-template-columns: 22px 1fr auto;
		gap: 10px;
		align-items: center;
		padding: 12px 0;
		border-bottom: var(--hairline);
		font-size: var(--fs-sm);
		color: var(--ink-muted);
	}
	.milestones li.done {
		color: var(--ink-primary);
	}
	.milestones .node {
		width: 10px;
		height: 10px;
		border-radius: 50%;
		border: 1px dashed var(--border-strong);
	}
	.milestones .done .node {
		background: var(--brand-emerald);
		border: 0;
		box-shadow: 0 0 0 3px var(--brand-emerald-tint);
	}
	.gate {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 16px;
		padding: 18px;
		background: var(--surface-subtle);
		border-radius: var(--radius-sm);
		font-size: var(--fs-sm);
	}
	.gate p {
		margin-top: 4px;
	}
	.side {
		display: flex;
		flex-direction: column;
		gap: var(--gap);
	}
	.side .card {
		padding: 22px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.ok {
		display: flex;
		align-items: center;
		gap: 8px;
		color: var(--brand-emerald);
		font-size: var(--fs-sm);
		font-weight: 500;
	}
	code {
		font-size: 11px;
		word-break: break-all;
		color: var(--ink-primary);
		background: var(--surface-canvas);
		border: var(--hairline);
		padding: 8px 10px;
		border-radius: var(--radius-sm);
	}
	.fine {
		font-size: 11px;
		color: var(--ink-muted);
		line-height: 1.5;
	}
	.footnote {
		margin-top: 28px;
		font-size: 11px;
		color: var(--ink-muted);
	}
	@media (max-width: 900px) {
		.v-grid {
			grid-template-columns: 1fr;
		}
		.facts {
			grid-template-columns: 1fr 1fr;
		}
	}
</style>
