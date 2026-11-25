<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { money } from '$lib/format';

	let tab = $state<'rfqs' | 'discover' | 'profile'>('rfqs');
	let q = $state('');
	let sector = $state('All');
	let tierF = $state('All');
	let sel = $state(ws.connect.rfqs[0]?.id);
	const rfq = $derived(ws.connect.rfqs.find((r) => r.id === sel) ?? ws.connect.rfqs[0]);
	const results = $derived(
		ws.connect.suppliers.filter((s) => (sector === 'All' || s.sector === sector) && (tierF === 'All' || s.tier === tierF) && (!q || `${s.name} ${s.city} ${s.cn.join(' ')}`.toLowerCase().includes(q.toLowerCase())))
	);
	const tierPill: Record<string, string> = { 'Self-Declared': 'pill-outline', 'Document-Reviewed': 'pill-cobalt', 'Independently Verified': 'pill-emerald' };
	let msg = $state('');
	let thread = $state([{ from: 'Lena Hoffmann', text: 'Can you hold the quoted price for 30 days and include the CBAM XML per batch?' }]);
</script>

<svelte:head><title>Zevra Connect — Zevra</title></svelte:head>

<PageHeader eyebrow="Zevra Connect · D1" title="Global B2B" serif="discovery" lede="Verified MSME profiles, carbon-enriched catalogues, structured RFQs and side-by-side quote comparison. Every claim carries a verification tier." />

<div class="tabs-line" role="tablist">
	<button role="tab" aria-selected={tab === 'rfqs'} onclick={() => (tab = 'rfqs')}>RFQs & quotes</button>
	<button role="tab" aria-selected={tab === 'discover'} onclick={() => (tab = 'discover')}>Discover suppliers</button>
	<button role="tab" aria-selected={tab === 'profile'} onclick={() => (tab = 'profile')}>My profile & catalogue</button>
</div>

