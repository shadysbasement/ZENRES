<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HintWhisper from '$lib/components/HintWhisper.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { activityEmissions } from '$lib/carbon';
	import { num, QUALITY_LABEL, QUALITY_PILL } from '$lib/format';

	const ocr = $derived(ws.manifest.capabilities.has('ocr_ingestion'));
	const bot = $derived(ws.manifest.capabilities.has('whatsapp_floorbot'));
	const EF: Record<string, { unit: string; ef: number; rule: string }> = {
		electricity: { unit: 'kWh', ef: 0.000727, rule: 'CEA-v20 / IR 2025/2621' },
		fuel: { unit: 'L diesel', ef: 0.00268, rule: 'IPCC 2006 / IR 2025/2621' },
		precursor: { unit: 't', ef: 3.9, rule: 'Supplier EPD / IR 2025/2621' },
		process: { unit: 't', ef: 1.62, rule: 'IR 2025/2621 Annex III' }
	};
	let f = $state({ type: 'electricity', period: '2026-11', qty: '', invoice: '', meter: '' });
	let file = $state<File | null>(null);
	let touched = $state(new Set<string>());
	let dragging = $state(false);
	let ocrJob = $state<null | { name: string; fields: { k: string; v: string; c: number }[] }>(null);

	function touch(k: string) {
		touched.add(k);
		touched = new Set(touched);
		if (file && touched.size >= 3) ws.hint('inputs', { kind: 'manual_effort', payload: 'bill_manual_fields', count: touched.size });
	}
	function onFile(fl: File) {
		const revised = ws.activity.some((a) => a.file.replace(/\.\w+$/, '').toLowerCase() === fl.name.replace(/(_v\d+|_rev\d*|\s\(\d\))?\.\w+$/i, '').toLowerCase());
		file = fl;
		if (revised) ws.hint('inputs-revised', { kind: 'manual_effort', payload: 'revised_upload' });
		if (ocr) {
			ocrJob = { name: fl.name, fields: [] };
			const fields = [
				{ k: 'Supplier', v: 'MSEDCL — HT Consumer 170019284551', c: 0.98 },
				{ k: 'Billing period', v: '01-Nov-2026 → 30-Nov-2026', c: 0.95 },
				{ k: 'Units consumed (kWh)', v: '421,560', c: 0.93 },
				{ k: 'Max demand (kVA)', v: '1,180', c: 0.71 },
				{ k: 'Amount payable', v: '₹38,41,200', c: 0.97 }
			];
			fields.forEach((fd, i) => setTimeout(() => ocrJob && (ocrJob.fields = [...ocrJob.fields, fd]), 350 * (i + 1)));
		}
	}
	function save() {
		const e = EF[f.type];
		const qty = Number(String(f.qty).replace(/,/g, ''));
		ws.activity.unshift({ id: `act${Date.now()}`, period: f.period, type: f.type, mode: ocrJob ? 'ocr' : 'manual', qty, unit: e.unit, ef: e.ef, rule: e.rule, file: file?.name ?? '—', quality: file ? 'measured' : 'estimated' });
		ws.toast(`${num(activityEmissions(qty, e.ef), 2)} tCO₂e recorded for ${f.period}`);
		f.qty = '';
		file = null;
		ocrJob = null;
		touched = new Set();
	}
	function acceptOcr() {
		f.qty = '421560';
		f.type = 'electricity';
		save();
	}

	const chat = [
		{ from: 'bot', text: 'नमस्ते राकेश जी! Line A shift-1 production? (tonnes)' },
		{ from: 'sup', text: 'Line A 4.2 ton, batch A-104. 11 log present.' },
		{ from: 'bot', text: 'Noted ✓ A-104 +4.2 t · PROD-A headcount 11. Diesel DG reading?' },
		{ from: 'sup', text: 'DG 2 ghanta chala, 96 litre' },
		{ from: 'bot', text: '96 L diesel → 0.257 tCO₂e (Scope 1, measured) linked to shift-1.' }
	];
</script>

<svelte:head><title>Activity Inputs — ZENRES</title></svelte:head>

