<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { env } from '$env/dynamic/public';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { QUALITY_LABEL, QUALITY_PILL, fmtDate } from '$lib/format';

	const SITE = env.PUBLIC_SITE_URL ?? 'http://localhost:5173';
	let qr = $state<Record<string, string>>({});
	let selC = $state(page.url.searchParams.get('container') ?? 'ZV-001');
	let view = $state<'public' | 'partner' | 'private'>('public');
	let events = $state([
		{ at: '2026-11-17T14:10:00Z', v: 1, text: 'A-101, A-102, A-103 packed into ZV-001 (18.0 t)' },
		{ at: '2026-11-18T09:42:00Z', v: 2, text: 'ZV-001 Merkle root anchored' },
		{ at: '2026-11-19T08:05:00Z', v: 3, text: 'Correction: A-102 qty 5.6 → 5.5 t (supersedes v1, original retained)' }
	]);
	const container = $derived(ws.containers.find((c) => c.code === selC) ?? ws.containers[0]);
	const linked = $derived(ws.batches.filter((b) => container.batches.includes(b.code)));
	const packable = $derived(ws.batches.filter((b) => (b.stage === 'ready_dispatch' || b.stage === 'packaging') && !ws.containers.some((c) => c.batches.includes(b.code))));

	async function gen(token: string, type: 'p' | 'c') {
		if (qr[token]) return;
		const Q = await import('qrcode');
		qr[token] = await Q.toString(`${SITE}/verify/${type}/${token}`, { type: 'svg', margin: 0, color: { dark: '#0E1412', light: '#FFFFFF' } });
	}
	onMount(() => {
		ws.containers.forEach((c) => gen(c.token, 'c'));
		ws.batches.forEach((b) => gen(b.qr_token, 'p'));
	});

	function pack(code: string) {
		container.batches.push(code);
		events.unshift({ at: new Date().toISOString(), v: events.length + 1, text: `${code} packed into ${container.code}` });
		ws.toast(`${code} linked to ${container.code}`);
	}
	function revoke() {
		container.status = 'revoked';
		events.unshift({ at: new Date().toISOString(), v: events.length + 1, text: `${container.code} label revoked (damaged) — public resolver now shows revocation` });
		ws.toast(`${container.code} label revoked`, 'warn');
	}
	function reprint() {
		container.status = 'reprinted';
		events.unshift({ at: new Date().toISOString(), v: events.length + 1, text: `${container.code} label reprinted with new token` });
	}
	function anchor() {
		container.anchored_at = new Date().toISOString();
		container.merkle_root = Array.from(crypto.getRandomValues(new Uint8Array(32)), (b) => b.toString(16).padStart(2, '0')).join('');
		events.unshift({ at: container.anchored_at, v: events.length + 1, text: `${container.code} Merkle root anchored` });
	}
	const fields = { public: ['Manufacturer', 'Product & CN', 'Batch & date', 'Carbon badge'], partner: ['+ Linked batches & qty', '+ Milestones & documents', '+ Calculation summary'], private: ['+ BOM & precursor lots', '+ Meter evidence & costing', '+ Audit hash trail'] };
</script>

<svelte:head><title>QR Studio — Zevra</title></svelte:head>

