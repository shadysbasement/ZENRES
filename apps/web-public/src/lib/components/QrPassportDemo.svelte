<script lang="ts">
	import { onMount } from 'svelte';
	import { PASSPORT_DEMO, DISCLAIMERS } from '$lib/content';
	import Segmented from '$design/components/Segmented.svelte';
	import { reveal } from '$design/motion';

	let view = $state<'public' | 'partner' | 'private'>('public');
	let qrSvg = $state('');
	let hovered = $state<string | null>(null);

	const allFields = [
		{ f: 'Manufacturer', min: 'public' },
		{ f: 'Product & CN code', min: 'public' },
		{ f: 'Batch & production date', min: 'public' },
		{ f: 'Approved carbon badge', min: 'public' },
		{ f: 'Linked batches & quantities', min: 'partner' },
		{ f: 'Dispatch milestones & documents', min: 'partner' },
		{ f: 'Calculation summary', min: 'partner' },
		{ f: 'BOM & precursor lots', min: 'private' },
		{ f: 'Meter evidence & costing', min: 'private' },
		{ f: 'Audit hash trail', min: 'private' }
	] as const;
	const rank = { public: 0, partner: 1, private: 2 } as const;

	onMount(async () => {
		const QR = await import('qrcode');
		qrSvg = await QR.toString(`${location.origin}/verify/c/${PASSPORT_DEMO.container.token}`, {
			type: 'svg',
			margin: 0,
			color: { dark: '#0E1412', light: '#FFFFFF' },
			errorCorrectionLevel: 'M'
		});
	});
</script>

