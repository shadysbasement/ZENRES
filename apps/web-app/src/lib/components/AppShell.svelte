<script lang="ts">
	import type { Snippet } from 'svelte';
	import { page } from '$app/state';
	import Logo from '$design/components/Logo.svelte';
	import Icon from './ui/Icon.svelte';
	import MailBellDrawer from './MailBellDrawer.svelte';
	import CommandPalette from './CommandPalette.svelte';
	import Toasts from './ui/Toasts.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { USER } from '$lib/demo/seed';
	import type { SubscriptionTier } from '$lib/entitlements/manifest';
	import { NAV } from '$lib/nav';
	import { goto } from '$app/navigation';
	import { moduleForPath } from '$lib/entitlements/manifest';

	let { children }: { children: Snippet } = $props();

	// Zero-lock navigation: groups/items are built strictly from visible_modules.
	const nav = $derived(
		NAV.map((g) => ({ ...g, items: g.items.filter((i) => ws.manifest.visible_modules.includes(i.key)) })).filter((g) => g.items.length)
	);

	const isActive = (key: string) => {
		const p = page.url.pathname.replace(/^\//, '');
		return p === key || p.startsWith(key + '/');
	};

	let menuOpen = $state(false);
	let collapsed = $state(false);
	let mobileNav = $state(false);
	const tiers: { id: SubscriptionTier; label: string }[] = [
		{ id: 'starter', label: 'MSME Starter' },
		{ id: 'growth', label: 'MSME Growth' },
		{ id: 'importer_hub', label: 'EU Importer Hub' }
	];

	const crumbs = $derived.by(() => {
		for (const g of NAV) for (const i of g.items) if (isActive(i.key)) return [g.group, i.label];
		if (page.url.pathname.startsWith('/ops-vault')) return ['Super-Admin', 'Ops Vault'];
		return ['Workspace'];
	});

	$effect(() => {
		void page.url.pathname;
		mobileNav = false;
	});
	// Switching to a lower tier removes modules: leave a page that no longer exists for this tenant.
	$effect(() => {
		const mod = moduleForPath(page.url.pathname);
		if (mod && !ws.manifest.visible_modules.includes(mod)) goto('/dashboard', { replaceState: true });
	});
</script>

<svelte:window
	onkeydown={(e) => {
		if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
			e.preventDefault();
			ws.paletteOpen = !ws.paletteOpen;
		}
	}}
/>

