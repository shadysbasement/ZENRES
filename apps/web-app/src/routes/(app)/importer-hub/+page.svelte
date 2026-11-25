<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HintWhisper from '$lib/components/HintWhisper.svelte';
	import Kpi from '$lib/components/ui/Kpi.svelte';
	import EChart from '$lib/components/charts/EChart.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { CERT_PRICE_EUR } from '$lib/carbon';
	import { timeAgo } from '$lib/format';

	const included = $derived(ws.manifest.limits.included_suppliers);
	const active = $derived(ws.suppliers.length);
	let emails = $state('');
	let confirmExtra = $state(false);
	let cadence = $state('weekly');
	const STATUS = ['invited', 'onboarding', 'pending_evidence', 'submitted', 'verified'] as const;
	const statusPill: Record<string, string> = { invited: 'pill-outline', onboarding: '', pending_evidence: 'pill-terracotta', submitted: 'pill-cobalt', verified: 'pill-emerald' };

	function invite() {
		const list = emails.split(/[\s,;]+/).filter((e) => /@/.test(e));
		if (!list.length) return;
		if (active + list.length > included && !confirmExtra) {
			ws.hint('hub', { kind: 'limit_hit', payload: 'invite_supplier', count: active });
			confirmExtra = true;
			return;
		}
		list.forEach((e, i) => ws.suppliers.push({ id: `sup${Date.now()}${i}`, name: e.split('@')[1].split('.')[0].replace(/-/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase()), city: '—', cn: '—', status: 'invited', intensity: null, tonnes: 0, last_reminder: null }));
		ws.toast(`${list.length} invitation link${list.length > 1 ? 's' : ''} sent${active + list.length > included ? ` · ${active + list.length - included} extra seat(s) at $25/mo` : ''}`);
		emails = '';
		confirmExtra = false;
		ws.clearHint('hub');
	}
	function remind() {
		const n = ws.suppliers.filter((s) => s.status === 'pending_evidence' || s.status === 'invited').length;
		ws.suppliers.forEach((s) => {
			if (s.status === 'pending_evidence' || s.status === 'invited') s.last_reminder = new Date().toISOString();
		});
		ws.toast(`Missing-evidence reminders sent to ${n} suppliers`);
	}
	const scatter = $derived({
		tooltip: { trigger: 'item', formatter: (p: { data: [number, number, string, number] }) => `${p.data[2]}<br/>${p.data[0]} tCO₂e/t · ${p.data[3]} t<br/>€${Math.round(p.data[1]).toLocaleString()} exposure` },
		grid: { top: 16, left: 4, right: 16, bottom: 4, containLabel: true },
		xAxis: { type: 'value', name: 'tCO₂e / t', nameLocation: 'middle', nameGap: 26, nameTextStyle: { fontSize: 10 } },
		yAxis: { type: 'value', name: '€ exposure', axisLabel: { formatter: (v: number) => `€${(v / 1000).toFixed(0)}k` } },
		series: [
			{
				type: 'scatter',
				symbolSize: (d: [number, number, string, number]) => 8 + Math.sqrt(d[3]) * 2.4,
				data: ws.suppliers.filter((s) => s.intensity).map((s) => [s.intensity, s.intensity! * s.tonnes * CERT_PRICE_EUR, s.name, s.tonnes]),
				itemStyle: { color: 'rgba(20,83,60,0.55)', borderColor: '#14533C' }
			}
		]
	});
	const ranked = $derived([...ws.suppliers].filter((s) => s.intensity).sort((a, b) => a.intensity! - b.intensity!));
</script>

<svelte:head><title>Importer Hub — ZENRES</title></svelte:head>

