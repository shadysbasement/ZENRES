<script lang="ts">
	import Logo from '$design/components/Logo.svelte';
	import { OPS } from '$lib/demo/seed';

	let { data } = $props();
	const tabs = ['Approvals', 'Trust & disputes', 'CBAM rule publisher', 'Margin guardrails', 'FreightLink KPIs', 'System & audit'] as const;
	let tab = $state<(typeof tabs)[number]>('Margin guardrails');
	let approvals = $state(OPS.approvals.map((a) => ({ ...a, decision: '' })));
	let rules = $state([
		{ v: 'IR (EU) 2025/2621', effective: '2026-01-01', status: 'live', note: 'Definitive regime default values + markup schedule' },
		{ v: 'IR (EU) 2023/1773', effective: '2023-10-01', status: 'archived', note: 'Transitional reporting' }
	]);
	let draft = $state({ v: '', effective: '', note: '' });
	const audit = [
		['verify_success', 'zv-founder-01', '10.4.2.17', 'now'],
		['challenge_issued', 'zv-founder-01', '10.4.2.17', '1 min ago'],
		['verify_failed', 'unknown', '185.220.101.4', '3 h ago'],
		['replay_blocked', 'zv-ops-02', '10.4.2.33', '2 d ago']
	];
</script>

<svelte:head><title>·</title><meta name="robots" content="noindex, nofollow, noarchive" /></svelte:head>

