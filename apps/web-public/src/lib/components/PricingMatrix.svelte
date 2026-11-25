<script lang="ts">
	import { TIERS, COMPARISON, DISCLAIMERS, FX, type Billing, type Currency } from '$lib/content';
	import Segmented from '$design/components/Segmented.svelte';
	import { reveal, tilt, tweenNumber } from '$design/motion';
	import { untrack } from 'svelte';

	interface Props {
		expanded?: boolean;
	}
	let { expanded = false }: Props = $props();

	let billing = $state<Billing>('monthly');
	let currency = $state<Currency>('USD');
	let showMatrix = $state(false);
	$effect(() => {
		if (expanded) showMatrix = true;
	});

	// animated displayed amounts per tier
	let shown = $state(TIERS.map((t) => t.monthly));
	let stops: (() => void)[] = [];
	$effect(() => {
		const targets = TIERS.map((t) => (billing === 'monthly' ? t.monthly : t.annual) * FX[currency].rate);
		stops.forEach((s) => s());
		stops = untrack(() => targets.map((to, i) => tweenNumber(shown[i], to, 700, (v) => (shown[i] = v))));
	});

	const fmt = (v: number) => `${FX[currency].symbol}${Math.round(v).toLocaleString(FX[currency].locale)}`;
</script>