<PageHeader eyebrow="ZENRES Connect · D2 · EU Importer Hub" title="Supplier" serif="network portal" lede="15 suppliers included, scalable to 500. Each supplier works in an isolated tenant — no supplier ever sees another’s pricing, BOM or records.">
	{#snippet actions()}
		<a class="btn btn-sm btn-primary" href="/carbon/exports"><Icon name="layers" size={14} />Roll into consolidated XML</a>
	{/snippet}
</PageHeader>

<div class="kpis">
	<Kpi i={0} label="Supplier seats" value="{active} / {included}" delta={active > included ? `${active - included} extra × $25/mo` : `${included - active} included seats free`} />
	<Kpi i={1} label="Verified" value={String(ws.suppliers.filter((s) => s.status === 'verified').length)} tone="up" />
	<Kpi i={2} label="Evidence outstanding" value={String(ws.suppliers.filter((s) => s.status === 'pending_evidence').length)} tone="down" />
	<Kpi i={3} label="Q4 imported tonnes" value="{ws.suppliers.reduce((s, x) => s + x.tonnes, 0).toFixed(0)} t" />
</div>

<div class="g2">
	<section class="panel">
		<div class="panel-head"><h2>Submission & evidence matrix</h2><button class="btn btn-sm" onclick={remind}><Icon name="clock" size={13} />Send reminders</button></div>
		<div class="table-wrap">
			<table class="table">
				<thead><tr><th>Supplier</th><th>CN</th>{#each STATUS as s (s)}<th class="c">{s.replace('_', ' ')}</th>{/each}<th>Last reminder</th></tr></thead>
				<tbody>
					{#each ws.suppliers as s (s.id)}
						{@const idx = STATUS.indexOf(s.status)}
						<tr>
							<td class="strong nowrap">{s.name}<div class="small muted">{s.city}</div></td>
							<td class="mono small">{s.cn}</td>
							{#each STATUS as st, i (st)}<td class="c"><span class="cell" class:done={i < idx} class:cur={i === idx} class:warn={i === idx && st === 'pending_evidence'}></span></td>{/each}
							<td class="small muted">{s.last_reminder ? timeAgo(s.last_reminder) : '—'}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</section>
	<div class="side">
		<section class="panel">
			<div class="panel-head"><h2>Bulk invitation links</h2></div>
			<textarea class="textarea" bind:value={emails} placeholder="compliance@supplier-one.in, exports@supplier-two.in …"></textarea>
			<HintWhisper context="hub" />
			<div class="row">
				<select class="select" bind:value={cadence}><option value="weekly">Weekly reminders</option><option value="biweekly">Every 2 weeks</option><option value="deadline">7 / 3 / 1 days before deadline</option></select>
				<button class="btn btn-sm btn-dark" onclick={invite}>{confirmExtra ? 'Confirm & invite ($25/mo each)' : 'Send invitations'}</button>
			</div>
		</section>
		<section class="panel">
			<div class="panel-head"><h2>Carbon intensity vs CBAM exposure</h2></div>
			<EChart option={scatter} height={220} />
		</section>
	</div>
</div>

<section class="panel" style="margin-top:16px">
	<div class="panel-head"><h2>Supplier benchmarking · ranked by intensity</h2><span class="muted small">€{CERT_PRICE_EUR}/tCO₂e indicative</span></div>
	<div class="bench">
		{#each ranked as s, i (s.id)}
			<div class="b"><span class="mono">#{i + 1}</span><span class="nm">{s.name}</span><span class="mono small">CN {s.cn}</span><div class="bar"><span style="width:{(s.intensity! / 8) * 100}%;{s.intensity! > 5 ? 'background:var(--status-terracotta)' : ''}"></span></div><span class="mono">{s.intensity}</span></div>
		{/each}
	</div>
	<p class="fine">Row-level security: each supplier record is scoped by org_id; suppliers cannot query or guess one another’s data.</p>
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
		grid-template-columns: 1.6fr 1fr;
		gap: 16px;
		align-items: start;
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
	.c {
		text-align: center;
		font-size: 10px;
	}
	.cell {
		display: inline-block;
		width: 14px;
		height: 14px;
		border-radius: 3px;
		background: var(--surface-subtle);
	}
	.cell.done {
		background: var(--brand-emerald);
	}
	.cell.cur {
		background: var(--accent-cobalt);
		box-shadow: 0 0 0 3px var(--accent-cobalt-tint);
	}
	.cell.warn {
		background: var(--status-terracotta);
		box-shadow: 0 0 0 3px var(--status-terracotta-tint);
	}
	.row {
		display: flex;
		gap: 8px;
		margin-top: 10px;
	}
	.bench {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 4px 32px;
	}
	.b {
		display: grid;
		grid-template-columns: 30px 1fr 110px 120px 40px;
		gap: 10px;
		align-items: center;
		padding: 7px 0;
		border-bottom: var(--hairline);
		font-size: 12.5px;
	}
	.nm {
		color: var(--ink-primary);
	}
	@media (max-width: 1200px) {
		.g2,
		.bench {
			grid-template-columns: 1fr;
		}
		.kpis {
			grid-template-columns: 1fr 1fr;
		}
	}
</style>
