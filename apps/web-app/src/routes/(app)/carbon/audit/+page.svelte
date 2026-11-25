<script lang="ts">
	import { onMount } from 'svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HintWhisper from '$lib/components/HintWhisper.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { VERIFIERS } from '$lib/demo/seed';
	import { buildChain, verifyChain, sha256Hex } from '$lib/carbon';

	const chainCap = $derived(ws.manifest.capabilities.has('tamper_audit_locker'));
	const caCap = $derived(ws.manifest.capabilities.has('priority_ca_workspace'));
	const customs = $derived(ws.manifest.capabilities.has('customs_audit_pack'));

	type Entry = { seq: number; at: string; actor: string; action: string; object: string; sha256: string; prev_hash: string; hash: string };
	let chain = $state<Entry[]>([]);
	let status = $state<'idle' | 'verifying' | 'ok' | 'broken'>('idle');
	let brokenAt = $state(-1);
	let evidence = $state<{ name: string; kind: string; sha: string; size: string }[]>([]);

	onMount(async () => {
		evidence = await Promise.all(
			ws.activity.map(async (a) => ({ name: a.file, kind: a.type, sha: await sha256Hex(a.file + a.qty), size: `${(120 + (a.qty % 900)).toFixed(0)} KB` }))
		);
		const events = [
			{ seq: 1, at: '2026-11-02T10:14:00Z', actor: 'Meera Joshi', action: 'upload', object: 'MSEDCL_HT_Sep26.pdf', sha256: evidence[4]?.sha ?? '' },
			{ seq: 2, at: '2026-11-02T10:16:00Z', actor: 'system', action: 'calculate', object: 'Q3 calculation v3', sha256: '' },
			{ seq: 3, at: '2026-11-05T15:40:00Z', actor: 'Rakesh Patil', action: 'upload', object: 'DG_log_Oct26.xlsx', sha256: evidence[1]?.sha ?? '' },
			{ seq: 4, at: '2026-11-09T11:02:00Z', actor: 'Meera Joshi', action: 'edit', object: 'Precursor allocation ALB-6063 0.76 → 0.78', sha256: '' },
			{ seq: 5, at: '2026-11-09T11:05:00Z', actor: 'system', action: 'calculate', object: 'Q3 calculation v4', sha256: '' },
			{ seq: 6, at: '2026-11-18T09:42:00Z', actor: 'system', action: 'anchor', object: 'Merkle root 9f2c4be1…e41a', sha256: '' }
		];
		chain = (await buildChain(events)) as Entry[];
	});

	async function verify() {
		status = 'verifying';
		await new Promise((r) => setTimeout(r, 600));
		brokenAt = await verifyChain(chain);
		status = brokenAt === -1 ? 'ok' : 'broken';
	}
	function tamper() {
		chain[3] = { ...chain[3], object: 'Precursor allocation ALB-6063 0.76 → 0.70' };
		status = 'idle';
		ws.toast('Entry #4 silently edited — run verification', 'warn');
	}
	async function restore() {
		chain[3] = { ...chain[3], object: 'Precursor allocation ALB-6063 0.76 → 0.78' };
		await verify();
	}

	let comments = $state([
		{ by: 'R. Kulkarni (CA)', text: 'Precursor allocation for ALB-6063 moved from 0.76 to 0.78 t/t — please attach the supplier EPD page.', open: true },
		{ by: 'Meera Joshi', text: 'Attached Hindalco_EPD_2026.pdf (p. 14). Allocation follows melt-shop yield records.', open: true },
		{ by: 'R. Kulkarni (CA)', text: 'October Scope 2 is estimated; we need the HT bill before sign-off.', open: true }
	]);
	let reply = $state('');
	const signoff = $state([
		{ label: 'Monitoring methodology reviewed', done: true },
		{ label: 'Activity data traced to evidence', done: true },
		{ label: 'Precursor allocation agreed', done: false },
		{ label: 'Final sign-off', done: false }
	]);
	function pack() {
		ws.toast('Typst rendering EU Customs Audit Defence Pack (PDF)…');
		setTimeout(() => window.print(), 400);
	}
</script>

<svelte:head><title>Audit Locker — Zevra</title></svelte:head>

