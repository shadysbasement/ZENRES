<script lang="ts">
	import { goto } from '$app/navigation';
	import Drawer from './ui/Drawer.svelte';
	import Icon from './ui/Icon.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { timeAgo, dueIn } from '$lib/format';

	let tab = $state<'buyer_agency' | 'deadline' | 'operational_alert'>('buyer_agency');
	const tabs = [
		{ id: 'buyer_agency', label: 'Buyers & Agencies' },
		{ id: 'deadline', label: 'Deadlines' },
		{ id: 'operational_alert', label: 'Alerts' }
	] as const;
	const items = $derived(ws.notifications.filter((n) => n.category === tab));
	const unreadBy = (c: string) => ws.notifications.filter((n) => n.category === c && !n.read).length;

	function open(href: string, id: string) {
		ws.markRead(id);
		ws.bellOpen = false;
		goto(href);
	}
</script>

<Drawer open={ws.bellOpen} title="Notifications" eyebrow="Connected mail · {ws.unread} unread" width={460} onclose={() => (ws.bellOpen = false)}>
	<div class="tabs" role="tablist">
		{#each tabs as t (t.id)}
			<button role="tab" aria-selected={tab === t.id} onclick={() => (tab = t.id)}>
				{t.label}
				{#if unreadBy(t.id)}<span class="n mono">{unreadBy(t.id)}</span>{/if}
			</button>
		{/each}
	</div>

	<div class="list">
		{#each items as n, i (n.id)}
			<article class="item" class:unread={!n.read} style="--i:{i}">
				<div class="row1">
					{#if n.category === 'buyer_agency'}
						<span class="org">{n.org_badge ?? n.sender_name}</span>
					{:else if n.deadline}
						{@const d = dueIn(n.deadline)}
						<span class="pill pill-{d.tone === 'muted' ? 'outline' : d.tone}">{d.label}</span>
					{:else}
						<span class="pill" class:pill-terracotta={n.severity === 'elevated'} class:pill-crimson={n.severity === 'critical'}>{n.severity === 'normal' ? 'Event' : n.severity}</span>
					{/if}
					<span class="time mono">{timeAgo(n.received_at)}</span>
				</div>
				<h3>{n.subject}</h3>
				<p>{n.snippet}</p>
				<div class="meta">
					{#if n.tag}<span class="pill pill-cobalt">{n.tag}</span>{/if}
					{#if n.category === 'buyer_agency' && n.deadline}{@const d = dueIn(n.deadline)}<span class="pill pill-{d.tone === 'muted' ? 'outline' : d.tone}"><Icon name="clock" size={11} />{d.label}</span>{/if}
				</div>
				<div class="actions">
					{#if n.category === 'buyer_agency'}
						<button class="btn btn-sm" onclick={() => open('/inbox?thread=' + n.id, n.id)}>Open thread</button>
						<button class="btn btn-sm btn-primary" onclick={() => ws.convertToRfq(n.id)}>Convert to RFQ</button>
						<button class="btn btn-sm btn-ghost" onclick={() => open('/carbon/exports', n.id)}>Attach CBAM pack</button>
					{:else if n.category === 'deadline'}
						{#if n.link}<button class="btn btn-sm btn-dark" onclick={() => open(n.link!.href, n.id)}>{n.link.label}</button>{/if}
						<button class="btn btn-sm btn-ghost" onclick={() => ws.snooze(n.id)}>Snooze 24h</button>
					{:else if n.link}
						<button class="btn btn-sm" onclick={() => open(n.link!.href, n.id)}>{n.link.label}</button>
					{/if}
					{#if !n.read}<button class="btn btn-sm btn-ghost" onclick={() => ws.markRead(n.id)}>Mark read</button>{/if}
				</div>
			</article>
		{:else}
			<p class="empty">Nothing here — you’re all caught up.</p>
		{/each}
	</div>

	{#snippet footer()}
		<button class="btn btn-sm btn-ghost" onclick={() => ws.markAllRead(tab)}>Mark all read</button>
		<a class="btn btn-sm" href="/inbox" onclick={() => (ws.bellOpen = false)}>Open inbox <span class="arrow">→</span></a>
	{/snippet}
</Drawer>

<style>
	.tabs {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 4px;
		padding: 3px;
		border-radius: var(--radius-sm);
		background: var(--surface-subtle);
		margin-bottom: 14px;
	}
	.tabs button {
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		border-radius: 3px;
		font-size: 12px;
		font-weight: 500;
		color: var(--ink-muted);
	}
	.tabs button[aria-selected='true'] {
		background: var(--surface-card);
		color: var(--ink-primary);
		box-shadow: 0 0 0 1px var(--border-hairline);
	}
	.n {
		min-width: 16px;
		height: 16px;
		border-radius: 8px;
		background: var(--brand-emerald);
		color: #fff;
		font-size: 9.5px;
		display: grid;
		place-items: center;
	}
	.list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.item {
		position: relative;
		padding: 14px;
		border: var(--hairline);
		border-radius: var(--radius-md);
		display: flex;
		flex-direction: column;
		gap: 6px;
		animation: in 500ms var(--ease-expo) both;
		animation-delay: calc(var(--i) * 40ms);
	}
	@keyframes in {
		from { opacity: 0; transform: translateX(10px); }
	}
	.item.unread {
		background: linear-gradient(90deg, var(--brand-emerald-tint), transparent 40%);
		border-color: #cfe1d7;
	}
	.item.unread::before {
		content: '';
		position: absolute;
		left: -1px;
		top: 14px;
		width: 3px;
		height: 18px;
		border-radius: 0 2px 2px 0;
		background: var(--brand-emerald);
	}
	.row1 {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 8px;
	}
	.org {
		font-size: 11.5px;
		font-weight: 600;
		color: var(--accent-cobalt);
	}
	.time {
		font-size: 10.5px;
		color: var(--ink-faint);
		white-space: nowrap;
	}
	h3 {
		font-size: 13.5px;
		font-weight: 600;
		line-height: 1.35;
	}
	p {
		font-size: 12.5px;
		line-height: 1.5;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.meta {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.actions {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
		margin-top: 4px;
	}
	.actions .btn {
		height: 28px;
		padding: 0 10px;
		font-size: 11.5px;
	}
	.empty {
		padding: 40px 0;
		text-align: center;
		color: var(--ink-muted);
		font-size: 13px;
	}
</style>
