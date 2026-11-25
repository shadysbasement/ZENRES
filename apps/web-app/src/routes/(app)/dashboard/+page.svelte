<script lang="ts">
	import { page } from '$app/state';
	import FactoryTwin from '$lib/components/factory3d/FactoryTwin.svelte';
	import Kpi from '$lib/components/ui/Kpi.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import EChart from '$lib/components/charts/EChart.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { MONTHS, USER } from '$lib/demo/seed';
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

	const costMix = [
		['Material', 74, 'var(--ink-secondary)'],
		['Energy', 14, 'var(--brand-emerald)'],
		['Labour', 7, 'var(--accent-cobalt)'],
		['Freight', 5, 'var(--status-terracotta)']
	] as const;

	const fx = (v: number) => (ws.currency === 'INR' ? v * 88 : ws.currency === 'EUR' ? v * 0.86 : v);
	const trendOption = $derived({
		tooltip: { trigger: 'axis' },
		legend: { top: 0, right: 0, data: ['Revenue', 'Gross profit', 'Net profit'] },
		grid: { top: 34, left: 4, right: 4, bottom: 0, containLabel: true },
		xAxis: { type: 'category', data: byMonth.map((d) => d.m), boundaryGap: false },
		yAxis: { type: 'value', axisLabel: { formatter: (v: number) => money(v / (ws.currency === 'INR' ? 88 : ws.currency === 'EUR' ? 0.86 : 1), ws.currency, { compact: true }) } },
		series: [
			{ name: 'Revenue', type: 'line', smooth: 0.35, symbol: 'none', data: byMonth.map((d) => Math.round(fx(d.revenue))), lineStyle: { width: 2 }, areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(20,83,60,0.16)' }, { offset: 1, color: 'rgba(20,83,60,0)' }] } } },
			{ name: 'Gross profit', type: 'line', smooth: 0.35, symbol: 'none', data: byMonth.map((d) => Math.round(fx(d.gross))), lineStyle: { width: 1.5, color: '#1E4D6B' }, itemStyle: { color: '#1E4D6B' } },
			{ name: 'Net profit', type: 'bar', barWidth: 10, data: byMonth.map((d) => Math.round(fx(d.net))), itemStyle: { color: '#C5CFCA', borderRadius: [2, 2, 0, 0] } }
		]
	});

	const buyers = $derived(ws.notifications.filter((n) => n.category === 'buyer_agency').slice(0, 3));
	const deadlines = $derived(ws.notifications.filter((n) => n.category === 'deadline'));
	const hour = new Date().getHours();
	const greet = hour < 12 ? 'Good morning' : hour < 17 ? 'Good afternoon' : 'Good evening';
</script>

<svelte:head><title>Dashboard — ZENRES</title></svelte:head>

<PageHeader eyebrow="{greet}, {USER.name.split(' ')[0]} · {new Date().toLocaleDateString('en-GB', { weekday: 'long', day: 'numeric', month: 'long' })}" title="Chakan Extrusion Plant," serif="live.">
	{#snippet actions()}
		<button class="btn btn-sm" onclick={() => ws.simulateInbound()}><Icon name="mail" size={14} />Simulate buyer email</button>
		<a class="btn btn-sm btn-dark" href="/carbon/cbam"><Icon name="cbam" size={14} />CBAM workbench</a>
	{/snippet}
</PageHeader>

<FactoryTwin initialZone={page.url.searchParams.get('zone')} />

<div class="kpis">
	<Kpi i={0} label="Monthly revenue · MTD" value={money(cur.revenue, ws.currency, { compact: true })} delta="{revDelta >= 0 ? '▲' : '▼'} {Math.abs(revDelta).toFixed(1)}% vs Oct" tone={revDelta >= 0 ? 'up' : 'down'} href="/finance">
		<div class="split"><span style="width:{exportShare * 100}%" class="exp"></span><span class="dom"></span></div>
		<div class="split-l mono"><span>Export {Math.round(exportShare * 100)}%</span><span>Domestic {100 - Math.round(exportShare * 100)}%</span></div>
	</Kpi>
	<Kpi i={1} label="Net profit · MTD" value={money(cur.net, ws.currency, { compact: true })} delta="Gross margin {grossMargin.toFixed(1)}%" tone="up" href="/finance">
		<div class="split">{#each costMix as [l, v, c] (l)}<span style="width:{v}%;background:{c}" title="{l} {v}%"></span>{/each}</div>
		<div class="split-l mono"><span>Mat · Energy · Labour · Freight</span></div>
	</Kpi>
	<Kpi i={2} label="Inventory valuation" value={money(invTotal, ws.currency, { compact: true })} delta="{lowStock} low-stock SKUs" tone={lowStock ? 'down' : 'flat'} href="/inventory">
		<div class="split">
			<span style="width:{(invSeg.raw / invTotal) * 100}%;background:var(--ink-secondary)"></span>
			<span style="width:{(invSeg.wip / invTotal) * 100}%;background:var(--accent-cobalt)"></span>
			<span style="width:{(invSeg.fg / invTotal) * 100}%;background:var(--brand-emerald)"></span>
		</div>
		<div class="split-l mono"><span>RM · WIP · FG</span></div>
	</Kpi>
	<Kpi i={3} label="Workforce on floor" value="{ws.totalWorkers} / {ws.totalTarget}" delta="{activeBatches.length} active batches · shift {shiftPct}%" tone="flat" href="/plants/plt_chakan/zones">
		<div class="bar"><span style="width:{shiftPct}%"></span></div>
	</Kpi>
	<Kpi i={4} label="CBAM exposure · est." value={money(cbam.savingsEur / 0.86, ws.currency, { compact: true })} delta="saved vs default · {exportTonnesYr} t/yr" tone="up" href="/carbon/cbam">
		<div class="split-l mono"><span>Default {money(cbam.defaultEur / 0.86, ws.currency, { compact: true })}</span><span>Actual {money(cbam.actualEur / 0.86, ws.currency, { compact: true })}</span></div>
		<span class="pill pill-terracotta est">Estimate</span>
	</Kpi>
	<Kpi i={5} label="Tally.ERP 9 & orders" value="14 min ago" delta="4 open export orders" href="/integrations/tally">
		<div class="split-l mono"><span>AR {money(181500, ws.currency, { compact: true })}</span><span>AP {money(127800, ws.currency, { compact: true })}</span></div>
	</Kpi>
</div>

<div class="row2">
	<section class="panel fade-in" style="--i:2">
		<div class="panel-head">
			<h2>Revenue & profit — last 12 months</h2>
			<span class="mono muted small">Tally vouchers · {ws.currency}</span>
		</div>
		<EChart option={trendOption} height={260} />
	</section>
	<section class="panel fade-in side" style="--i:3">
		<div class="panel-head">
			<h2>Buyers & agencies</h2>
			<button class="link-u small" onclick={() => (ws.bellOpen = true)}>Open bell</button>
		</div>
		<ul class="mini" role="list">
			{#each buyers as n (n.id)}
				<li class:unread={!n.read}>
					<a href="/inbox?thread={n.id}">
						<span class="org">{n.org_badge?.split(' · ')[0]}</span>
						<span class="subj">{n.subject}</span>
						<span class="mono t">{timeAgo(n.received_at)}{#if n.tag} · {n.tag}{/if}</span>
					</a>
				</li>
			{/each}
		</ul>
		<div class="panel-head dl"><h3>Deadlines</h3></div>
		<ul class="mini" role="list">
			{#each deadlines as n (n.id)}
				{@const d = dueIn(n.deadline!)}
				<li>
					<a href={n.link?.href ?? '/carbon/cbam'}>
						<span class="subj">{n.subject}</span>
						<span class="pill pill-{d.tone === 'muted' ? 'outline' : d.tone}">{d.label}</span>
					</a>
				</li>
			{/each}
		</ul>
	</section>
</div>

<section class="panel fade-in" style="--i:4">
	<div class="panel-head">
		<h2>Production & packaging — live batches</h2>
		<a class="link-u small" href="/trace/qr">QR Studio →</a>
	</div>
	<div class="table-wrap">
		<table class="table">
			<thead><tr><th>Batch</th><th>SKU</th><th>Zone</th><th>Stage</th><th>Progress</th><th class="num">tCO₂e (S1+S2)</th><th>Data quality</th></tr></thead>
			<tbody>
				{#each ws.batches.filter((b) => b.stage !== 'shipped') as b (b.id)}
					{@const p = ws.products.find((x) => x.id === b.product_id)}
					{@const prog = b.stage === 'packaging' || b.stage === 'ready_dispatch' ? b.packaged_qty / b.target_qty : b.produced_qty / b.target_qty}
					<tr>
						<td class="strong mono">{b.code}</td>
						<td>{p?.sku} <span class="muted">· CN {p?.cn}</span></td>
						<td><button class="link-u" onclick={() => (ws.selectedZone = b.zone_code)}>{b.zone_code}</button></td>
						<td><span class="pill" class:pill-emerald={b.stage === 'ready_dispatch'} class:pill-cobalt={b.stage === 'packaging'}>{b.stage.replace('_', ' ')}</span></td>
						<td style="min-width:160px"><div class="prog"><div class="bar"><span style="width:{prog * 100}%"></span></div><span class="mono small">{Math.round(prog * 100)}%</span></div></td>
						<td class="num">{num(b.direct_tco2e + b.indirect_tco2e, 2)}</td>
						<td><span class="pill {QUALITY_PILL[b.quality]}">{QUALITY_LABEL[b.quality]}</span></td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
	<p class="fine">CBAM figures are illustrative estimates at €{CERT_PRICE_EUR}/tCO₂e until verified. Logistics emissions are reported separately from embedded emissions.</p>
</section>

<style>
	.kpis {
		display: grid;
		grid-template-columns: repeat(6, minmax(0, 1fr));
		gap: 12px;
		margin: 16px 0;
	}
	.split {
		display: flex;
		height: 4px;
		border-radius: 2px;
		overflow: hidden;
		background: var(--surface-sunken);
		gap: 1px;
	}
	.split .exp {
		background: var(--brand-emerald);
	}
	.split .dom {
		flex: 1;
		background: var(--accent-cobalt-tint);
	}
	.split-l {
		display: flex;
		justify-content: space-between;
		gap: 6px;
		margin-top: 6px;
		font-size: 10px;
		color: var(--ink-muted);
		white-space: nowrap;
	}
	.est {
		margin-top: 6px;
	}
	.row2 {
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: 16px;
		margin-bottom: 16px;
	}
	.small {
		font-size: 11.5px;
	}
	.mini {
		display: flex;
		flex-direction: column;
	}
	.mini li a {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: 10px 0;
		border-bottom: var(--hairline);
		font-size: 12.5px;
	}
	.mini li:last-child a {
		border-bottom: 0;
	}
	.dl + .mini li a {
		flex-direction: row;
		justify-content: space-between;
		align-items: center;
		gap: 10px;
	}
	.mini .org {
		font-size: 11px;
		font-weight: 600;
		color: var(--accent-cobalt);
	}
	.mini .subj {
		color: var(--ink-primary);
		line-height: 1.35;
	}
	.mini .unread .subj {
		font-weight: 600;
	}
	.mini .t {
		font-size: 10.5px;
		color: var(--ink-faint);
	}
	.dl {
		margin-top: 16px;
		margin-bottom: 0;
	}
	.prog {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.prog .bar {
		flex: 1;
	}
	@media (max-width: 1400px) {
		.kpis {
			grid-template-columns: repeat(3, minmax(0, 1fr));
		}
	}
	@media (max-width: 1000px) {
		.row2 {
			grid-template-columns: 1fr;
		}
	}
	@media (max-width: 640px) {
		.kpis {
			grid-template-columns: 1fr 1fr;
		}
	}
</style>