<PageHeader eyebrow="Zevra Trace · C1" title="Dual QR identity &" serif="traceability studio" lede="Product/Batch and Container/Shipment QR passports with non-guessable tokens, versioned corrections and selective off-chain hash anchoring.">
	{#snippet actions()}
		<button class="btn btn-sm" onclick={() => window.print()}><Icon name="download" size={14} />Print label sheet (Typst PDF)</button>
	{/snippet}
</PageHeader>

<div class="g2">
	<section class="panel">
		<div class="panel-head">
			<div class="ctabs">{#each ws.containers as c (c.id)}<button class:on={selC === c.code} onclick={() => (selC = c.code)}>{c.code}</button>{/each}</div>
			<span class="pill" class:pill-emerald={container.status === 'active'} class:pill-crimson={container.status === 'revoked'} class:pill-cobalt={container.status === 'reprinted'}>{container.status}</span>
		</div>
		<div class="tree">
			<div class="cnode">
				<div class="qr">{@html qr[container.token] ?? ''}</div>
				<div class="cm">
					<span class="mono code">{container.code}</span>
					<span class="small">{container.dispatch} · {container.route}</span>
					<span class="mono tok">{container.token}</span>
					<a class="link-u small" href="{SITE}/verify/c/{container.token}" target="_blank" rel="noopener">Open public resolver ↗</a>
				</div>
			</div>
			<div class="batches">
				{#each linked as b (b.id)}
					{@const p = ws.products.find((x) => x.id === b.product_id)}
					<div class="bnode">
						<div class="qr sm">{@html qr[b.qr_token] ?? ''}</div>
						<div>
							<div class="row"><span class="mono strong">{b.code}</span><span class="pill {QUALITY_PILL[b.quality]}">{QUALITY_LABEL[b.quality]}</span></div>
							<div class="small muted">{p?.name} · CN {p?.cn} · {b.packaged_qty} t</div>
						</div>
					</div>
				{:else}<p class="muted small">No batches packed yet.</p>{/each}
			</div>
		</div>
		{#if packable.length}
			<div class="packer">
				<span class="eyebrow">Batch-to-container packer</span>
				<div class="chips">{#each packable as b (b.id)}<button class="btn btn-sm" onclick={() => pack(b.code)}>+ {b.code}</button>{/each}</div>
			</div>
		{/if}
		<div class="acts">
			<button class="btn btn-sm btn-ghost" onclick={reprint}>Reprint label</button>
			<button class="btn btn-sm btn-ghost" onclick={revoke}>Revoke damaged label</button>
		</div>
	</section>

	<div class="side">
		<section class="panel">
			<div class="panel-head"><h2>Role-based view preview</h2></div>
			<div class="vtabs">{#each ['public', 'partner', 'private'] as v (v)}<button class:on={view === v} onclick={() => (view = v as typeof view)}>{v}</button>{/each}</div>
			<ul class="fields" role="list">
				{#each [...fields.public, ...(view !== 'public' ? fields.partner : []), ...(view === 'private' ? fields.private : [])] as f (f)}<li>{f}</li>{/each}
			</ul>
		</section>
		<section class="panel">
			<div class="panel-head"><h2>Selective hash anchoring</h2></div>
			{#if container.merkle_root}
				<p class="small">Anchored {fmtDate(container.anchored_at!)}</p>
				<code class="hash">{container.merkle_root}</code>
			{:else}<p class="small muted">Not anchored yet.</p>{/if}
			<button class="btn btn-sm btn-dark" onclick={anchor}>Anchor current state</button>
			<p class="fine">Anchoring proves post-anchor immutability, not source-input truthfulness. Only hashes leave Zevra — never business data.</p>
		</section>
		<section class="panel">
			<div class="panel-head"><h2>Versioned event log</h2></div>
			<ol class="ev" role="list">{#each events as e (e.v + e.at)}<li><span class="mono">v{e.v}</span><div><p>{e.text}</p><span class="mono small muted">{fmtDate(e.at)}</span></div></li>{/each}</ol>
		</section>
	</div>
</div>

<style>
	.g2 {
		display: grid;
		grid-template-columns: 1.5fr 1fr;
		gap: 16px;
		align-items: start;
	}
	.side {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.ctabs,
	.vtabs {
		display: flex;
		gap: 4px;
	}
	.ctabs button,
	.vtabs button {
		height: 28px;
		padding: 0 12px;
		border-radius: 3px;
		font-family: var(--font-mono);
		font-size: 11.5px;
		color: var(--ink-muted);
		border: var(--hairline);
		text-transform: capitalize;
	}
	.ctabs button.on,
	.vtabs button.on {
		background: var(--obsidian);
		color: var(--ink-inverse);
		border-color: var(--obsidian);
	}
	.tree {
		display: grid;
		grid-template-columns: 1fr 1.2fr;
		gap: 16px;
		align-items: center;
	}
	.cnode {
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding: 18px;
		background: var(--obsidian);
		color: var(--ink-inverse);
		border-radius: var(--radius-md);
	}
	.qr {
		width: 120px;
		aspect-ratio: 1;
		padding: 8px;
		background: #fff;
		border-radius: var(--radius-sm);
	}
	.qr.sm {
		width: 56px;
		padding: 4px;
		border: var(--hairline);
	}
	.qr :global(svg) {
		width: 100%;
		height: 100%;
	}
	.cm {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.code {
		font-size: 1.4rem;
	}
	.tok {
		font-size: 10px;
		color: var(--vault-dim);
		word-break: break-all;
	}
	.cnode .link-u {
		color: var(--vault-phosphor);
	}
	.batches {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.bnode {
		display: grid;
		grid-template-columns: 56px 1fr;
		gap: 12px;
		align-items: center;
		padding: 10px;
		border: var(--hairline);
		border-radius: var(--radius-sm);
		animation: zv-fade-in 500ms var(--ease-expo);
	}
	.row {
		display: flex;
		justify-content: space-between;
		gap: 6px;
		align-items: center;
	}
	.strong {
		color: var(--ink-primary);
	}
	.small {
		font-size: 11.5px;
	}
	.packer {
		margin-top: 16px;
		padding-top: 14px;
		border-top: var(--hairline);
	}
	.chips {
		display: flex;
		gap: 6px;
		margin-top: 8px;
	}
	.acts {
		display: flex;
		gap: 8px;
		justify-content: flex-end;
		margin-top: 14px;
	}
	.vtabs {
		margin-bottom: 10px;
	}
	.fields li {
		padding: 7px 0;
		border-bottom: var(--hairline);
		font-size: 12.5px;
		color: var(--ink-primary);
		animation: zv-fade-in 400ms var(--ease-expo);
	}
	.hash {
		display: block;
		margin: 8px 0 12px;
		padding: 8px;
		font-size: 10.5px;
		word-break: break-all;
		background: var(--surface-canvas);
		border: var(--hairline);
		border-radius: 3px;
	}
	.ev li {
		display: grid;
		grid-template-columns: 30px 1fr;
		gap: 8px;
		padding: 8px 0;
		border-bottom: var(--hairline);
		font-size: 12.5px;
	}
	.ev .mono:first-child {
		color: var(--brand-emerald);
		font-size: 11px;
	}
	@media (max-width: 1100px) {
		.g2,
		.tree {
			grid-template-columns: 1fr;
		}
	}
</style>