<PageHeader eyebrow="Zevra Carbon · B4" title="Audit evidence locker &" serif="verification" lede="Every bill, meter reading, invoice and calculation version in one encrypted repository — linked to the numbers they support.">
	{#snippet actions()}
		<button class="btn btn-sm" onclick={() => ws.hint('audit', { kind: 'manual_effort', payload: 'revised_upload' })}><Icon name="upload" size={14} />Upload revised version</button>
		{#if customs}<button class="btn btn-sm btn-primary" onclick={pack}><Icon name="shield" size={14} />EU Customs Audit Defence Pack</button>{/if}
	{/snippet}
</PageHeader>

<HintWhisper context="audit" />

<div class="g2">
	<section class="panel">
		<div class="panel-head"><h2>Evidence vault</h2><span class="pill pill-outline">AES-256 · S3/R2</span></div>
		<div class="table-wrap">
			<table class="table">
				<thead><tr><th>File</th><th>Linked to</th><th>Size</th><th>SHA-256</th></tr></thead>
				<tbody>{#each evidence as e (e.name)}<tr><td class="strong">{e.name}</td><td class="cap">{e.kind}</td><td class="mono small">{e.size}</td><td class="mono small">{e.sha.slice(0, 12)}…{e.sha.slice(-6)}</td></tr>{/each}</tbody>
			</table>
		</div>
	</section>

	{#if chainCap}
		<section class="panel">
			<div class="panel-head">
				<h2>Tamper-evident hash chain</h2>
				<span class="pill" class:pill-emerald={status === 'ok'} class:pill-crimson={status === 'broken'}>{status === 'ok' ? 'Chain intact' : status === 'broken' ? `Broken at #${brokenAt + 1}` : status === 'verifying' ? 'Verifying…' : 'Not verified'}</span>
			</div>
			<ol class="chain" role="list">
				{#each chain as c, i (c.seq)}
					<li class:bad={status === 'broken' && i >= brokenAt}>
						<span class="node"></span>
						<div>
							<div class="r"><span class="strong">{c.action}</span> · {c.object}</div>
							<div class="mono small muted">{c.actor} · {c.at.slice(0, 16).replace('T', ' ')} · {c.hash.slice(0, 16)}…</div>
						</div>
					</li>
				{/each}
			</ol>
			<div class="row">
				<button class="btn btn-sm btn-dark" onclick={verify}>Verify chain</button>
				<button class="btn btn-sm btn-ghost" onclick={tamper}>Simulate silent edit</button>
				{#if status === 'broken'}<button class="btn btn-sm btn-ghost" onclick={restore}>Restore</button>{/if}
			</div>
			<p class="fine">Each entry hashes SHA-256(prev_hash ‖ canonical JSON). Any edit breaks every subsequent link.</p>
		</section>
	{/if}
</div>

<div class="g2" style="margin-top:16px">
	{#if caCap}
		<section class="panel">
			<div class="panel-head"><h2>Priority CA / verifier workspace</h2><span class="pill pill-cobalt">R. Kulkarni & Associates</span></div>
			<div class="thread">
				{#each comments as c, i (i)}<div class="cm" class:me={c.by === 'Meera Joshi'}><span class="by">{c.by}</span><p>{c.text}</p></div>{/each}
			</div>
			<div class="rep"><input class="input" bind:value={reply} placeholder="Reply to the verifier…" /><button class="btn btn-sm btn-dark" disabled={!reply} onclick={() => { comments.push({ by: 'Meera Joshi', text: reply, open: true }); reply = ''; }}>Send</button></div>
			<div class="so">
				<span class="eyebrow">Sign-off checklist</span>
				{#each signoff as s (s.label)}<label><input type="checkbox" bind:checked={s.done} />{s.label}</label>{/each}
			</div>
		</section>
	{/if}
	<section class="panel">
		<div class="panel-head"><h2>Verifier marketplace</h2><span class="muted small">Referrals disclosed transparently</span></div>
		{#each VERIFIERS as v (v.name)}
			<div class="ver">
				<div><strong>{v.name}</strong><div class="muted small">{v.type} · {v.city} · {v.accredited}</div></div>
				<span class="mono">★ {v.rating}</span>
				<button class="btn btn-sm">Request quote</button>
			</div>
		{/each}
		<p class="fine">Zevra may receive a disclosed referral fee from partners marked as such. It never affects verification outcomes.</p>
	</section>
	{#if customs}
		<section class="panel am">
			<div class="panel-head"><h2>Dedicated account manager</h2></div>
			<p><strong>Katrin Vogel</strong> · EU customs & CBAM · Düsseldorf</p>
			<p class="small muted">Next review: 3 Dec 2026 · Consolidated Q4 pack across 15 suppliers</p>
			<button class="btn btn-sm" onclick={() => ws.toast('Meeting request sent to Katrin Vogel')}>Book a review</button>
		</section>
	{/if}
</div>

<style>
	.g2 {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 16px;
		align-items: start;
	}
	.strong {
		color: var(--ink-primary);
		font-weight: 500;
	}
	.small {
		font-size: 11.5px;
	}
	.cap {
		text-transform: capitalize;
	}
	.chain {
		position: relative;
		display: flex;
		flex-direction: column;
	}
	.chain li {
		position: relative;
		display: grid;
		grid-template-columns: 22px 1fr;
		gap: 8px;
		padding: 8px 0;
		font-size: 12.5px;
	}
	.chain li::before {
		content: '';
		position: absolute;
		left: 5px;
		top: 0;
		bottom: 0;
		border-left: 1px solid var(--border-hairline);
	}
	.node {
		position: relative;
		width: 11px;
		height: 11px;
		margin-top: 4px;
		border-radius: 50%;
		background: var(--brand-emerald);
		box-shadow: 0 0 0 3px var(--brand-emerald-tint);
		transition: all 400ms;
	}
	.bad .node {
		background: var(--status-crimson);
		box-shadow: 0 0 0 3px var(--status-crimson-tint);
	}
	.bad .r {
		color: var(--status-crimson);
	}
	.row {
		display: flex;
		gap: 8px;
		margin-top: 10px;
	}
	.thread {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.cm {
		max-width: 88%;
		padding: 10px 12px;
		border-radius: 8px;
		background: var(--surface-subtle);
		font-size: 12.5px;
	}
	.cm.me {
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
		margin: 12px 0;
	}
	.so {
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 12.5px;
		accent-color: var(--brand-emerald);
	}
	.so label {
		display: flex;
		gap: 8px;
	}
	.ver {
		display: grid;
		grid-template-columns: 1fr auto auto;
		gap: 12px;
		align-items: center;
		padding: 12px 0;
		border-bottom: var(--hairline);
		font-size: 13px;
	}
	.am {
		display: flex;
		flex-direction: column;
		gap: 8px;
		font-size: 13px;
	}
	@media (max-width: 1100px) {
		.g2 {
			grid-template-columns: 1fr;
		}
	}
</style>
