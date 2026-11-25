<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Kpi from '$lib/components/ui/Kpi.svelte';
	import EChart from '$lib/components/charts/EChart.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { MONTHS, ORDERS } from '$lib/demo/seed';
	import { money } from '$lib/format';

	let tab = $state<'sales' | 'purchases' | 'invoices'>('sales');
	const mk = (i: number) => (i === 0 ? '2025-12' : `2026-${String(i).padStart(2, '0')}`);
	const months = $derived(
		MONTHS.map((m, i) => {
			const v = ws.vouchers.filter((x) => x.date.startsWith(mk(i)));
			const sum = (t: string, f: 'amount_usd' | 'cost_usd' = 'amount_usd') => v.filter((x) => x.type === t).reduce((s, x) => s + x[f], 0);
			const revenue = sum('sales');
			const cogs = sum('sales', 'cost_usd');
			const expenses = (sum('expense') + sum('utility') + sum('freight')) * 0.35;
			return { m, revenue, expenses: cogs + expenses, gross: revenue - cogs, net: revenue - cogs - expenses };
		})
	);
	const ytd = $derived(months.reduce((a, m) => ({ revenue: a.revenue + m.revenue, net: a.net + m.net, gross: a.gross + m.gross }), { revenue: 0, net: 0, gross: 0 }));
	const k = $derived(ws.currency === 'INR' ? 88 : ws.currency === 'EUR' ? 0.86 : 1);
	const pnl = $derived({
		tooltip: { trigger: 'axis' },
		legend: { top: 0, right: 0 },
		grid: { top: 34, left: 4, right: 4, bottom: 0, containLabel: true },
		xAxis: { type: 'category', data: months.map((m) => m.m) },
		yAxis: { type: 'value', axisLabel: { formatter: (v: number) => money(v / k, ws.currency, { compact: true }) } },
		series: [
			{ name: 'Revenue', type: 'bar', barGap: '10%', barWidth: 9, data: months.map((m) => Math.round(m.revenue * k)), itemStyle: { color: '#14533C', borderRadius: [2, 2, 0, 0] } },
			{ name: 'Expenses', type: 'bar', barWidth: 9, data: months.map((m) => Math.round(m.expenses * k)), itemStyle: { color: '#C5CFCA', borderRadius: [2, 2, 0, 0] } },
			{ name: 'Net profit', type: 'line', smooth: 0.3, symbolSize: 5, data: months.map((m) => Math.round(m.net * k)), lineStyle: { color: '#1E4D6B', width: 2 }, itemStyle: { color: '#1E4D6B' } }
		]
	});
	const aging = (rows: { bucket: string; amount: number }[], color: string) => ({
		grid: { top: 8, left: 4, right: 30, bottom: 0, containLabel: true },
		xAxis: { type: 'value', show: false },
		yAxis: { type: 'category', data: rows.map((r) => r.bucket), inverse: true },
		tooltip: { trigger: 'axis' },
		series: [{ type: 'bar', barWidth: 10, data: rows.map((r) => Math.round(r.amount * k)), itemStyle: { color, borderRadius: 2 }, label: { show: true, position: 'right', fontFamily: 'IBM Plex Mono', fontSize: 10, color: '#47544F', formatter: (p: { value: number }) => money(p.value / k, ws.currency, { compact: true }) } }]
	});
	const invoices = $derived(ws.vouchers.filter((v) => v.type === 'sales').slice(-12).reverse());
</script>

<svelte:head><title>Finance — Zevra</title></svelte:head>

<PageHeader eyebrow="Zevra Business · A4" title="Financials, orders &" serif="production costing" lede="Revenue, expenses and profit sourced from Tally vouchers, with order books, invoice register and receivables/payables aging." />

<div class="kpis">
	<Kpi i={0} label="Revenue · 12 mo" value={money(ytd.revenue, ws.currency, { compact: true })} />
	<Kpi i={1} label="Gross profit · 12 mo" value={money(ytd.gross, ws.currency, { compact: true })} delta="{((ytd.gross / ytd.revenue) * 100).toFixed(1)}% margin" tone="up" />
	<Kpi i={2} label="Net profit · 12 mo" value={money(ytd.net, ws.currency, { compact: true })} delta="{((ytd.net / ytd.revenue) * 100).toFixed(1)}% net margin" tone="up" />
	<Kpi i={3} label="Receivables" value={money(ORDERS.receivables.reduce((s, r) => s + r.amount, 0), ws.currency, { compact: true })} delta="{money(ORDERS.receivables[3].amount, ws.currency, { compact: true })} over 90 days" tone="down" />
