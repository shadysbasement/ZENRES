<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import HintWhisper from '$lib/components/HintWhisper.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { money } from '$lib/format';
	import { parseTallyCsv, autoMap, TALLY_FIELDS, upsertVouchers, SAMPLE_TALLY_CSV, type ColumnMap } from '$lib/tally';

	let step = $state(0);
	let fileName = $state('');
	let headers = $state<string[]>([]);
	let rows = $state<string[][]>([]);
	let map = $state<ColumnMap>({});
	let result = $state<{ inserted: number; duplicates: number; errors: number } | null>(null);
	let dragging = $state(false);
	let log = $state([
		{ at: '2026-11-26 09:16', source: 'tally_xml', detail: 'XML-over-HTTP connector · 42 vouchers · 0 duplicates', status: 'ok' },
		{ at: '2026-11-26 02:00', source: 'tally_odbc', detail: 'Nightly ODBC stock sync · 11 stock items', status: 'ok' },
		{ at: '2026-11-25 18:42', source: 'manual', detail: 'Correction: SAL/202611/0407 amount (₹21,48,000 → ₹21,84,000) by P. Iyer', status: 'warn' },
		{ at: '2026-11-25 09:15', source: 'tally_xml', detail: '38 vouchers · 2 mismatches flagged', status: 'warn' }
	]);

	async function load(file: File) {
		fileName = file.name;
		ingest(await file.text());
	}
	function ingest(text: string) {
		const parsed = parseTallyCsv(text);
		headers = parsed.headers;
		rows = parsed.rows;
		map = autoMap(headers);
		step = 1;
		result = null;
	}
	function sample() {
		fileName = 'Tally_DayBook_Nov26.csv';
		ingest(SAMPLE_TALLY_CSV);
	}
	function run() {
		const r = upsertVouchers(ws.vouchers, headers, rows, map, 'tally_csv');
		ws.vouchers = r.vouchers;
		result = { inserted: r.inserted, duplicates: r.duplicates, errors: r.errors };
		log.unshift({ at: new Date().toISOString().slice(0, 16).replace('T', ' '), source: 'tally_csv', detail: `${fileName} · ${r.inserted} inserted · ${r.duplicates} duplicates skipped`, status: r.errors ? 'warn' : 'ok' });
		step = 2;
		// multi-supplier CSVs whisper the Importer Hub on Growth
		const parties = new Set(rows.map((row) => row[headers.indexOf(Object.keys(map).find((k) => map[k] === 'counterparty') ?? '')])).size;
		if (parties > 4) ws.hint('tally', { kind: 'entity_pattern', payload: 'multi_supplier_csv' });
	}
</script>

<svelte:head><title>Tally.ERP 9 Hub — ZENRES</title></svelte:head>

<PageHeader eyebrow="ZENRES Business · A5" title="Tally.ERP 9" serif="integration hub" lede="ZENRES connects to Tally — it never replaces it. Import day books, run a local middleware connector, and keep an audit trail of every sync and correction." />