<PageHeader eyebrow="ZENRES Carbon" title="Activity data" serif="capture" lede="Electricity, fuel, process and raw-material inputs — from bills, meters and the factory floor — each with a data-quality label." />

<div class="g2">
	<section class="panel">
		<div class="panel-head"><h2>Upload bill or evidence</h2>{#if ocr}<span class="pill pill-emerald">OCR extraction</span>{/if}</div>
		<label class="drop" class:dragging ondragover={(e) => { e.preventDefault(); dragging = true; }} ondragleave={() => (dragging = false)} ondrop={(e) => { e.preventDefault(); dragging = false; const fl = e.dataTransfer?.files[0]; if (fl) onFile(fl); }}>
			<input type="file" accept=".pdf,.png,.jpg,.jpeg,.csv,.xlsx" onchange={(e) => { const fl = (e.currentTarget as HTMLInputElement).files?.[0]; if (fl) onFile(fl); }} />
			<Icon name="upload" size={20} />
			{#if file}<strong>{file.name}</strong><span class="muted small">{(file.size / 1024).toFixed(0)} KB · stored in AES-256 evidence vault</span>{:else}<strong>Drop a utility bill, fuel log or invoice</strong><span class="muted small">PDF, photo or Tally CSV</span>{/if}
		</label>
		<HintWhisper context="inputs-revised" />

		{#if ocrJob}
			<div class="ocr card">
				<div class="row"><span class="eyebrow">OCR queue · {ocrJob.name}</span><span class="mono small">{ocrJob.fields.length}/5 fields</span></div>
				{#each ocrJob.fields as fd (fd.k)}
					<div class="ofield" class:low={fd.c < 0.8}><span>{fd.k}</span><span class="mono">{fd.v}</span><span class="conf mono">{Math.round(fd.c * 100)}%</span></div>
				{/each}
				{#if ocrJob.fields.length === 5}
					<p class="small muted">Low-confidence fields are highlighted for human review before posting.</p>
					<button class="btn btn-sm btn-primary" onclick={acceptOcr}>Review & post 421,560 kWh</button>
				{/if}
			</div>
		{/if}

		<div class="form">
			<div class="two">
				<label class="field"><span class="label">Activity type</span><select class="select" bind:value={f.type} onchange={() => touch('type')}><option value="electricity">Electricity</option><option value="fuel">Fuel (diesel)</option><option value="precursor">Precursor input</option><option value="process">Process emissions</option></select></label>
				<label class="field"><span class="label">Period</span><input class="input mono" type="month" bind:value={f.period} oninput={() => touch('period')} /></label>
			</div>
			<div class="two">
				<label class="field"><span class="label">Quantity ({EF[f.type].unit})</span><input class="input num" bind:value={f.qty} oninput={() => touch('qty')} placeholder="412800" /></label>
				<label class="field"><span class="label">Meter / invoice no.</span><input class="input mono" bind:value={f.invoice} oninput={() => touch('invoice')} placeholder="MSEDCL-HT-4471" /></label>
			</div>
			<div class="ef mono small">EF {EF[f.type].ef} tCO₂e/{EF[f.type].unit} · {EF[f.type].rule}{#if f.qty} · = <strong>{num(activityEmissions(Number(String(f.qty).replace(/,/g, '')) || 0, EF[f.type].ef), 3)} tCO₂e</strong>{/if}</div>
			<HintWhisper context="inputs" />
			<button class="btn btn-sm btn-dark" disabled={!f.qty} onclick={save}>Record activity</button>
		</div>
	</section>

	<section class="panel floor">
		<div class="panel-head"><h2>Factory-floor capture</h2>{#if bot}<span class="pill pill-emerald"><span class="dot dot-live"></span>Floor-bot live</span>{/if}</div>
		{#if bot}
			<div class="chat">
				{#each chat as m, i (i)}<div class="msg {m.from}" style="--i:{i}">{m.text}</div>{/each}
			</div>
			<p class="fine">Vernacular floor-bot (Hindi / Marathi / English) maps supervisor messages to shifts, batches and meters via the WhatsApp Business webhook.</p>
		{:else}
			<p class="small">Supervisors can share the daily floor sheet as a photo, and you can key the values here.</p>
			<button class="btn btn-sm" onclick={() => ws.hint('inputs-wa', { kind: 'manual_effort', payload: 'whatsapp_click' })}><Icon name="whatsapp" size={14} />Share floor sheet via WhatsApp</button>
			<HintWhisper context="inputs-wa" />
		{/if}
		<div class="tally card">
			<Icon name="tally" size={16} />
			<div><strong>Tally CSV</strong><p class="small muted">Utility & fuel purchase vouchers sync automatically from the Tally hub.</p></div>
			<a class="btn btn-sm btn-ghost" href="/integrations/tally">Open</a>
		</div>
	</section>
</div>

<section class="panel" style="margin-top:16px">
	<div class="panel-head"><h2>Activity records</h2><span class="mono muted small">{ws.activity.length} records</span></div>
	<div class="table-wrap">
		<table class="table">
			<thead><tr><th>Period</th><th>Type</th><th>Ingestion</th><th class="num">Quantity</th><th class="num">EF</th><th class="num">tCO₂e</th><th>Rule version</th><th>Evidence</th><th>Quality</th></tr></thead>
			<tbody>
				{#each ws.activity as a (a.id)}
					<tr><td class="mono">{a.period}</td><td class="cap">{a.type}</td><td><span class="pill pill-outline">{a.mode.replace('_', ' ')}</span></td><td class="num">{num(a.qty)} {a.unit}</td><td class="num">{a.ef}</td><td class="num strong">{num(activityEmissions(a.qty, a.ef), 2)}</td><td class="small">{a.rule}</td><td class="mono small">{a.file}</td><td><span class="pill {QUALITY_PILL[a.quality]}">{QUALITY_LABEL[a.quality]}</span></td></tr>
				{/each}
			</tbody>
		</table>
	</div>
</section>

<style>
	.g2 {
		display: grid;
		grid-template-columns: 1.3fr 1fr;
		gap: 16px;
		align-items: start;
	}
	.drop {
		position: relative;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 6px;
		padding: 30px;
		border: 1.5px dashed var(--border-strong);
		border-radius: var(--radius-md);
		text-align: center;
		cursor: pointer;
		transition: all 250ms;
	}
	.drop:hover,
	.drop.dragging {
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
	.form {
		display: flex;
		flex-direction: column;
		gap: 12px;
		margin-top: 16px;
	}
	.two {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 10px;
	}
	.ef {
		color: var(--accent-cobalt);
	}
	.ocr {
		margin-top: 12px;
		padding: 14px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.row {
		display: flex;
		justify-content: space-between;
	}
	.ofield {
		display: grid;
		grid-template-columns: 150px 1fr 44px;
		gap: 8px;
		padding: 7px 8px;
		border-radius: 3px;
		font-size: 12px;
		background: var(--surface-canvas);
		animation: zv-fade-in 400ms var(--ease-expo);
	}
	.ofield.low {
		background: var(--status-terracotta-tint);
	}
	.conf {
		text-align: right;
		color: var(--brand-emerald);
	}
	.low .conf {
		color: var(--status-terracotta);
	}
	.floor {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.chat {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 14px;
		border-radius: var(--radius-md);
		background: var(--surface-subtle);
	}
	.msg {
		max-width: 82%;
		padding: 8px 12px;
		border-radius: 10px;
		font-size: 12.5px;
		line-height: 1.45;
		animation: zv-fade-in 500ms var(--ease-expo) both;
		animation-delay: calc(var(--i) * 250ms);
	}
	.msg.bot {
		background: var(--surface-card);
		border: var(--hairline);
		color: var(--ink-primary);
	}
	.msg.sup {
		align-self: flex-end;
		background: var(--brand-emerald);
		color: #fff;
	}
	.tally {
		display: grid;
		grid-template-columns: auto 1fr auto;
		gap: 12px;
		align-items: center;
		padding: 12px;
		font-size: 13px;
	}
	.cap {
		text-transform: capitalize;
	}
	.strong {
		color: var(--ink-primary);
		font-weight: 500;
	}
	@media (max-width: 1100px) {
		.g2 {
			grid-template-columns: 1fr;
		}
	}
</style>