<div class="shell" class:collapsed>
	<aside class="side" class:mobile-open={mobileNav}>
		<div class="side-top">
			<a href="/dashboard" class="brand" aria-label="ZENRES dashboard"><Logo size={22} wordmark={!collapsed} /></a>
			<button class="collapse" onclick={() => (collapsed = !collapsed)} aria-label="Toggle sidebar"><Icon name="layers" size={15} /></button>
		</div>
		<button class="org" onclick={() => (menuOpen = !menuOpen)}>
			<span class="org-mark">{ws.orgName.slice(0, 1)}</span>
			<span class="org-txt"><span class="org-name">{ws.orgName.replace(' Pvt. Ltd.', '')}</span><span class="org-tier mono">{ws.tierLabel}</span></span>
		</button>
		<nav aria-label="Workspace">
			{#each nav as g (g.group)}
				<div class="group">
					<span class="g-label">{g.group}</span>
					{#each g.items as item (item.key)}
						<a href="/{item.key}" class:active={isActive(item.key)} title={collapsed ? item.label : undefined}>
							<Icon name={item.icon} />
							<span class="lbl">{item.label}</span>
							{#if item.key === 'inbox' && ws.unread}<span class="count mono">{ws.unread}</span>{/if}
						</a>
					{/each}
				</div>
			{/each}
		</nav>
		<div class="side-foot mono">
			<span class="dot dot-live text-emerald"></span>
			<span class="lbl">{ws.mode === 'live' ? 'Live · core API' : 'Pilot data · demo mode'}</span>
		</div>
	</aside>
	<div class="scrim" class:on={mobileNav} onclick={() => (mobileNav = false)} role="presentation"></div>

	<div class="main">
		<header class="top">
			<button class="hamb" onclick={() => (mobileNav = true)} aria-label="Open navigation"><Icon name="layers" /></button>
			<div class="crumbs">
				{#each crumbs as c, i (c)}
					{#if i}<Icon name="chevron" size={12} class="sep" />{/if}
					<span class:last={i === crumbs.length - 1}>{c}</span>
				{/each}
			</div>
			<button class="search" onclick={() => (ws.paletteOpen = true)}>
				<Icon name="search" size={15} />
				<span>Search records, SKUs, actions…</span>
				<span class="kbds"><kbd>⌘</kbd><kbd>K</kbd></span>
			</button>
			<div class="top-r">
				<div class="ccy">
					{#each ['USD', 'INR', 'EUR'] as const as c (c)}
						<button class:on={ws.currency === c} onclick={() => ws.setCurrency(c)}>{c}</button>
					{/each}
				</div>
				<button class="bell" class:ring={ws.unread > 0} onclick={() => (ws.bellOpen = true)} aria-label="Notifications, {ws.unread} unread">
					<Icon name="bell" size={18} />
					{#if ws.unread}
						{#key ws.unread}<span class="badge mono">{ws.unread}</span>{/key}
					{/if}
				</button>
				<div class="user">
					<button class="avatar" onclick={() => (menuOpen = !menuOpen)} aria-expanded={menuOpen} aria-label="Account menu">{USER.initials}</button>
					{#if menuOpen}
						<div class="menu card" role="menu">
							<div class="menu-head">
								<strong>{USER.name}</strong>
								<span class="muted">{USER.email}</span>
							</div>
							<div class="menu-sec">
								<span class="eyebrow">Demo organisation</span>
								{#each tiers as t (t.id)}
									<button role="menuitemradio" aria-checked={ws.tier === t.id} onclick={() => { ws.setTier(t.id); menuOpen = false; }}>
										<span class="radio" class:on={ws.tier === t.id}></span>{t.label}
									</button>
								{/each}
							</div>
							<a href="/settings/team" role="menuitem" onclick={() => (menuOpen = false)}><Icon name="settings" size={14} />Workspace settings</a>
							<a href="/logout" role="menuitem" data-sveltekit-reload><Icon name="logout" size={14} />Sign out</a>
						</div>
					{/if}
				</div>
			</div>
		</header>

		<div class="content">
			{@render children()}
		</div>
	</div>
</div>

<MailBellDrawer />
<CommandPalette />
<Toasts />

<svelte:document onclick={(e) => { if (menuOpen && !(e.target as HTMLElement).closest('.user,.org')) menuOpen = false; }} />

<style>
	.shell {
		--side: 248px;
		display: grid;
		grid-template-columns: var(--side) 1fr;
		min-height: 100dvh;
		transition: grid-template-columns 500ms var(--ease-expo);
	}
	.shell.collapsed {
		--side: 68px;
	}
	.side {
		position: sticky;
		top: 0;
		height: 100dvh;
		display: flex;
		flex-direction: column;
		border-right: var(--hairline);
		background: var(--surface-canvas);
		overflow: hidden;
		z-index: 60;
	}
	.side-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		height: 60px;
		padding: 0 14px 0 18px;
	}
	.collapse {
		width: 28px;
		height: 28px;
		display: grid;
		place-items: center;
		border-radius: var(--radius-sm);
		color: var(--ink-muted);
	}
	.collapse:hover {
		background: var(--surface-subtle);
	}
	.collapsed .collapse {
		display: none;
	}
	.org {
		display: flex;
		align-items: center;
		gap: 10px;
		margin: 0 10px 8px;
		padding: 8px;
		border: var(--hairline);
		border-radius: var(--radius-md);
		background: var(--surface-card);
		text-align: left;
		transition: border-color 200ms;
	}
	.org:hover {
		border-color: var(--border-strong);
	}
	.org-mark {
		flex: none;
		width: 30px;
		height: 30px;
		display: grid;
		place-items: center;
		border-radius: var(--radius-sm);
		background: var(--obsidian);
		color: var(--vault-phosphor);
		font-weight: 600;
		font-size: 13px;
	}
	.org-txt {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}
	.org-name {
		font-size: 12.5px;
		font-weight: 600;
		color: var(--ink-primary);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.org-tier {
		font-size: 10px;
		color: var(--brand-emerald);
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}
	nav {
		flex: 1;
		overflow-y: auto;
		padding: 4px 10px 16px;
		scrollbar-width: thin;
	}
	.group {
		display: flex;
		flex-direction: column;
		gap: 1px;
		margin-top: 14px;
	}
	.g-label {
		padding: 0 10px 6px;
		font-family: var(--font-mono);
		font-size: 10px;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--ink-faint);
		white-space: nowrap;
	}
	nav a {
		position: relative;
		display: flex;
		align-items: center;
		gap: 10px;
		height: 34px;
		padding: 0 10px;
		border-radius: var(--radius-sm);
		font-size: 13.5px;
		color: var(--ink-secondary);
		white-space: nowrap;
		transition: background 200ms, color 200ms;
	}
	nav a:hover {
		background: var(--surface-subtle);
		color: var(--ink-primary);
	}
	nav a.active {
		background: var(--surface-card);
		color: var(--ink-primary);
		font-weight: 500;
		box-shadow: 0 0 0 1px var(--border-hairline), 0 1px 2px rgba(14, 20, 18, 0.04);
	}
	nav a.active::before {
		content: '';
		position: absolute;
		left: -10px;
		top: 9px;
		bottom: 9px;
		width: 2px;
		border-radius: 2px;
		background: var(--brand-emerald);
	}
	nav a.active :global(.icon) {
		color: var(--brand-emerald);
	}
	.count {
		margin-left: auto;
		min-width: 18px;
		height: 18px;
		padding: 0 5px;
		border-radius: 9px;
		background: var(--brand-emerald);
		color: #fff;
		font-size: 10px;
		display: grid;
		place-items: center;
	}
	.side-foot {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 14px 20px;
		border-top: var(--hairline);
		font-size: 10.5px;
		color: var(--ink-muted);
		white-space: nowrap;
	}
	.collapsed .lbl,
	.collapsed .g-label,
	.collapsed .org-txt,
	.collapsed .count {
		display: none;
	}
	.collapsed .org {
		justify-content: center;
		padding: 6px;
	}
	.collapsed nav a {
		justify-content: center;
	}
	.collapsed .side-top {
		justify-content: center;
		padding: 0;
	}

	.main {
		min-width: 0;
		display: flex;
		flex-direction: column;
	}
	.top {
		position: sticky;
		top: 0;
		z-index: 50;
		display: flex;
		align-items: center;
		gap: 16px;
		height: 60px;
		padding: 0 24px;
		background: color-mix(in srgb, var(--surface-canvas) 84%, transparent);
		backdrop-filter: blur(12px) saturate(1.3);
		-webkit-backdrop-filter: blur(12px) saturate(1.3);
		border-bottom: var(--hairline);
	}
	.hamb {
		display: none;
	}
	.crumbs {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12.5px;
		color: var(--ink-muted);
		white-space: nowrap;
	}
	.crumbs :global(.sep) {
		color: var(--ink-faint);
	}
	.crumbs .last {
		color: var(--ink-primary);
		font-weight: 500;
	}
	.search {
		margin-left: auto;
		display: flex;
		align-items: center;
		gap: 10px;
		width: min(380px, 34vw);
		height: 36px;
		padding: 0 8px 0 12px;
		border: var(--hairline);
		border-radius: var(--radius-sm);
		background: var(--surface-card);
		color: var(--ink-muted);
		font-size: 12.5px;
		transition: border-color 200ms;
	}
	.search:hover {
		border-color: var(--border-strong);
	}
	.search span:nth-child(2) {
		flex: 1;
		text-align: left;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.kbds {
		display: flex;
		gap: 3px;
	}
	.top-r {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.ccy {
		display: flex;
		padding: 2px;
		border: var(--hairline);
		border-radius: var(--radius-sm);
		background: var(--surface-subtle);
	}
	.ccy button {
		height: 26px;
		padding: 0 8px;
		border-radius: 3px;
		font-family: var(--font-mono);
		font-size: 10.5px;
		color: var(--ink-muted);
		transition: all 200ms;
	}
	.ccy button.on {
		background: var(--surface-card);
		color: var(--ink-primary);
		box-shadow: 0 0 0 1px var(--border-hairline);
	}
	.bell {
		position: relative;
		width: 38px;
		height: 38px;
		display: grid;
		place-items: center;
		border-radius: var(--radius-sm);
		color: var(--ink-secondary);
		transition: background 200ms;
	}
	.bell:hover {
		background: var(--surface-subtle);
		color: var(--ink-primary);
	}
	.bell.ring :global(.icon) {
		animation: ring 4s ease-in-out infinite;
		transform-origin: 50% 2px;
	}
	@keyframes ring {
		0%, 88%, 100% { transform: rotate(0); }
		90% { transform: rotate(14deg); }
		93% { transform: rotate(-12deg); }
		96% { transform: rotate(6deg); }
	}
	.badge {
		position: absolute;
		top: 4px;
		right: 3px;
		min-width: 17px;
		height: 17px;
		padding: 0 4px;
		border-radius: 9px;
		background: var(--status-crimson);
		color: #fff;
		font-size: 10px;
		display: grid;
		place-items: center;
		border: 2px solid var(--surface-canvas);
		animation: pop 500ms var(--ease-expo);
	}
	@keyframes pop {
		from { transform: scale(0.3); }
	}
	.user {
		position: relative;
	}
	.avatar {
		width: 34px;
		height: 34px;
		border-radius: 50%;
		background: var(--brand-emerald);
		color: #fff;
		font-size: 12px;
		font-weight: 600;
	}
	.menu {
		position: absolute;
		right: 0;
		top: calc(100% + 8px);
		width: 260px;
		padding: 6px;
		box-shadow: var(--shadow-pop);
		display: flex;
		flex-direction: column;
		animation: menu 400ms var(--ease-expo);
		z-index: 80;
	}
	@keyframes menu {
		from { opacity: 0; transform: translateY(-6px) scale(0.98); }
	}
	.menu-head {
		display: flex;
		flex-direction: column;
		padding: 10px;
		font-size: 13px;
		color: var(--ink-primary);
	}
	.menu-head .muted {
		font-size: 11.5px;
	}
	.menu-sec {
		display: flex;
		flex-direction: column;
		padding: 8px 4px;
		border-block: var(--hairline);
		margin-bottom: 4px;
	}
	.menu-sec .eyebrow {
		padding: 0 6px 6px;
	}
	.menu button,
	.menu a {
		display: flex;
		align-items: center;
		gap: 10px;
		height: 32px;
		padding: 0 8px;
		border-radius: var(--radius-sm);
		font-size: 12.5px;
		color: var(--ink-secondary);
		text-align: left;
	}
	.menu button:hover,
	.menu a:hover {
		background: var(--surface-subtle);
		color: var(--ink-primary);
	}
	.radio {
		width: 12px;
		height: 12px;
		border-radius: 50%;
		border: 1.5px solid var(--border-strong);
	}
	.radio.on {
		border: 4px solid var(--brand-emerald);
	}
	.content {
		flex: 1;
		padding: 28px 28px 64px;
		max-width: 1560px;
		width: 100%;
		margin: 0 auto;
	}
	.scrim {
		display: none;
	}
	@media (max-width: 1100px) {
		.ccy {
			display: none;
		}
	}
	@media (max-width: 900px) {
		.shell,
		.shell.collapsed {
			grid-template-columns: 1fr;
		}
		.side {
			position: fixed;
			left: 0;
			top: 0;
			width: 270px;
			transform: translateX(-100%);
			transition: transform 500ms var(--ease-expo);
			box-shadow: var(--shadow-pop);
		}
		.side.mobile-open {
			transform: none;
		}
		.scrim {
			display: block;
			position: fixed;
			inset: 0;
			z-index: 55;
			background: rgba(14, 20, 18, 0.2);
			opacity: 0;
			pointer-events: none;
			transition: opacity 300ms;
		}
		.scrim.on {
			opacity: 1;
			pointer-events: auto;
		}
		.hamb {
			display: grid;
			place-items: center;
			width: 34px;
			height: 34px;
		}
		.crumbs {
			display: none;
		}
		.search span:nth-child(2),
		.kbds {
			display: none;
		}
		.search {
			width: 38px;
			margin-left: auto;
			justify-content: center;
			padding: 0;
		}
		.content {
			padding: 20px 16px 48px;
		}
	}
</style>