<div class="g2">
	<section class="panel wizard">
		<div class="panel-head">
			<h2>CSV / Excel import wizard</h2>
			<div class="steps mono">{#each ['Upload', 'Map columns', 'Result'] as s, i (s)}<span class:on={step >= i}>{i + 1}. {s}</span>{/each}</div>
		</div>

		{#if step === 0}
			<label class="drop" class:dragging ondragover={(e) => { e.preventDefault(); dragging = true; }} ondragleave={() => (dragging = false)} ondrop={(e) => { e.preventDefault(); dragging = false; const f = e.dataTransfer?.files[0]; if (f) load(f); }}>
				<input type="file" accept=".csv,.txt" onchange={(e) => { const f = (e.currentTarget as HTMLInputElement).files?.[0]; if (f) load(f); }} />
				<Icon name="upload" size={22} />
				<strong>Drop a Tally Day Book export (.csv)</strong>
				<span class="muted">Gateway of Tally → Display → Day Book → Export → CSV</span>
			</label>
			<button class="btn btn-sm" onclick={sample}>Use sample Day Book</button>
		{:else if step === 1}
			<p class="muted small">{fileName} · {rows.length} rows · columns auto-mapped</p>
			<div class="maps">
				{#each headers as h (h)}
					<label class="maprow"><span class="mono">{h}</span><span class="arrow">→</span>
						<select class="select" bind:value={map[h]}>
							<option value="">Ignore</option>
							{#each TALLY_FIELDS as f (f.key)}<option value={f.key}>{f.label}</option>{/each}
						</select>
					</label>
				{/each}
			</div>
			<div class="preview table-wrap">
				<table class="table"><thead><tr>{#each headers as h (h)}<th>{h}</th>{/each}</tr></thead><tbody>{#each rows.slice(0, 4) as r, i (i)}<tr>{#each r as c, j (j)}<td class="small">{c}</td>{/each}</tr>{/each}</tbody></table>
			</div>
			<div class="row"><button class="btn btn-sm btn-ghost" onclick={() => (step = 0)}>Back</button><button class="btn btn-sm btn-primary" onclick={run}>Import with deduplication</button></div>
		{:else if result}
			<div class="res">
				<div><span class="num big text-emerald">{result.inserted}</span><span>new vouchers</span></div>
				<div><span class="num big">{result.duplicates}</span><span>duplicates skipped</span></div>
				<div><span class="num big" class:text-crimson={result.errors}>{result.errors}</span><span>rows with errors</span></div>
			</div>
			<p class="small muted">Idempotency key: <span class="mono">(org_id, source, external_ref)</span>. Import the same file again — it inserts nothing.</p>
			<div class="row"><button class="btn btn-sm" onclick={run}>Re-import same file</button><button class="btn btn-sm btn-ghost" onclick={() => (step = 0)}>Import another</button><a class="btn btn-sm btn-dark" href="/finance">View finance</a></div>
		{/if}
		<HintWhisper context="tally" />
	</section>

	<div class="side">
		<section class="panel">
			<div class="panel-head"><h2>Local middleware connector</h2><span class="pill pill-emerald"><span class="dot dot-live"></span>Online</span></div>
			<ol class="setup" role="list">
				<li><span class="mono">01</span>Install <span class="mono">zevra-tally-bridge</span> on the Tally machine (Windows service).</li>
				<li><span class="mono">02</span>Enable Tally ODBC (port 9000) & XML-over-HTTP — kept on the LAN only.</li>
				<li><span class="mono">03</span>The bridge dials out over TLS; Tally is never exposed to the internet.</li>
			</ol>
			<pre class="cfg mono">bridge.toml
tally_url    = "http://127.0.0.1:9000"
mode         = ["xml", "odbc"]
schedule     = "*/15 * * * *"
upstream     = "https://api.zenres.example/api/v1/tally/connector-sync"
token        = "zvt_••••••••••••3f9a"</pre>
			{#if ws.manifest.capabilities.has('freightlink_core')}
				<div class="fsync"><Icon name="truck" size={14} /><span>FreightLink freight invoices sync back to Tally as purchase vouchers.</span></div>
			{/if}
		</section>
		<section class="panel">
			<div class="panel-head"><h2>Sync audit trail</h2></div>
			<ul class="log" role="list">
				{#each log as l, i (i)}
					<li><span class="dot" class:text-emerald={l.status === 'ok'} class:text-terracotta={l.status !== 'ok'}></span><div><span class="mono small">{l.at} · {l.source}</span><p>{l.detail}</p></div></li>
				{/each}
			</ul>
			<p class="fine">Totals: {ws.vouchers.length} vouchers · {money(ws.vouchers.filter((v) => v.type === 'sales').reduce((s, v) => s + v.amount_usd, 0), ws.currency, { compact: true })} sales</p>
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
	.wizard {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.steps {
		display: flex;
		gap: 12px;
		font-size: 10.5px;
		color: var(--ink-faint);
	}
	.steps .on {
		color: var(--brand-emerald);
	}
	.drop {
		position: relative;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		padding: 44px 20px;
		border: 1.5px dashed var(--border-strong);
		border-radius: var(--radius-md);
		text-align: center;
		color: var(--ink-secondary);
		cursor: pointer;
		transition: all 250ms;
	}
	.drop.dragging,
	.drop:hover {
		border-color: var(--brand-emerald);
		background: var(--brand-emerald-tint);
	}
	.drop input {
		position: absolute;
		inset: 0;
		opacity: 0;
		cursor: pointer;
	}
	.small {
		font-size: 11.5px;
	}
	.maps {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px;
	}
	.maprow {
		display: grid;
		grid-template-columns: 1fr 16px 1fr;
		align-items: center;
		gap: 8px;
		font-size: 12px;
	}
	.maprow .select {
		height: 34px;
	}
	.row {
		display: flex;
		gap: 8px;
		justify-content: flex-end;
	}
	.res {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 10px;
	}
	.res > div {
		padding: 16px;
		border: var(--hairline);
		border-radius: var(--radius-sm);
		display: flex;
		flex-direction: column;
		gap: 4px;
		font-size: 12px;
	}
	.big {
		font-size: 2rem;
		letter-spacing: -0.04em;
		color: var(--ink-primary);
	}
	.setup {
		display: flex;
		flex-direction: column;
		gap: 8px;
		font-size: 12.5px;
	}
	.setup li {
		display: grid;
		grid-template-columns: 26px 1fr;
	}
	.setup .mono:first-child {
		color: var(--brand-emerald);
		font-size: 11px;
	}
	.cfg {
		margin-top: 12px;
		padding: 12px;
		font-size: 11px;
		background: var(--obsidian);
		color: #a9d8c2;
		border-radius: var(--radius-sm);
		overflow-x: auto;
	}
	.fsync {
		display: flex;
		gap: 8px;
		align-items: center;
		margin-top: 10px;
		font-size: 12px;
		color: var(--accent-cobalt);
	}
	.log li {
		display: grid;
		grid-template-columns: 14px 1fr;
		gap: 8px;
		padding: 9px 0;
		border-bottom: var(--hairline);
		font-size: 12.5px;
	}
	.log .dot {
		margin-top: 6px;
	}
	@media (max-width: 1100px) {
		.g2 {
			grid-template-columns: 1fr;
		}
	}
</style>
