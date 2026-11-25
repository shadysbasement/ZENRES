<script lang="ts">
	import { page } from '$app/state';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { ORG } from '$lib/demo/seed';
	import { timeAgo, dueIn, fmtDate } from '$lib/format';

	let filter = $state<'all' | 'buyer_agency' | 'deadline' | 'operational_alert'>('all');
	let selected = $state<string | null>(page.url.searchParams.get('thread'));
	let showConnect = $state(false);
	let reply = $state('');
	const list = $derived(ws.notifications.filter((n) => filter === 'all' || n.category === filter));
	const cur = $derived(ws.notifications.find((n) => n.id === selected) ?? list[0]);
	$effect(() => {
		if (cur && !cur.read) ws.markRead(cur.id);
	});

	const connections = [
		{ id: 'gmail', name: 'Gmail', mode: 'OAuth2 · read-only + optional reply', status: 'Connected', account: 'exports@deccanalloy.in' },
		{ id: 'outlook', name: 'Microsoft 365 / Outlook', mode: 'Graph webhooks', status: 'Available', account: '' },
		{ id: 'imap', name: 'IMAP / SMTP', mode: 'TLS required · Zoho, cPanel, Proton Bridge', status: 'Available', account: '' },
		{ id: 'inbound', name: 'ZENRES inbound address', mode: 'Zero-config forwarding', status: 'Active', account: ORG.inbound }
	];
	const rules = [
		['Domain', '*.de, *.nl, *.it, *.fr', 'Buyers & Agencies'],
		['Keywords', 'RFQ, CBAM XML, sample order, container specs, price quote', 'Buyers & Agencies'],
		['Sender', 'arjun@euro-sourcing-agents.com', 'Buying agency'],
		['Extracted', 'due / by / deadline + date', 'Deadlines']
	];

	function copyInbound() {
		navigator.clipboard.writeText(ORG.inbound);
		ws.toast('Forwarding address copied');
	}
	function sendReply() {
		ws.toast(`Reply sent to ${cur?.sender_email} via Gmail`);
		reply = '';
	}
</script>

<svelte:head><title>Inbox — ZENRES</title></svelte:head>

