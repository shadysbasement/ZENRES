<script lang="ts">
	import Logo from '$design/components/Logo.svelte';
	import { onMount } from 'svelte';
	import WorkspaceBackdrop from '$lib/components/WorkspaceBackdrop.svelte';
	import AppearancePanel from '$lib/components/AppearancePanel.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { appearance } from '$lib/stores/appearance.svelte';
	let appearanceOpen = $state(false);
	onMount(() => appearance.init());
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
	<WorkspaceBackdrop/>
	<header>
		<Logo size={26} />
		<span class="tag">Admin workspace</span><a class="btn btn-sm out" href="/ops-vault/support">Support operations ↗</a>
		<span class="sp"></span>
		<span class="mono who">PGP · {data.admin.handle} · …{data.admin.fingerprint.slice(-16)}</span>
		<button class="appearance-button" aria-label="Customize appearance" onclick={() => (appearanceOpen = true)}><Icon name="settings" size={18}/></button><a class="btn btn-sm out" href="/ops-vault/logout">Exit vault ↗</a>
	</header>
	<nav>{#each tabs as t, i (t)}<button class:on={tab === t} onclick={() => (tab = t)}><span class="mono">0{i + 1}</span>{t}</button>{/each}</nav>
	<main>
		<div class="admin-welcome"><span>Welcome back, {data.admin.handle}</span><h2>Your operations, in focus.</h2></div>
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
			<h1>Sample system signals & audit events</h1>
			<div class="kp">{#each [['auth-api', 'Example latency: p99 41 ms'], ['core-api', 'Example latency: p99 88 ms'], ['intel-worker', 'Example queue depth: 3'], ['postgres', 'Example replica lag: 0.2 s']] as [s, v] (s)}<div class="card-d"><span class="mono k">{s}</span><div class="big mono">DEMO</div><p>{v}</p></div>{/each}</div>
			<table><thead><tr><th>event</th><th>admin</th><th>ip</th><th>when</th></tr></thead><tbody>{#each audit as [e, a, ip, t] (e + t)}<tr class:bad={e !== 'verify_success' && e !== 'challenge_issued'}><td>{e}</td><td>{a}</td><td>{ip}</td><td>{t}</td></tr>{/each}</tbody></table>
		{/if}
	</main>
</div>

<AppearancePanel open={appearanceOpen} onclose={() => (appearanceOpen = false)}/>

<style>
	:global(body:has(.ov)) {
		background: transparent;
	}
	.ov {
		min-height: 100dvh;
		background: transparent;
		color: var(--ink-secondary);
		position: relative;
		isolation: isolate;
	}
	header {
		display: flex;
		align-items: center;
		gap: 14px;
		height: 56px;
		padding: 0 24px;
		border-bottom: 1px solid var(--border-hairline);
	}
	.tag {
		font-size: 10.5px;
		letter-spacing: 0.14em;
		color: #6986b0;
		padding: 4px 8px;
		border: 1px solid var(--border-hairline);
	}
	.sp {
		flex: 1;
	}
	.who,
	.out {
		font-size: 11px;
		color: var(--ink-muted);
	}
	.out:hover {
		color: var(--status-crimson);
	}
	nav {
		display: flex;
		gap: 2px;
		padding: 0 16px;
		border-bottom: 1px solid var(--border-hairline);
		overflow-x: auto;
	}
	nav button {
		display: flex;
		gap: 8px;
		height: 44px;
		align-items: center;
		padding: 0 14px;
		font-size: 12.5px;
		color: var(--ink-muted);
		border-bottom: 2px solid transparent;
		white-space: nowrap;
	}
	nav button .mono {
		font-size: 10px;
	}
	nav button.on {
		color: var(--ink-primary);
		border-color: #6986b0;
	}
	main {
		padding: 28px 24px;
		max-width: 1280px;
		margin: 0 auto;
		animation: zv-fade-in 500ms var(--ease-expo);
	}
	h1 {
		color: var(--ink-primary);
		font-size: 1.5rem;
		margin-bottom: 18px;
	}
	.k {
		font-size: 10.5px;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: #6986b0;
		display: block;
		margin-bottom: 4px;
	}
	.rowc {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 16px;
		padding: 16px 0;
		border-bottom: 1px solid var(--border-hairline);
		font-size: 13px;
	}
	.rowc strong {
		color: var(--ink-primary);
	}
	.rowc p {
		font-size: 12px;
		color: var(--ink-muted);
		margin-top: 2px;
	}
	.st {
		font-size: 11px;
		color: var(--ink-muted);
	}
	.st.live {
		color: #6986b0;
	}
	.btns {
		display: flex;
		gap: 6px;
	}
	.btns button,
	.draft button {
		height: 32px;
		padding: 0 14px;
		border: 1px solid #c7d6ed;
		color: #6986b0;
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
		background: #fcfdfff0;
		border: 1px solid var(--border-hairline);
		color: var(--ink-primary);
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
		border: 1px solid var(--border-hairline);
		background: #fcfdfff0;
	}
	.card-d p {
		font-size: 11.5px;
		color: var(--ink-muted);
		margin-top: 6px;
	}
	.big {
		font-size: 2rem;
		color: var(--ink-primary);
		letter-spacing: -0.04em;
	}
	.big.warn {
		color: var(--status-terracotta);
	}
	.meter {
		height: 3px;
		background: var(--border-hairline);
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
		border-top: 1px solid var(--border-hairline);
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
		border-bottom: 1px solid var(--border-hairline);
	}
	th {
		color: var(--ink-muted);
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

    header { height: 100px; padding-inline: clamp(20px, 4vw, 64px); border: 0; }
    .tag { border: 0; border-radius: 100px; background: #fcfdff80; padding: 5px 12px; font-size: 12px; letter-spacing: 0; }
    .appearance-button { width: 36px; height: 36px; display: grid; place-items: center; background: #fcfdffb0; border: 1px solid white; border-radius: 50%; }
    nav { justify-content: center; gap: 8px; border: 0; padding: 18px 24px; }
    nav button { border: 1px solid #ffffffdf; border-radius: 100px; background: #fcfdffa8; transition: background 250ms, transform 300ms var(--ease-out); }
    nav button:hover { transform: translateY(-2px); background: #fcfdff; }
    nav button.on { background: linear-gradient(100deg, #a4d4f6, #cfb7ee); color: #3b4d68; border-color: #fff; }
    nav button .mono { font-family: var(--font-sans); }
    main { padding-top: 24px; }
    .admin-welcome { margin: 0 0 25px; }
    .admin-welcome > span { font-size: 13px; color: var(--ink-muted); }
    .admin-welcome h2 { font-size: 32px; font-weight: 450; margin-top: 8px; }
    main > .dim { font-size: 10px; color: var(--ink-muted); margin-bottom: 24px; }
    h1 { font-size: 21px; font-weight: 500; }
    .card-d, .rowc { border-radius: 20px; border: 1px solid #fff; background: #fcfdffed; padding: 24px; backdrop-filter: blur(15px); }
    .rowc { margin-bottom: 12px; }
    .card-d p { font-size: 12px; }
    .k { font-family: var(--font-sans); font-size: 12px; letter-spacing: 0; text-transform: capitalize; }
    .big { font-family: var(--font-sans); font-size: 42px; font-weight: 450; margin: 20px 0 10px; }
    .meter { height: 5px; border-radius: 20px; overflow: hidden; background: #e7edf5; }
    .meter span { background: linear-gradient(90deg, #91c9f2, #bda5e8); }
    .btns button, .draft button, .draft input { border-radius: 100px; height: 38px; font-family: var(--font-sans); }
    .draft button { background: linear-gradient(100deg, #bbdcf5, #d8c5f2); }
    table { background: #fcfdffeb; border-radius: 20px; overflow: hidden; }
    th, td { padding: 16px; }
    @media (max-width: 900px) { header { height: auto; min-height: 100px; flex-wrap: wrap; padding: 22px; } .who { order: 5; width: 100%; } nav { justify-content: flex-start; } .tag { display: none; } }
</style>