{#if tab === 'rfqs'}
	<div class="g2">
		<section class="panel">
			{#each ws.connect.rfqs as r (r.id)}
				<button class="rfq" class:on={r.id === rfq.id} onclick={() => (sel = r.id)}>
					<div class="row"><span class="mono">{r.id}</span><span class="pill" class:pill-emerald={r.status === 'awarded'} class:pill-cobalt={r.status === 'negotiating'}>{r.status}</span></div>
					<strong>{r.buyer}</strong>
					<span class="small muted">{r.product} · {r.qty} t · {r.incoterm} · {r.quotes.length} quote{r.quotes.length === 1 ? '' : 's'}</span>
				</button>
			{/each}
		</section>
		<section class="panel">
			<div class="panel-head"><h2>{rfq.id} · {rfq.buyer}</h2><span class="mono small">{rfq.product} · {rfq.qty} t</span></div>
			{#if rfq.quotes.length}
				<div class="table-wrap">
					<table class="table">
						<thead><tr><th>Supplier</th><th class="num">Price / t</th><th class="num">Lead</th><th class="num">tCO₂e/t</th><th>Verification</th></tr></thead>
						<tbody>
							{#each [...rfq.quotes].sort((a, b) => a.price - b.price) as qq (qq.supplier)}
								<tr><td class="strong">{qq.supplier}</td><td class="num">{money(qq.price, ws.currency)}</td><td class="num">{qq.lead} d</td><td class="num">{qq.intensity}</td><td><span class="pill {tierPill[qq.tier]}">{qq.tier}</span></td></tr>
							{/each}
						</tbody>
					</table>
				</div>
			{:else}
				<div class="quote card">
					<p class="small">No quote yet. Submit a structured quote with your QR passport and emissions evidence attached.</p>
					<button class="btn btn-sm btn-primary" onclick={() => { rfq.quotes.push({ supplier: 'Deccan Alloy Works', price: 2985, lead: 21, intensity: 5.5, tier: 'Document-Reviewed' }); rfq.status = 'quoted'; }}>Quote $2,985/t · 21 days</button>
				</div>
			{/if}
			<div class="msgs">
				<span class="eyebrow">Structured messaging</span>
				{#each thread as m, i (i)}<div class="m" class:me={m.from === 'Deccan Alloy Works'}><span class="by">{m.from}</span>{m.text}</div>{/each}
				<div class="rep"><input class="input" bind:value={msg} placeholder="Message the buyer…" /><button class="btn btn-sm btn-dark" disabled={!msg} onclick={() => { thread.push({ from: 'Deccan Alloy Works', text: msg }); msg = ''; }}>Send</button></div>
			</div>
		</section>
	</div>
{:else if tab === 'discover'}
	<section class="panel">
		<div class="filters">
			<label class="search"><Icon name="search" size={14} /><input bind:value={q} placeholder="Product, CN code, city…" /></label>
			<select class="select" bind:value={sector}><option>All</option><option>Aluminium</option><option>Iron & Steel</option><option>Cement</option></select>
			<select class="select" bind:value={tierF}><option>All</option><option>Self-Declared</option><option>Document-Reviewed</option><option>Independently Verified</option></select>
		</div>
		<div class="cards">
			{#each results as s, i (s.name)}
				<article class="sup fade-in" style="--i:{i}">
					<div class="row"><strong>{s.name}</strong><span class="pill {tierPill[s.tier]}">{s.tier}</span></div>
					<span class="small muted">{s.city}, {s.country} · {s.sector} · {s.capacity}</span>
					<div class="cns">{#each s.cn as c (c)}<span class="pill pill-outline">CN {c}</span>{/each}</div>
					<div class="row"><span class="mono small">{s.intensity} tCO₂e/t</span><span class="small muted">{s.certs.join(' · ')}</span></div>
					<button class="btn btn-sm">Request quote</button>
				</article>
			{/each}
		</div>
		<p class="fine">Intermediaries are introduced only with explicit buyer and seller permission, and their role is always disclosed.</p>
	</section>
{:else}
	<div class="g2">
		<section class="panel">
			<div class="panel-head"><h2>Verification checklist</h2><span class="pill pill-cobalt">Document-Reviewed</span></div>
			{#each [['GSTIN & Udyam registration', true], ['Factory address & photos', true], ['ISO 9001 / 14001 certificates', true], ['Measured emissions on ≥ 2 SKUs', true], ['Independent verifier statement', false]] as [l, ok] (l)}
				<div class="ck" class:ok><span>{ok ? '✓' : '○'}</span>{l}</div>
			{/each}
		</section>
		<section class="panel">
			<div class="panel-head"><h2>Export catalogue</h2></div>
			{#each ws.products as p (p.id)}
				<div class="cat"><div><strong>{p.name}</strong><div class="mono small muted">CN {p.cn} · {(p.direct_tco2e_t + p.indirect_tco2e_t + p.precursor_tco2e_t).toFixed(2)} tCO₂e/t</div></div><label class="small"><input type="checkbox" bind:checked={p.published_connect} /> Listed</label></div>
			{/each}
		</section>
	</div>
{/if}

<style>
	.g2 {
		display: grid;
		grid-template-columns: 1fr 1.6fr;
		gap: 16px;
		align-items: start;
	}
	.small {
		font-size: 11.5px;
	}
	.strong {
		color: var(--ink-primary);
	}
	.rfq {
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: 4px;
		align-items: flex-start;
		padding: 12px;
		margin-bottom: 8px;
		border: var(--hairline);
		border-radius: var(--radius-sm);
		text-align: left;
		font-size: 13px;
	}
	.rfq.on {
		border-color: var(--brand-emerald);
		box-shadow: 0 0 0 3px var(--brand-emerald-tint);
	}
	.row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 8px;
		width: 100%;
	}
	.quote {
		padding: 14px;
		display: flex;
		flex-direction: column;
		gap: 10px;
		align-items: flex-start;
	}
	.msgs {
		margin-top: 18px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.m {
		max-width: 85%;
		padding: 9px 12px;
		border-radius: 8px;
		background: var(--surface-subtle);
		font-size: 12.5px;
		display: flex;
		flex-direction: column;
	}
	.m.me {
		align-self: flex-end;
		background: var(--brand-emerald-tint);
	}
	.by {
		font-size: 11px;
		font-weight: 600;
		color: var(--accent-cobalt);
	}
	.rep {
		display: flex;
		gap: 8px;
	}
	.filters {
		display: grid;
		grid-template-columns: 1fr 200px 220px;
		gap: 10px;
		margin-bottom: 16px;
	}
	.search {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 0 12px;
		border: var(--hairline);
		border-radius: var(--radius-sm);
		color: var(--ink-muted);
	}
	.search input {
		flex: 1;
		border: 0;
		outline: none;
		height: 40px;
		background: transparent;
	}
	.cards {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 12px;
	}
	.sup {
		padding: 16px;
		border: var(--hairline);
		border-radius: var(--radius-md);
		display: flex;
		flex-direction: column;
		gap: 10px;
		align-items: flex-start;
		font-size: 13px;
	}
	.cns {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
	}
	.ck {
		display: flex;
		gap: 10px;
		padding: 9px 0;
		border-bottom: var(--hairline);
		font-size: 13px;
		color: var(--ink-muted);
	}
	.ck.ok {
		color: var(--ink-primary);
	}
	.ck.ok span {
		color: var(--brand-emerald);
	}
	.cat {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 10px 0;
		border-bottom: var(--hairline);
		font-size: 13px;
		accent-color: var(--brand-emerald);
	}
	@media (max-width: 1000px) {
		.g2,
		.filters {
			grid-template-columns: 1fr;
		}
	}
</style>