</div>

<div class="g2">
	<section class="panel"><div class="panel-head"><h2>Profit & loss</h2><span class="muted mono small">Tally-synced · {ws.currency}</span></div><EChart option={pnl} height={280} /></section>
	<div class="side">
		<section class="panel"><div class="panel-head"><h2>Receivables aging</h2></div><EChart option={aging(ORDERS.receivables, '#14533C')} height={120} /></section>
		<section class="panel"><div class="panel-head"><h2>Payables aging</h2></div><EChart option={aging(ORDERS.payables, '#1E4D6B')} height={120} /></section>
	</div>
</div>

<section class="panel" style="margin-top:16px">
	<div class="tabs-line" role="tablist">
		<button role="tab" aria-selected={tab === 'sales'} onclick={() => (tab = 'sales')}>Customer sales orders</button>
		<button role="tab" aria-selected={tab === 'purchases'} onclick={() => (tab = 'purchases')}>Supplier purchase orders</button>
		<button role="tab" aria-selected={tab === 'invoices'} onclick={() => (tab = 'invoices')}>Invoice register</button>
	</div>
	<div class="table-wrap">
		{#if tab === 'sales'}
			<table class="table"><thead><tr><th>Order</th><th>Customer</th><th>SKU</th><th class="num">Qty</th><th class="num">Value</th><th>Incoterm</th><th>Due</th><th>Status</th></tr></thead>
				<tbody>{#each ORDERS.sales as o (o.ref)}<tr><td class="mono strong">{o.ref}</td><td>{o.customer} {#if o.export}<span class="pill pill-cobalt">Export</span>{/if}</td><td class="mono">{o.product}</td><td class="num">{o.qty} t</td><td class="num">{money(o.value, ws.currency)}</td><td>{o.incoterm}</td><td>{o.due}</td><td><span class="pill">{o.status}</span></td></tr>{/each}</tbody></table>
		{:else if tab === 'purchases'}
			<table class="table"><thead><tr><th>PO</th><th>Supplier</th><th>Item</th><th class="num">Qty</th><th class="num">Value</th><th>Due</th><th>Status</th></tr></thead>
				<tbody>{#each ORDERS.purchases as o (o.ref)}<tr><td class="mono strong">{o.ref}</td><td>{o.supplier}</td><td>{o.item}</td><td class="num">{o.qty} t</td><td class="num">{money(o.value, ws.currency)}</td><td>{o.due}</td><td><span class="pill">{o.status}</span></td></tr>{/each}</tbody></table>
		{:else}
			<table class="table"><thead><tr><th>Voucher</th><th>Counterparty</th><th>Date</th><th>Source</th><th class="num">Amount</th><th class="num">Margin</th></tr></thead>
				<tbody>{#each invoices as v (v.id)}<tr><td class="mono strong">{v.ref}</td><td>{v.counterparty}</td><td>{v.date}</td><td><span class="pill pill-outline">{v.source.replace('_', ' ')}</span></td><td class="num">{money(v.amount_usd, ws.currency)}</td><td class="num">{(((v.amount_usd - v.cost_usd) / v.amount_usd) * 100).toFixed(1)}%</td></tr>{/each}</tbody></table>
		{/if}
	</div>
</section>

<style>
	.kpis {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 12px;
		margin-bottom: 16px;
	}
	.g2 {
		display: grid;
		grid-template-columns: 1.7fr 1fr;
		gap: 16px;
	}
	.side {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.small {
		font-size: 11px;
	}
	.strong {
		color: var(--ink-primary);
	}
	@media (max-width: 1100px) {
		.g2 {
			grid-template-columns: 1fr;
		}
		.kpis {
			grid-template-columns: 1fr 1fr;
		}
	}
</style>