<div class="ov">
	<header>
		<Logo size={22} inverse />
		<span class="mono tag">SUPER-ADMIN VAULT</span>
		<span class="sp"></span>
		<span class="mono who">PGP · {data.admin.handle} · …{data.admin.fingerprint.slice(-16)}</span>
		<a class="mono out" href="/ops-vault/logout">EXIT</a>
	</header>
	<nav>{#each tabs as t, i (t)}<button class:on={tab === t} onclick={() => (tab = t)}><span class="mono">0{i + 1}</span>{t}</button>{/each}</nav>
	<main>
		<p class="mono dim">DEMO OPERATIONS · Sample records and edits in this view stay in this browser session.</p>
		{#if tab === 'Approvals'}
			<h1>Account & transporter approvals</h1>
			{#each approvals as a (a.id)}
				<div class="rowc"><div><span class="mono k">{a.kind}</span><strong>{a.name}</strong><p>{a.detail}</p></div>
					{#if a.decision}<span class="mono st">{a.decision}</span>{:else}<div class="btns"><button onclick={() => (a.decision = 'APPROVED')}>Approve</button><button class="neg" onclick={() => (a.decision = 'REJECTED')}>Reject</button></div>{/if}
				</div>
			{/each}
		{:else if tab === 'Trust & disputes'}
			<h1>Marketplace trust, abuse & disputes</h1>
			{#each OPS.disputes as d (d.id)}<div class="rowc"><div><span class="mono k">{d.ref}</span><strong>{d.parties}</strong><p>{d.issue} · {d.amount}</p></div><span class="mono st">{d.status}</span></div>{/each}
		{:else if tab === 'CBAM rule publisher'}
			<h1>CBAM rule & emission-factor publisher</h1>
			{#each rules as r (r.v)}<div class="rowc"><div><span class="mono k">effective {r.effective}</span><strong>{r.v}</strong><p>{r.note}</p></div><span class="mono st" class:live={r.status === 'live'}>{r.status}</span></div>{/each}
			<div class="draft">
				<input bind:value={draft.v} placeholder="Rule set version" /><input type="date" bind:value={draft.effective} /><input bind:value={draft.note} placeholder="Change note" />
				<button disabled={!draft.v || !draft.effective} onclick={() => { rules = [{ ...draft, status: 'scheduled' }, ...rules]; draft = { v: '', effective: '', note: '' }; }}>Publish (date-effective)</button>
			</div>
		{:else if tab === 'Margin guardrails'}
			<h1>Gross-margin & direct-cost guardrails</h1>
			<div class="mg">
				{#each OPS.margins as m (m.tier)}
					{@const pct = (m.actual / m.ceiling) * 100}
					<div class="card-d">
						<span class="mono k">{m.tier} · ${m.price}/mo</span>
						<div class="big mono" class:warn={pct > 90}>${m.actual.toFixed(2)}</div>
						<p>direct cost / tenant / mo · ceiling ${m.ceiling} ({m.ceiling_pct}% GM) · annual ${m.annual_ceiling}</p>
						<div class="meter"><span style="width:{Math.min(100, pct)}%" class:warn={pct > 90}></span></div>
						<dl>{#each Object.entries(m.breakdown) as [k, v] (k)}<div><dt>{k.replace('_', ' / ')}</dt><dd class="mono">${v}</dd></div>{/each}</dl>
					</div>
				{/each}
			</div>
		{:else if tab === 'FreightLink KPIs'}
			<h1>FreightLink & pilot KPI telemetry</h1>
			<div class="kp">{#each OPS.kpis as [l, v, d] (l)}<div class="card-d"><span class="mono k">{l}</span><div class="big mono">{v}</div><p>{d}</p></div>{/each}</div>
		{:else}
			<h1>System health & immutable PGP audit log</h1>
			<div class="kp">{#each [['auth-api', 'p99 41 ms'], ['core-api', 'p99 88 ms'], ['intel-worker', 'queue 3'], ['postgres', 'replica lag 0.2 s']] as [s, v] (s)}<div class="card-d"><span class="mono k">{s}</span><div class="big mono ok">UP</div><p>{v}</p></div>{/each}</div>
			<table><thead><tr><th>event</th><th>admin</th><th>ip</th><th>when</th></tr></thead><tbody>{#each audit as [e, a, ip, t] (e + t)}<tr class:bad={e !== 'verify_success' && e !== 'challenge_issued'}><td>{e}</td><td>{a}</td><td>{ip}</td><td>{t}</td></tr>{/each}</tbody></table>
		{/if}
	</main>
</div>

<style>
	:global(body:has(.ov)) {
		background: var(--vault-bg);
	}
	.ov {
		min-height: 100dvh;
		background: var(--vault-bg);
		color: #c4d6ce;
	}
	header {
		display: flex;
		align-items: center;
		gap: 14px;
		height: 56px;
		padding: 0 24px;
		border-bottom: 1px solid var(--vault-border);
	}
	.tag {
		font-size: 10.5px;
		letter-spacing: 0.14em;
		color: var(--vault-phosphor);
		padding: 4px 8px;
		border: 1px solid var(--vault-border);
	}
	.sp {
		flex: 1;
	}
	.who,
	.out {
		font-size: 11px;
		color: var(--vault-dim);
	}
	.out:hover {
		color: var(--status-crimson);
	}
	nav {
		display: flex;
		gap: 2px;
		padding: 0 16px;
		border-bottom: 1px solid var(--vault-border);
		overflow-x: auto;
	}
	nav button {
		display: flex;
		gap: 8px;
		height: 44px;
		align-items: center;
		padding: 0 14px;
		font-size: 12.5px;
		color: var(--vault-dim);
		border-bottom: 2px solid transparent;
		white-space: nowrap;
	}
	nav button .mono {
		font-size: 10px;
	}
	nav button.on {
		color: #e8f3ee;
		border-color: var(--vault-phosphor);
	}
	main {
		padding: 28px 24px;
		max-width: 1280px;
		margin: 0 auto;
		animation: zv-fade-in 500ms var(--ease-expo);
	}
	h1 {
		color: #e8f3ee;
		font-size: 1.5rem;
		margin-bottom: 18px;
	}
	.k {
		font-size: 10.5px;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--vault-phosphor);
		display: block;
		margin-bottom: 4px;
	}
	.rowc {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 16px;
		padding: 16px 0;
		border-bottom: 1px solid var(--vault-border);
		font-size: 13px;
	}
	.rowc strong {
		color: #e8f3ee;
	}
	.rowc p {
		font-size: 12px;
		color: var(--vault-dim);
		margin-top: 2px;
	}
	.st {
		font-size: 11px;
		color: var(--vault-dim);
	}
	.st.live {
		color: var(--vault-phosphor);
	}
	.btns {
		display: flex;
		gap: 6px;
	}
	.btns button,
	.draft button {
		height: 32px;
		padding: 0 14px;
		border: 1px solid var(--vault-phosphor);
		color: var(--vault-phosphor);
		font-family: var(--font-mono);
		font-size: 11px;
	}
	.btns .neg {
		border-color: var(--status-crimson);
		color: #e48a8a;
	}
	.draft {
		display: grid;
		grid-template-columns: 1fr 160px 2fr auto;
		gap: 8px;
		margin-top: 18px;
	}
	.draft input {
		height: 34px;
		padding: 0 10px;
		background: var(--vault-panel);
		border: 1px solid var(--vault-border);
		color: #e8f3ee;
		font-family: var(--font-mono);
		font-size: 12px;
	}
	.mg,
	.kp {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 12px;
		margin-bottom: 18px;
	}
	.kp {
		grid-template-columns: repeat(4, 1fr);
	}
	.card-d {
		padding: 18px;
		border: 1px solid var(--vault-border);
		background: var(--vault-panel);
	}
	.card-d p {
		font-size: 11.5px;
		color: var(--vault-dim);
		margin-top: 6px;
	}
	.big {
		font-size: 2rem;
		color: #e8f3ee;
		letter-spacing: -0.04em;
	}
	.big.warn {
		color: var(--status-terracotta);
	}
	.big.ok {
		color: var(--vault-phosphor);
	}
	.meter {
		height: 3px;
		background: var(--vault-border);
		margin: 12px 0;
	}
	.meter span {
		display: block;
		height: 100%;
		background: var(--vault-phosphor);
	}
	.meter span.warn {
		background: var(--status-terracotta);
	}
	dl div {
		display: flex;
		justify-content: space-between;
		font-size: 11.5px;
		padding: 4px 0;
		border-top: 1px solid var(--vault-border);
		text-transform: capitalize;
	}
	table {
		width: 100%;
		font-family: var(--font-mono);
		font-size: 12px;
	}
	th,
	td {
		text-align: left;
		padding: 10px 8px;
		border-bottom: 1px solid var(--vault-border);
	}
	th {
		color: var(--vault-dim);
		font-weight: 400;
		font-size: 10.5px;
		text-transform: uppercase;
	}
	tr.bad td:first-child {
		color: #e48a8a;
	}
	@media (max-width: 900px) {
		.mg,
		.kp,
		.draft {
			grid-template-columns: 1fr;
		}
	}
</style>