<div class="qp">
	<div class="schematic card" use:reveal>
		<div class="sch-top">
			<span class="eyebrow">Container / Shipment QR</span>
			<Segmented
				size="sm"
				label="Passport view"
				bind:value={view}
				options={[
					{ value: 'public', label: 'Public' },
					{ value: 'partner', label: 'Partner' },
					{ value: 'private', label: 'Private' }
				]}
			/>
		</div>

		<div class="tree">
			<a class="container-node" href="/verify/c/{PASSPORT_DEMO.container.token}" data-cursor="Scan">
				<div class="qr" aria-label="QR code for container ZV-001">
					{#if qrSvg}{@html qrSvg}{:else}<div class="qr-skel"></div>{/if}
				</div>
				<div class="cmeta">
					<span class="code mono">{PASSPORT_DEMO.container.code}</span>
					<span class="pill pill-emerald"><span class="dot"></span>{PASSPORT_DEMO.container.status}</span>
					<span class="route">{PASSPORT_DEMO.container.route}</span>
					<span class="token mono">token · {PASSPORT_DEMO.container.token.slice(0, 10)}…</span>
				</div>
			</a>

			<svg class="links" viewBox="0 0 60 240" preserveAspectRatio="none" aria-hidden="true">
				{#each [40, 120, 200] as y, i (y)}
					<path d="M0 120 C 30 120, 30 {y}, 60 {y}" class:hot={hovered === PASSPORT_DEMO.batches[i].code} style="--i:{i}" />
				{/each}
			</svg>

			<div class="batches">
				{#each PASSPORT_DEMO.batches as b, i (b.code)}
					<a
						class="batch"
						href="/verify/p/{b.token}"
						style="--i:{i}"
						onpointerenter={() => (hovered = b.code)}
						onpointerleave={() => (hovered = null)}
					>
						<div class="b-top">
							<span class="mono code">{b.code}</span>
							<span class="pill" class:pill-emerald={b.quality === 'Measured'} class:pill-cobalt={b.quality !== 'Measured'}>{b.quality}</span>
						</div>
						<div class="b-row"><span>{b.product}</span><span class="mono">CN {b.cn}</span></div>
						<div class="b-row">
							<span class="mono">{view === 'public' ? '— t' : b.qty}</span>
							<span class="mono">{b.tco2e} tCO₂e/t</span>
						</div>
					</a>
				{/each}
			</div>
		</div>
	</div>

	<div class="views" use:reveal={{ delay: 120 }}>
		<span class="eyebrow">Role-based visibility — {view} view</span>
		<ul role="list">
			{#each allFields as field, i (field.f)}
				{@const visible = rank[view] >= rank[field.min]}
				<li class:visible style="--i:{i}">
					<span class="state" aria-hidden="true"></span>
					<span>{field.f}</span>
					<span class="mono lvl">{field.min}</span>
				</li>
			{/each}
		</ul>
		<p class="fine">
			Public tokens are non-guessable 26-character identifiers; no confidential data is encoded in the QR payload itself. {DISCLAIMERS.anchoring}
		</p>
	</div>
</div>

<style>
	.qp {
		display: grid;
		grid-template-columns: 1.55fr 1fr;
		gap: var(--gap);
	}
	.schematic {
		padding: 22px;
	}
	.sch-top {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;
		margin-bottom: 22px;
	}
	.tree {
		display: grid;
		grid-template-columns: 1fr 60px 1.1fr;
		align-items: center;
	}
	.container-node {
		display: flex;
		flex-direction: column;
		gap: 14px;
		padding: 18px;
		background: var(--obsidian);
		color: var(--ink-inverse);
		border-radius: var(--radius-md);
		transition: transform 600ms var(--ease-expo);
	}
	.container-node:hover {
		transform: translateY(-3px);
	}
	.qr {
		width: 112px;
		aspect-ratio: 1;
		padding: 8px;
		background: #fff;
		border-radius: var(--radius-sm);
	}
	.qr :global(svg) {
		width: 100%;
		height: 100%;
	}
	.qr-skel {
		width: 100%;
		height: 100%;
		background: var(--surface-subtle);
	}
	.cmeta {
		display: flex;
		flex-direction: column;
		gap: 8px;
		align-items: flex-start;
	}
	.code {
		font-size: 1.4rem;
		letter-spacing: -0.02em;
	}
	.container-node .pill-emerald {
		background: rgba(56, 184, 132, 0.14);
		color: var(--vault-phosphor);
	}
	.route {
		font-size: var(--fs-xs);
		color: #b4c2bc;
	}
	.token {
		font-size: 10.5px;
		color: var(--vault-dim);
	}
	.links {
		width: 60px;
		height: 240px;
	}
	.links path {
		fill: none;
		stroke: var(--border-strong);
		stroke-width: 1.2;
		vector-effect: non-scaling-stroke;
		transition: stroke 300ms;
	}
	.links path.hot {
		stroke: var(--brand-emerald);
		stroke-width: 2;
	}
	.batches {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.batch {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 12px 14px;
		background: var(--surface-card);
		border: var(--hairline);
		border-radius: var(--radius-sm);
		transition: border-color 300ms, transform 500ms var(--ease-expo);
	}
	.batch:hover {
		border-color: var(--brand-emerald);
		transform: translateX(4px);
	}
	.b-top {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	.b-top .code {
		font-size: 1rem;
		color: var(--ink-primary);
	}
	.b-row {
		display: flex;
		justify-content: space-between;
		gap: 8px;
		font-size: 11.5px;
		color: var(--ink-secondary);
	}
	.views {
		padding: 22px 4px;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.views ul {
		display: flex;
		flex-direction: column;
	}
	.views li {
		display: grid;
		grid-template-columns: 18px 1fr auto;
		align-items: center;
		gap: 10px;
		padding: 11px 0;
		border-bottom: var(--hairline);
		font-size: var(--fs-sm);
		color: var(--ink-faint);
		transition: color 400ms var(--ease-out);
		transition-delay: calc(var(--i) * 25ms);
	}
	.views li.visible {
		color: var(--ink-primary);
	}
	.state {
		width: 12px;
		height: 12px;
		border-radius: 50%;
		border: 1px dashed var(--border-strong);
		transition: all 400ms var(--ease-expo);
		transition-delay: calc(var(--i) * 25ms);
	}
	.visible .state {
		border: 0;
		background: var(--brand-emerald);
		box-shadow: 0 0 0 3px var(--brand-emerald-tint);
	}
	.lvl {
		font-size: 10px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--ink-faint);
	}
	.fine {
		font-size: 11.5px;
		line-height: 1.55;
		color: var(--ink-muted);
	}
	@media (max-width: 1024px) {
		.qp {
			grid-template-columns: 1fr;
		}
	}
	@media (max-width: 620px) {
		.tree {
			grid-template-columns: 1fr;
			gap: 16px;
		}
		.links {
			display: none;
		}
		.container-node {
			flex-direction: row;
		}
	}
</style>