<PageHeader eyebrow="ZENRES Business · A6" title="Connected mail &" serif="buying agency inbox" lede="Buyer RFQs, agency requests and compliance deadlines classified from your business mailboxes — feeding the top-right bell.">
	{#snippet actions()}
		<button class="btn btn-sm" onclick={() => (showConnect = !showConnect)}><Icon name="link" size={14} />Mail connections</button>
		<button class="btn btn-sm btn-dark" onclick={() => ws.simulateInbound()}><Icon name="mail" size={14} />Send mock inbound webhook</button>
	{/snippet}
</PageHeader>

{#if showConnect}
	<section class="panel conns fade-in">
		<div class="cgrid">
			{#each connections as c (c.id)}
				<div class="conn">
					<div class="row"><strong>{c.name}</strong><span class="pill" class:pill-emerald={c.status !== 'Available'}>{c.status}</span></div>
					<span class="muted small">{c.mode}</span>
					{#if c.id === 'inbound'}
						<button class="addr mono" onclick={copyInbound}>{c.account} <Icon name="link" size={12} /></button>
					{:else if c.account}<span class="mono small">{c.account}</span>{:else}<button class="btn btn-sm">Connect</button>{/if}
				</div>
			{/each}
		</div>
		<div class="rules">
			<span class="eyebrow">Buying-agency sender rules</span>
			{#each rules as [k, v, cat] (v)}<div class="rule"><span class="pill pill-outline">{k}</span><span class="mono small">{v}</span><span class="muted small">→ {cat}</span></div>{/each}
		</div>
		<p class="fine">OAuth refresh tokens and IMAP credentials are envelope-encrypted (KMS). ZENRES never stores mail you have not routed to it.</p>
	</section>
{/if}

<div class="mail">
	<aside class="panel listp">
		<div class="filters">
			{#each [['all', 'All'], ['buyer_agency', 'Buyers'], ['deadline', 'Deadlines'], ['operational_alert', 'Alerts']] as [id, l] (id)}
				<button class:on={filter === id} onclick={() => (filter = id as typeof filter)}>{l}</button>
			{/each}
		</div>
		<ul role="list">
			{#each list as n (n.id)}
				<li>
					<button class:sel={cur?.id === n.id} class:unread={!n.read} onclick={() => (selected = n.id)}>
						<span class="r1"><span class="from">{n.sender_name}</span><span class="mono t">{timeAgo(n.received_at)}</span></span>
						<span class="subj">{n.subject}</span>
						<span class="snip">{n.snippet}</span>
					</button>
				</li>
			{/each}
		</ul>
	</aside>

	{#if cur}
		<article class="panel reader">
			<header>
				<div>
					{#if cur.org_badge}<span class="pill pill-cobalt">{cur.org_badge}</span>{/if}
					<h2>{cur.subject}</h2>
					<p class="muted small">{cur.sender_name} &lt;{cur.sender_email}&gt; · {fmtDate(cur.received_at)}</p>
				</div>
				<div class="tags">
					{#if cur.tag}<span class="pill pill-outline">{cur.tag}</span>{/if}
					{#if cur.deadline}{@const d = dueIn(cur.deadline)}<span class="pill pill-{d.tone === 'muted' ? 'outline' : d.tone}">Deadline · {d.label}</span>{/if}
				</div>
			</header>
			<div class="body">
				<p>Dear Deccan Alloy Works team,</p>
				<p>{cur.snippet}</p>
				<p>Kind regards,<br />{cur.sender_name}</p>
			</div>
			<div class="extract card">
				<span class="eyebrow">Triage extraction · intel-worker</span>
				<dl>
					<div><dt>Category</dt><dd>{cur.category.replace('_', ' & ')}</dd></div>
					<div><dt>Detected CN</dt><dd class="mono">{cur.tag ?? '—'}</dd></div>
					<div><dt>Deadline</dt><dd>{cur.deadline ? fmtDate(cur.deadline) : '—'}</dd></div>
					<div><dt>Severity</dt><dd>{cur.severity}</dd></div>
				</dl>
			</div>
			<div class="acts">
				{#if cur.category === 'buyer_agency'}
					<button class="btn btn-sm btn-primary" onclick={() => ws.convertToRfq(cur.id)}>Convert to ZENRES RFQ</button>
					<a class="btn btn-sm" href="/carbon/audit">Link attachment to Evidence Vault</a>
					<a class="btn btn-sm" href="/carbon/exports">Attach CBAM pack</a>
				{:else if cur.link}
					<a class="btn btn-sm btn-dark" href={cur.link.href}>{cur.link.label}</a>
				{/if}
			</div>
			{#if cur.category === 'buyer_agency'}
				<div class="reply">
					<textarea class="textarea" bind:value={reply} placeholder="Reply to {cur.sender_name}…"></textarea>
					<button class="btn btn-sm btn-dark" disabled={!reply} onclick={sendReply}>Send reply</button>
				</div>
			{/if}
		</article>
	{/if}
</div>

<style>
	.conns {
		margin-bottom: 16px;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.cgrid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 10px;
	}
	.conn {
		padding: 14px;
		border: var(--hairline);
		border-radius: var(--radius-sm);
		display: flex;
		flex-direction: column;
		gap: 8px;
		align-items: flex-start;
		font-size: 13px;
	}
	.row {
		display: flex;
		justify-content: space-between;
		width: 100%;
		gap: 8px;
	}
	.addr {
		display: flex;
		gap: 6px;
		align-items: center;
		font-size: 11px;
		padding: 6px 8px;
		border-radius: 3px;
		background: var(--brand-emerald-tint);
		color: var(--brand-emerald);
	}
	.rules {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.rule {
		display: grid;
		grid-template-columns: 90px 1fr auto;
		gap: 10px;
		align-items: center;
	}
	.small {
		font-size: 11.5px;
	}
	.mail {
		display: grid;
		grid-template-columns: 380px 1fr;
		gap: 16px;
		align-items: start;
	}
	.listp {
		padding: 10px;
		position: sticky;
		top: 76px;
		max-height: calc(100dvh - 100px);
		overflow: auto;
	}
	.filters {
		display: flex;
		gap: 4px;
		padding: 4px;
		margin-bottom: 6px;
	}
	.filters button {
		height: 28px;
		padding: 0 10px;
		border-radius: 3px;
		font-size: 12px;
		color: var(--ink-muted);
	}
	.filters button.on {
		background: var(--obsidian);
		color: var(--ink-inverse);
	}
	.listp ul button {
		width: 100%;
		text-align: left;
		display: flex;
		flex-direction: column;
		gap: 3px;
		padding: 12px;
		border-radius: var(--radius-sm);
		transition: background 200ms;
	}
	.listp ul button:hover {
		background: var(--surface-canvas);
	}
	.listp ul button.sel {
		background: var(--brand-emerald-tint);
	}
	.r1 {
		display: flex;
		justify-content: space-between;
		gap: 8px;
	}
	.from {
		font-size: 12.5px;
		color: var(--ink-primary);
	}
	.unread .from,
	.unread .subj {
		font-weight: 600;
	}
	.t {
		font-size: 10.5px;
		color: var(--ink-faint);
	}
	.subj {
		font-size: 12.5px;
		color: var(--ink-primary);
	}
	.snip {
		font-size: 12px;
		color: var(--ink-muted);
		display: -webkit-box;
		-webkit-line-clamp: 1;
		line-clamp: 1;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.reader {
		padding: 24px 28px;
		display: flex;
		flex-direction: column;
		gap: 18px;
		animation: zv-fade-in 500ms var(--ease-expo);
	}
	.reader header {
		display: flex;
		justify-content: space-between;
		gap: 16px;
		align-items: flex-start;
	}
	.reader h2 {
		font-size: 1.4rem;
		margin: 10px 0 4px;
	}
	.tags {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
		justify-content: flex-end;
	}
	.body {
		display: flex;
		flex-direction: column;
		gap: 12px;
		font-size: 14px;
		line-height: 1.65;
		color: var(--ink-primary);
		max-width: 70ch;
	}
	.extract {
		padding: 14px;
	}
	.extract dl {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 10px;
		margin-top: 10px;
	}
	.extract dt {
		font-size: 11px;
		color: var(--ink-muted);
	}
	.extract dd {
		font-size: 13px;
		color: var(--ink-primary);
		text-transform: capitalize;
	}
	.acts {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.reply {
		display: flex;
		flex-direction: column;
		gap: 8px;
		align-items: flex-end;
	}
	@media (max-width: 1100px) {
		.mail,
		.cgrid {
			grid-template-columns: 1fr;
		}
		.listp {
			position: static;
		}
	}
</style>