<div class="pricing">
	<div class="controls" use:reveal>
		<Segmented
			label="Billing cycle"
			bind:value={billing}
			options={[
				{ value: 'monthly', label: 'Monthly' },
				{ value: 'annual', label: 'Annual — 2 months free' }
			]}
		/>
		<Segmented
			label="Display currency"
			bind:value={currency}
			options={[
				{ value: 'USD', label: 'USD' },
				{ value: 'INR', label: 'INR' },
				{ value: 'EUR', label: 'EUR' }
			]}
		/>
	</div>

	<div class="cards">
		{#each TIERS as tier, i (tier.id)}
			<article class="tier" class:flagship={tier.flagship} id={tier.id} use:reveal={{ delay: i * 90 }}>
				<div class="tier-inner" use:tilt={2.5}>
					<div class="tier-head">
						<span class="eyebrow">{String(i + 1).padStart(2, '0')} — {tier.tagline}</span>
						{#if tier.flagship}<span class="pill pill-emerald">Flagship</span>{/if}
					</div>
					<h3>{tier.name}</h3>
					<div class="price">
						<span class="amount num">{fmt(shown[i])}</span>
						<span class="per mono">/ {billing === 'monthly' ? 'mo' : 'yr'}</span>
					</div>
					<p class="alt mono">
						{#if billing === 'monthly'}
							or {fmt(tier.annual * FX[currency].rate)} billed annually
						{:else}
							equivalent to {fmt((tier.annual / 12) * FX[currency].rate)} / mo
						{/if}
						{#if currency !== 'USD'}<span class="ind">· indicative</span>{/if}
					</p>
					<p class="segment">{tier.segment}</p>
					<a href="/signup?plan={tier.id}" class="btn {tier.flagship ? 'btn-primary' : 'btn-dark'} btn-lg cta" data-cursor="Select">
						{tier.id === 'importer_hub' ? 'Open Importer Hub' : 'Start with ' + tier.name.replace('MSME ', '')}
						<span class="arrow">→</span>
					</a>
					<ul role="list" class="feats">
						{#each tier.features as f (f)}
							<li><span class="tick" aria-hidden="true"></span>{f}</li>
						{/each}
					</ul>
					<div class="addon"><span class="eyebrow">Add-ons & setup</span><p>{tier.addons}</p></div>
				</div>
			</article>
		{/each}
	</div>

	<div class="disclosure" use:reveal>
		<span class="pill pill-cobalt">Commercial disclosure</span>
		<p>{DISCLAIMERS.commercial}</p>
	</div>

	<div class="matrix-wrap" id="compare">
		<button class="matrix-toggle" onclick={() => (showMatrix = !showMatrix)} aria-expanded={showMatrix}>
			<span class="eyebrow">Plan comparison matrix — 14 dimensions</span>
			<span class="plus" class:open={showMatrix} aria-hidden="true"></span>
		</button>
		<div class="matrix" class:open={showMatrix}>
			<div class="matrix-inner">
				<div class="table-scroll">
					<table class="table">
						<thead>
							<tr>
								<th>Dimension</th>
								{#each TIERS as t (t.id)}<th>{t.name}</th>{/each}
							</tr>
						</thead>
						<tbody>
							{#each COMPARISON as row (row.dimension)}
								<tr>
									<td class="strong">{row.dimension}</td>
									{#each row.values as v, j (j)}
										<td class:dash={v === '—' || v === 'Not included'}>{v}</td>
									{/each}
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		</div>
		<p class="fine">{DISCLAIMERS.pricing}</p>
	</div>
</div>

<style>
	.controls {
		display: flex;
		justify-content: space-between;
		gap: 12px;
		flex-wrap: wrap;
		margin-bottom: 28px;
	}
	.cards {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: var(--gap);
		align-items: stretch;
	}
	.tier {
		display: flex;
	}
	.tier-inner {
		position: relative;
		flex: 1;
		display: flex;
		flex-direction: column;
		padding: 28px;
		background: var(--surface-card);
		border: var(--hairline);
		border-radius: var(--radius-md);
		overflow: hidden;
	}
	.tier-inner::before {
		content: '';
		position: absolute;
		inset: 0;
		background: radial-gradient(420px circle at var(--mx, 50%) var(--my, 0%), rgba(20, 83, 60, 0.06), transparent 60%);
		opacity: 0;
		transition: opacity 400ms;
		pointer-events: none;
	}
	.tier-inner:hover::before {
		opacity: 1;
	}
	.flagship .tier-inner {
		background: var(--obsidian);
		border-color: var(--obsidian);
		color: #b9c6c0;
	}
	.flagship .tier-inner::before {
		background: radial-gradient(420px circle at var(--mx, 50%) var(--my, 0%), rgba(56, 184, 132, 0.14), transparent 60%);
	}
	.flagship h3,
	.flagship .amount {
		color: var(--ink-inverse);
	}
	.flagship .eyebrow {
		color: var(--vault-dim);
	}
	.flagship .pill-emerald {
		background: rgba(56, 184, 132, 0.14);
		color: var(--vault-phosphor);
	}
	.tier-head {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 8px;
		min-height: 22px;
	}
	h3 {
		margin-top: 20px;
		font-size: 1.6rem;
		letter-spacing: -0.025em;
	}
	.price {
		margin-top: 18px;
		display: flex;
		align-items: baseline;
		gap: 6px;
	}
	.amount {
		font-size: clamp(2.6rem, 4vw, 3.4rem);
		font-weight: 400;
		letter-spacing: -0.05em;
		color: var(--ink-primary);
		line-height: 1;
	}
	.per {
		font-size: 13px;
		color: var(--ink-muted);
	}
	.alt {
		margin-top: 8px;
		font-size: 11.5px;
		color: var(--ink-muted);
	}
	.ind {
		color: var(--status-terracotta);
	}
	.segment {
		margin-top: 16px;
		font-size: var(--fs-sm);
		min-height: 3em;
	}
	.cta {
		margin-top: 20px;
		width: 100%;
	}
	.flagship .cta {
		--btn-bd: var(--vault-phosphor);
	}
	.feats {
		margin-top: 24px;
		padding-top: 20px;
		border-top: var(--hairline);
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.flagship .feats,
	.flagship .addon {
		border-color: var(--obsidian-line);
	}
	.feats li {
		display: flex;
		gap: 10px;
		font-size: var(--fs-sm);
		line-height: 1.4;
		color: var(--ink-primary);
	}
	.flagship .feats li {
		color: #dfe7e3;
	}
	.tick {
		width: 14px;
		height: 14px;
		flex: none;
		margin-top: 2px;
		border-radius: 3px;
		background: var(--brand-emerald-tint)
			url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='8' height='6' fill='none'%3E%3Cpath stroke='%2314533C' stroke-width='1.5' d='m1 3 2 2 4-4'/%3E%3C/svg%3E")
			center no-repeat;
	}
	.flagship .tick {
		background-color: rgba(56, 184, 132, 0.16);
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='8' height='6' fill='none'%3E%3Cpath stroke='%2338B884' stroke-width='1.5' d='m1 3 2 2 4-4'/%3E%3C/svg%3E");
	}
	.addon {
		margin-top: auto;
		padding-top: 18px;
		border-top: 1px dashed var(--border-hairline);
		font-size: var(--fs-xs);
	}
	.feats + .addon {
		margin-top: 24px;
	}
	.addon p {
		margin-top: 4px;
	}
	.disclosure {
		margin-top: 28px;
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 16px 20px;
		background: var(--surface-subtle);
		border: var(--hairline);
		border-radius: var(--radius-md);
		font-size: var(--fs-sm);
	}
	.matrix-wrap {
		margin-top: 28px;
		border-top: var(--hairline);
	}
	.matrix-toggle {
		width: 100%;
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 22px 0;
	}
	.matrix-toggle .eyebrow {
		color: var(--ink-primary);
	}
	.plus {
		position: relative;
		width: 14px;
		height: 14px;
	}
	.plus::before,
	.plus::after {
		content: '';
		position: absolute;
		left: 0;
		right: 0;
		top: 50%;
		height: 1.5px;
		background: var(--ink-primary);
		transition: transform 500ms var(--ease-expo);
	}
	.plus::after {
		transform: rotate(90deg);
	}
	.plus.open::after {
		transform: rotate(0);
	}
	.matrix {
		display: grid;
		grid-template-rows: 0fr;
		transition: grid-template-rows 700ms var(--ease-expo);
	}
	.matrix.open {
		grid-template-rows: 1fr;
	}
	.matrix-inner {
		overflow: hidden;
	}
	.table-scroll {
		overflow-x: auto;
		border: var(--hairline);
		border-radius: var(--radius-md);
		background: var(--surface-card);
	}
	.table {
		min-width: 860px;
	}
	.table td {
		font-size: var(--fs-xs);
		line-height: 1.5;
	}
	.table td.dash {
		color: var(--ink-faint);
	}
	.table th:not(:first-child),
	.table td:not(:first-child) {
		width: 28%;
	}
	.fine {
		margin-top: 14px;
		font-size: 11.5px;
		color: var(--ink-muted);
	}
	@media (max-width: 1024px) {
		.cards {
			grid-template-columns: 1fr;
		}
		.segment {
			min-height: 0;
		}
	}
	@media (max-width: 560px) {
		.disclosure {
			flex-direction: column;
			align-items: flex-start;
		}
	}
</style>
