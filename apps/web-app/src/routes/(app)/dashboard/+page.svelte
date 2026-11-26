<script lang="ts">
	import { page } from '$app/state';
	import FactoryTwin from '$lib/components/factory3d/FactoryTwin.svelte';
	import Kpi from '$lib/components/ui/Kpi.svelte';
	import { slide } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { appearance } from '$lib/stores/appearance.svelte';
	import EChart from '$lib/components/charts/EChart.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { MONTHS } from '$lib/demo/seed';
	import { money, num, dueIn, timeAgo, QUALITY_LABEL, QUALITY_PILL } from '$lib/format';
	import { compareDefaultVsActual, CERT_PRICE_EUR } from '$lib/carbon';

	const monthKey = (i: number) => (i === 0 ? '2025-12' : `2026-${String(i).padStart(2, '0')}`);
	const byMonth = $derived(
		MONTHS.map((m, i) => {
			const v = ws.vouchers.filter((x) => x.date.startsWith(monthKey(i)));
			const revenue = v.filter((x) => x.type === 'sales').reduce((s, x) => s + x.amount_usd, 0);
			const cogs = v.filter((x) => x.type === 'sales').reduce((s, x) => s + x.cost_usd, 0);
			const opex = v.filter((x) => x.type === 'expense' || x.type === 'utility' || x.type === 'freight').reduce((s, x) => s + x.amount_usd, 0) * 0.35;
			const exp = v.filter((x) => x.type === 'sales' && x.is_export).reduce((s, x) => s + x.amount_usd, 0);
			return { m, revenue, gross: revenue - cogs, net: revenue - cogs - opex, exportRev: exp };
		})
	);
	const cur = $derived(byMonth[11]);
	const prev = $derived(byMonth[10]);
	const revDelta = $derived(((cur.revenue - prev.revenue) / prev.revenue) * 100);
	const grossMargin = $derived((cur.gross / cur.revenue) * 100);
	const exportShare = $derived(cur.exportRev / cur.revenue);

	const invSeg = $derived({
		raw: ws.inventory.filter((i) => i.category === 'raw_material').reduce((s, i) => s + i.qty * i.unit_value_usd, 0),
		wip: ws.inventory.filter((i) => i.category === 'wip').reduce((s, i) => s + i.qty * i.unit_value_usd, 0),
		fg: ws.inventory.filter((i) => i.category === 'finished_good').reduce((s, i) => s + i.qty * i.unit_value_usd, 0)
	});
	const invTotal = $derived(invSeg.raw + invSeg.wip + invSeg.fg);
	const lowStock = $derived(ws.inventory.filter((i) => i.qty < i.reorder).length);

	const activeBatches = $derived(ws.batches.filter((b) => b.stage === 'in_production' || b.stage === 'packaging'));
	const shiftPct = $derived(Math.round((ws.totalWorkers / ws.totalTarget) * 100));

	const exportTonnesYr = 420;
	const cbam = $derived(compareDefaultVsActual(exportTonnesYr, ws.products[0]));


	const fx = (v: number) => (ws.currency === 'INR' ? v * 88 : ws.currency === 'EUR' ? v * 0.86 : v);
	const trendOption = $derived({
		tooltip: { trigger: 'axis' },
		legend: { top: 0, right: 0, data: ['Revenue', 'Gross profit', 'Net profit'] },
		grid: { top: 34, left: 4, right: 4, bottom: 0, containLabel: true },
		xAxis: { type: 'category', data: byMonth.map((d) => d.m), boundaryGap: false },
		yAxis: { type: 'value', axisLabel: { formatter: (v: number) => money(v / (ws.currency === 'INR' ? 88 : ws.currency === 'EUR' ? 0.86 : 1), ws.currency, { compact: true }) } },
		series: [
			{ name: 'Revenue', type: 'line', smooth: 0.35, symbol: 'none', data: byMonth.map((d) => Math.round(fx(d.revenue))), lineStyle: { width: 2.5, color: '#8ABDE8' }, itemStyle: { color: '#8ABDE8' }, areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(106,184,255,0.25)' }, { offset: 1, color: 'rgba(106,184,255,0)' }] } } },
			{ name: 'Gross profit', type: 'line', smooth: 0.35, symbol: 'none', data: byMonth.map((d) => Math.round(fx(d.gross))), lineStyle: { width: 1.5, color: '#BAA0DE' }, itemStyle: { color: '#BAA0DE' } },
			{ name: 'Net profit', type: 'bar', barWidth: 10, data: byMonth.map((d) => Math.round(fx(d.net))), itemStyle: { color: '#CFDEED', borderRadius: [2, 2, 0, 0] } }
		]
	});

	const buyers = $derived(ws.notifications.filter((n) => n.category === 'buyer_agency').slice(0, 3));
	const deadlines = $derived(ws.notifications.filter((n) => n.category === 'deadline'));
	let twinOpen = $state(false);
	$effect(() => { if (page.url.searchParams.get('zone')) twinOpen = true; });
	const spark = $derived(byMonth.map((d, i) => `${i * 11},${44 - (d.revenue / Math.max(...byMonth.map((m) => m.revenue))) * 36}`).join(' '));
	const packaged = $derived(ws.batches.filter((b) => b.stage === 'packaging').length);
	function openZone(code: string) { twinOpen = true; ws.selectedZone = code; }

</script>
<svelte:head><title>Dashboard — ZENRES</title><meta name="description" content="Your manufacturing, finance and carbon workspace, beautifully connected." /></svelte:head>

<div class="overview-heading">
	<div><span class="location"><span class="dot"></span>Chakan Extrusion Plant</span><span class="overview-caption">Your business, at a glance.</span></div>
	<div class="overview-actions"><span class="period">November 2026 <span class="period-dot"></span> Month to date</span><button class="btn btn-sm" onclick={() => ws.simulateInbound()}><Icon name="plus" size={13}/> Simulate buyer email</button></div>
</div>

<div class="kpis">
	<Kpi spacious i={0} icon="finance" label="Revenue this month" value={money(cur.revenue, ws.currency, { compact: true })} delta="{revDelta >= 0 ? '↗' : '↘'} {Math.abs(revDelta).toFixed(1)}% vs last month" tone={revDelta >= 0 ? 'up' : 'down'} href="/finance">
		<svg class="sparkline" viewBox="0 0 121 48" aria-label="Revenue trend over 12 months"><defs><linearGradient id="revenue-wash" x1="0" y1="0" x2="0" y2="1"><stop stop-color="#a0d0f3" stop-opacity=".65"/><stop offset="1" stop-color="#a0d0f3" stop-opacity="0"/></linearGradient></defs><polygon points="0,48 {spark} 121,48" fill="url(#revenue-wash)"/><polyline points={spark} fill="none" stroke="#91bde3" stroke-width="1.7" stroke-linejoin="round"/></svg>
		<div class="mini-caption"><span>Dec</span><span>Nov</span></div>
	</Kpi>
	<Kpi spacious i={1} icon="simulator" label="Profit overview" value={money(cur.net, ws.currency, { compact: true })} delta="Net profit this month" tone="flat" href="/finance">
		<dl class="metric-facts"><div><dt>Gross margin</dt><dd>{grossMargin.toFixed(1)}%</dd></div><div><dt>Export share</dt><dd>{Math.round(exportShare * 100)}%</dd></div></dl>
	</Kpi>
	<Kpi spacious i={2} icon="inventory" label="Inventory value" value={money(invTotal, ws.currency, { compact: true })} delta="{lowStock} items below reorder" tone={lowStock ? 'down' : 'flat'} href="/inventory">
		<dl class="metric-facts"><div><dt>Raw materials</dt><dd>{money(invSeg.raw, ws.currency, { compact: true })}</dd></div><div><dt>Finished goods</dt><dd>{money(invSeg.fg, ws.currency, { compact: true })}</dd></div></dl>
	</Kpi>
	<Kpi spacious i={3} icon="users" label="On the factory floor" value={String(ws.totalWorkers).padStart(2, '0')} delta="Team members on shift" href="/plants/plt_chakan/zones">
		<dl class="metric-facts"><div><dt>Active batches</dt><dd>{String(activeBatches.length).padStart(2, '0')}</dd></div><div><dt>Packaging</dt><dd>{String(packaged).padStart(2, '0')}</dd></div></dl><div class="bar workforce"><span style="width:{shiftPct}%"></span></div>
	</Kpi>
	<Kpi spacious i={4} icon="cbam" label="Carbon savings potential" value={money(cbam.savingsEur / 0.86, ws.currency, { compact: true })} delta="Estimated vs default values" href="/carbon/cbam">
		<dl class="metric-facts"><div><dt>Export volume</dt><dd>{exportTonnesYr} t/yr</dd></div><div><dt>Reporting</dt><dd><span class="pill pill-cobalt">Estimated</span></dd></div></dl>
	</Kpi>
	<Kpi spacious i={5} icon="tally" label="Accounts & orders" value="04" delta="Open export orders" href="/integrations/tally">
		<dl class="metric-facts"><div><dt>Receivable</dt><dd>{money(181500, ws.currency, { compact: true })}</dd></div><div><dt>Payable</dt><dd>{money(127800, ws.currency, { compact: true })}</dd></div></dl>
	</Kpi>
</div>

<div class="section-heading"><div><span class="section-kicker">A closer look</span><h2>Everything is connected.</h2></div><a href="/carbon/cbam" class="btn btn-sm">Open CBAM workbench <span class="arrow">↗</span></a></div>
<div class="row2">
	<section class="panel fade-in" style="--i:2">
		<div class="panel-head"><div><h2>Business performance</h2><p class="panel-sub">Revenue and profitability over the past year</p></div><span class="pill pill-outline">{ws.currency} · 12 months</span></div>
		<EChart option={trendOption} height={255}/>
		<div class="chart-foot"><span>Connected to Tally vouchers</span><a href="/finance">View financials ↗</a></div>
	</section>
	<section class="panel fade-in" style="--i:3">
		<div class="panel-head"><h2>Your inbox</h2><button class="inbox-count" onclick={() => (ws.bellOpen = true)} aria-label="Open notifications">{ws.unread} new</button></div>
		<ul class="mini" role="list">{#each buyers as n, i (n.id)}<li><a href="/inbox?thread={n.id}"><span class="sender" class:lilac={i % 2}>{n.org_badge?.slice(0, 1) ?? 'B'}</span><span class="message"><span class="org">{n.org_badge?.split(' · ')[0]}</span><span class="subj">{n.subject}</span><span class="time">{timeAgo(n.received_at)}</span></span><span class="message-arrow">↗</span></a></li>{/each}</ul>
		<a class="inbox-link" href="/inbox">See all messages <span>→</span></a>
	</section>
</div>

<section class="panel production-panel fade-in" style="--i:4">
	<div class="panel-head"><div><h2>Production in motion</h2><p class="panel-sub">Live batches, from the factory floor to dispatch</p></div><a class="btn btn-sm" href="/trace/qr"><Icon name="qr" size={14}/> QR studio</a></div>
	<div class="table-wrap"><table class="table"><thead><tr><th>Batch reference</th><th>Product / SKU</th><th>Zone</th><th>Status</th><th>Production progress</th><th class="num">tCO₂e (S1 + S2)</th><th>Data quality</th></tr></thead><tbody>
		{#each ws.batches.filter((b) => b.stage !== 'shipped') as b (b.id)}
			{@const p = ws.products.find((x) => x.id === b.product_id)}
			{@const prog = b.stage === 'packaging' || b.stage === 'ready_dispatch' ? b.packaged_qty / b.target_qty : b.produced_qty / b.target_qty}
			<tr><td class="strong batch-ref"><span class="batch-icon"><Icon name="box" size={14}/></span>{b.code}</td><td>{p?.sku}<div class="muted small">CN {p?.cn}</div></td><td><button class="zone-link" onclick={() => openZone(b.zone_code)}>{b.zone_code} ↗</button></td><td><span class="pill" class:pill-emerald={b.stage === 'ready_dispatch'} class:pill-cobalt={b.stage === 'packaging'}>{b.stage.replaceAll('_', ' ')}</span></td><td style="min-width:160px"><div class="prog"><div class="bar"><span style="width:{prog * 100}%"></span></div><span class="small">{Math.round(prog * 100)}%</span></div></td><td class="num">{num(b.direct_tco2e + b.indirect_tco2e, 2)}</td><td><span class="pill {QUALITY_PILL[b.quality]}">{QUALITY_LABEL[b.quality]}</span></td></tr>
		{/each}
	</tbody></table></div>
	<p class="fine table-note">CBAM figures are illustrative estimates at €{CERT_PRICE_EUR}/tCO₂e until verified. Logistics emissions are reported separately.</p>
</section>

<div class="bottom-row">
	<section class="twin-section">
		<button class="twin-toggle panel" aria-expanded={twinOpen} aria-controls="factory-twin" onclick={() => (twinOpen = !twinOpen)}><span class="twin-symbol"><Icon name="plant" size={23}/></span><span class="twin-description"><strong>A window into your factory</strong><span>Explore zones, production lines, and live activity in 3D.</span></span><span class="twin-action">{twinOpen ? 'Close view' : 'Open digital twin'} <span class:rotated={twinOpen}>↗</span></span></button>
	</section>
	{#if deadlines.length}<a class="deadline panel" href={deadlines[0].link?.href ?? '/carbon/cbam'}><span class="deadline-icon"><Icon name="clock" size={20}/></span><div><span class="section-kicker">On your horizon</span><strong>{deadlines[0].subject}</strong></div><span class="pill pill-outline">{dueIn(deadlines[0].deadline!).label}</span></a>{/if}
</div>
{#if twinOpen}<div id="factory-twin" class="twin-container" transition:slide={{ duration: appearance.motion ? 500 : 0, easing: cubicOut }}><FactoryTwin initialZone={page.url.searchParams.get('zone')}/></div>{/if}

<style>
	.overview-heading { display: flex; justify-content: space-between; gap: 20px; align-items: flex-start; padding-top: 5px; margin-bottom: 72px; }
	.overview-heading > div:first-child { display: flex; flex-direction: column; gap: 5px; }
	.location { display: flex; align-items: center; gap: 7px; font-size: 12px; color: #64738a; }
	.location .dot { width: 5px; height: 5px; color: #8c9fc0; }
	.overview-caption { color: #7b8796; font-size: 12px; padding-left: 12px; }
	.overview-actions { display: flex; gap: 20px; align-items: center; }
	.period { display: flex; align-items: center; gap: 8px; font-size: 12px; color: #6e7b8f; }
	.period-dot { width: 3px; height: 3px; background: #8d9aae; border-radius: 50%; }
	.kpis { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 18px; }
	.sparkline { height: 43px; width: 100%; }
	.mini-caption { display: flex; justify-content: space-between; font-size: 9px; color: #8a97a8; margin-top: 3px; }
	.metric-facts { display: flex; flex-direction: column; gap: 8px; }
	.metric-facts > div { display: flex; align-items: center; justify-content: space-between; gap: 12px; font-size: 11px; }
	.metric-facts dt { color: #7a8796; white-space: nowrap; }
	.metric-facts dd { color: #536174; font-variant-numeric: tabular-nums; white-space: nowrap; }
	.metric-facts .pill { height: 19px; font-size: 9px; }
	.workforce { margin-top: 10px; }
	.section-heading { display: flex; align-items: flex-end; justify-content: space-between; gap: 18px; padding: 36px 2px 21px; }
	.section-kicker { color: #6e7c90; font-size: 11px; display: block; margin-bottom: 5px; }
	.section-heading h2 { font-size: 25px; font-weight: 450; letter-spacing: -.035em; }
	.row2 { display: grid; grid-template-columns: 1.8fr 1fr; gap: 18px; margin-bottom: 18px; }
	.panel-sub { font-size: 12px; color: var(--ink-muted); margin-top: 6px; }
	.chart-foot { border-top: var(--hairline); padding-top: 15px; margin-top: 15px; display: flex; justify-content: space-between; gap: 10px; font-size: 11px; color: #7b899c; }
	.chart-foot a { color: #617eab; }
	.inbox-count { font-size: 10px; background: #eef0fb; border: 1px solid #e5e7f8; color: #7b76a1; border-radius: 20px; padding: 4px 10px; }
	.mini { display: flex; flex-direction: column; }
	.mini li a { display: flex; align-items: flex-start; gap: 11px; padding: 15px 0; border-bottom: var(--hairline); transition: transform 250ms var(--ease-out); }
	.mini li:first-child a { padding-top: 4px; }
	.mini li a:hover { transform: translateX(3px); }
	.sender { display: grid; place-items: center; flex: none; width: 34px; height: 34px; background: #e7f0f9; border-radius: 50%; color: #7696ba; font-size: 13px; }
	.sender.lilac { background: #f0e9f7; color: #9a85b6; }
	.message { display: flex; flex-direction: column; min-width: 0; gap: 4px; }
	.org { font-size: 12px; font-weight: 600; color: #4f5b6d; }
	.subj { font-size: 12px; line-height: 1.4; color: #6d7b8e; }
	.time { font-size: 10px; color: #92a0b2; }
	.message-arrow { margin-left: auto; color: #9ca9bb; font-size: 16px; }
	.inbox-link { display: flex; justify-content: space-between; margin-top: 17px; font-size: 12px; color: #6c7b93; }
	.production-panel { margin-bottom: 18px; }
	.batch-ref { white-space: nowrap; }
	.batch-icon { display: inline-flex; vertical-align: middle; margin-right: 9px; background: #edf2f9; border-radius: 8px; padding: 7px; color: #8a9ab0; }
	.zone-link { color: #7b8eae; background: #edf1f7; padding: 3px 9px; border-radius: 30px; font-size: 11px; }
	.prog { display: flex; align-items: center; gap: 12px; }
	.prog .bar { flex: 1; }
	.small { font-size: 11px; }
	.table-note { border-top: var(--hairline); margin-top: 10px; padding-top: 16px; font-size: 10px; color: #8b97a7; }
	.bottom-row { display: grid; grid-template-columns: 1.3fr 1fr; gap: 18px; }
	.twin-toggle { display: flex; align-items: center; width: 100%; height: 100%; gap: 15px; text-align: left; transition: transform 400ms var(--ease-out), box-shadow 400ms; }
	.twin-toggle:hover { transform: translateY(-3px); box-shadow: 0 12px 30px -20px #52678860; }
	.twin-symbol, .deadline-icon { width: 48px; height: 48px; flex: none; display: grid; place-items: center; color: #93a5c3; background: linear-gradient(120deg, #e0effa, #eee7f8); border-radius: 14px; }
	.twin-description { display: flex; flex-direction: column; gap: 6px; }
	.twin-description strong { font-size: 16px; font-weight: 550; }
	.twin-description > span { font-size: 11px; color: #7c899b; }
	.twin-action { margin-left: auto; font-size: 11px; white-space: nowrap; color: #6b7c96; }
	.twin-action > span { display: inline-block; font-size: 18px; margin-left: 10px; transition: transform 350ms var(--ease-out); }
	.twin-action > span.rotated { transform: rotate(180deg); }
	.deadline { display: flex; gap: 13px; align-items: center; }
	.deadline strong { display: block; font-size: 12px; font-weight: 500; line-height: 1.4; }
	.deadline .pill { margin-left: auto; font-size: 10px; }
	.twin-container { margin-top: 18px; }
	@media (max-width: 1150px) { .kpis { grid-template-columns: repeat(2, minmax(0, 1fr)); } .overview-heading { margin-bottom: 42px; } .period { display: none; } .bottom-row { grid-template-columns: 1fr; } }
	@media (max-width: 850px) { .row2 { grid-template-columns: 1fr; } }
	@media (max-width: 650px) { .overview-heading { margin-bottom: 34px; align-items: center; gap: 10px; } .overview-actions .btn { font-size: 10px; padding: 0 10px; } .overview-caption { display: none; } .location { font-size: 10px; } .kpis { grid-template-columns: 1fr; gap: 12px; } .section-heading h2 { font-size: 21px; } .section-heading .btn { font-size: 10px; padding: 0 12px; } .twin-toggle { flex-wrap: wrap; } .twin-action { width: 100%; padding-left: 63px; } .deadline { flex-wrap: wrap; } }
</style>
