<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import { ws } from '$lib/stores/workspace.svelte';

	let f = $state({ name: '', gstin: '', contact: '', rc: '', type: '32 ft MXL container', payload: 21, insurance: '', permit: 'National permit' });
	const checklist = $derived([
		['Business identity & GSTIN', !!f.name && f.gstin.length === 15],
		['Contact person', !!f.contact],
		['Vehicle registration (RC)', /^[A-Z]{2}-?\d{2}-?[A-Z]{1,3}-?\d{3,4}$/i.test(f.rc)],
		['Insurance valid', !!f.insurance && new Date(f.insurance) > new Date()],
		['National / state permit', !!f.permit]
	] as const);
	function submit(e: SubmitEvent) {
		e.preventDefault();
		ws.freight.transporters.push({ name: f.name, city: '—', vehicles: 1, verified: false, reliability: 5, on_time: 0, docs: { rc: true, insurance: true, permit: true, kyc: false } });
		ws.toast(`${f.name} submitted for free verification`);
	}
</script>

<svelte:head><title>Transporters — ZENRES</title></svelte:head>

<PageHeader eyebrow="ZENRES Trace & FreightLink · C4" title="Transporter onboarding &" serif="verified directory" lede="Free carrier onboarding. Verified badges and organic ranking never depend on paid promotion." />

<div class="g2">
	<form class="panel" onsubmit={submit}>
		<div class="panel-head"><h2>Register a carrier — free</h2></div>
		<div class="grid">
			<label class="field"><span class="label">Business name</span><input class="input" bind:value={f.name} /></label>
			<label class="field"><span class="label">GSTIN</span><input class="input mono" bind:value={f.gstin} maxlength="15" /></label>
			<label class="field"><span class="label">Contact person & phone</span><input class="input" bind:value={f.contact} /></label>
			<label class="field"><span class="label">Vehicle RC no.</span><input class="input mono" bind:value={f.rc} placeholder="MH-04-KT-7781" /></label>
			<label class="field"><span class="label">Vehicle type</span><select class="select" bind:value={f.type}><option>32 ft MXL container</option><option>32 ft SXL container</option><option>19 ft closed body</option><option>20 ft open body</option><option>Trailer 40 ft</option></select></label>
			<label class="field"><span class="label">Rated payload (t)</span><input class="input num" type="number" bind:value={f.payload} /></label>
			<label class="field"><span class="label">Insurance valid until</span><input class="input" type="date" bind:value={f.insurance} /></label>
			<label class="field"><span class="label">Permit</span><select class="select" bind:value={f.permit}><option>National permit</option><option>State permit</option></select></label>
		</div>
		<div class="check">
			<span class="eyebrow">Verification checklist</span>
			{#each checklist as [l, ok] (l)}<div class:ok><span class="m">{ok ? '✓' : '·'}</span>{l}</div>{/each}
		</div>
		<button class="btn btn-primary" disabled={!checklist.every(([, ok]) => ok)}>Submit for verification</button>
	</form>
	<section class="panel">
		<div class="panel-head"><h2>Verified directory</h2></div>
		{#each [...ws.freight.transporters].sort((a, b) => b.reliability - a.reliability) as t (t.name)}
			<div class="tr">
				<div><strong>{t.name}</strong> {#if t.verified}<span class="pill pill-emerald">Verified</span>{:else}<span class="pill pill-terracotta">Pending</span>{/if}<div class="muted small">{t.city} · {t.vehicles} vehicles</div></div>
				<div class="stats mono small"><span>★ {t.reliability}</span><span>{t.on_time}% on-time</span></div>
				<div class="docs">{#each Object.entries(t.docs) as [k, v] (k)}<span class="d" class:ok={v}>{k.toUpperCase()}</span>{/each}</div>
			</div>
		{/each}
		<p class="fine">Ranking uses reliability and on-time history only. There is no paid placement.</p>
	</section>
</div>

<style>
	.g2 {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 16px;
		align-items: start;
	}
	.grid {
		grid-template-columns: 1fr 1fr;
		gap: 12px;
	}
	.check {
		display: flex;
		flex-direction: column;
		gap: 6px;
		margin: 16px 0;
		font-size: 12.5px;
		color: var(--ink-muted);
	}
	.check .ok {
		color: var(--brand-emerald);
	}
	.m {
		display: inline-block;
		width: 18px;
	}
	.tr {
		display: grid;
		grid-template-columns: 1fr auto;
		gap: 6px 12px;
		padding: 12px 0;
		border-bottom: var(--hairline);
		font-size: 13px;
	}
	.stats {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
	}
	.docs {
		grid-column: 1 / -1;
		display: flex;
		gap: 4px;
	}
	.d {
		font-family: var(--font-mono);
		font-size: 9.5px;
		padding: 2px 6px;
		border-radius: 2px;
		background: var(--status-terracotta-tint);
		color: var(--status-terracotta);
	}
	.d.ok {
		background: var(--brand-emerald-tint);
		color: var(--brand-emerald);
	}
	.small {
		font-size: 11.5px;
	}
	@media (max-width: 1000px) {
		.g2,
		.grid {
			grid-template-columns: 1fr;
		}
	}
</style>
