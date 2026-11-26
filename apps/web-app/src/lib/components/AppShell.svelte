<script lang="ts">
	import { onMount, type Snippet } from 'svelte';
	import { dev } from '$app/environment';
	import { env } from '$env/dynamic/public';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import Icon from './ui/Icon.svelte';
	import MailBellDrawer from './MailBellDrawer.svelte';
	import CommandPalette from './CommandPalette.svelte';
	import Toasts from './ui/Toasts.svelte';
	import WorkspaceBackdrop from './WorkspaceBackdrop.svelte';
	import AppearancePanel from './AppearancePanel.svelte';
	import { appearance } from '$lib/stores/appearance.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { USER } from '$lib/demo/seed';
	import { NAV } from '$lib/nav';
	import { moduleForPath, type SubscriptionTier } from '$lib/entitlements/manifest';

	let { children }: { children: Snippet } = $props();
	let menuOpen = $state(false);
	let openGroup = $state('');
	let appearanceOpen = $state(false);
	const labels: Record<string, string> = { Overview: 'Dashboard', 'ZENRES Business': 'Business', 'ZENRES Carbon': 'Carbon', 'Trace & FreightLink': 'Freight & trace', Connect: 'Connect', Settings: 'Workspace' };
	const nav = $derived(NAV.map((g) => ({ ...g, label: labels[g.group], items: g.items.filter((i) => ws.manifest.visible_modules.includes(i.key)) })).filter((g) => g.items.length));
	const isActive = (key: string) => page.url.pathname === '/' + key || page.url.pathname.startsWith('/' + key + '/');
	const currentPage = $derived(NAV.flatMap((g) => g.items).find((i) => isActive(i.key))?.label ?? 'Workspace');
	const tiers: { id: SubscriptionTier; label: string }[] = [{ id: 'starter', label: 'MSME Starter' }, { id: 'growth', label: 'MSME Growth' }, { id: 'importer_hub', label: 'EU Importer Hub' }];
	const date = new Date().toLocaleDateString('en-GB', { weekday: 'long', day: 'numeric', month: 'long' });
	const legalAvailable = dev || env.PUBLIC_LEGAL_PUBLISHED === 'true';
	const siteURL = env.PUBLIC_SITE_URL ?? 'http://localhost:5173';
	onMount(() => appearance.init());
	$effect(() => { void page.url.pathname; openGroup = ''; menuOpen = false; });
	$effect(() => { const mod = moduleForPath(page.url.pathname); if (mod && !ws.manifest.visible_modules.includes(mod)) goto('/dashboard', { replaceState: true }); });
	function closePopovers(e: MouseEvent) { if (!(e.target as HTMLElement).closest('.nav-group,.account')) { openGroup = ''; menuOpen = false; } }
</script>

<svelte:window onkeydown={(e) => {
	if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') { e.preventDefault(); ws.paletteOpen = !ws.paletteOpen; }
	if (e.key === 'Escape') { openGroup = ''; menuOpen = false; }
}} />
<svelte:document onclick={closePopovers} />

<div class="shell">
	<WorkspaceBackdrop />
	<a class="skip-link" href="#workspace-content">Skip to content</a>
	<header class="masthead">
		<div class="account">
			<button class="account-button glass-pill" aria-label="Account menu" aria-expanded={menuOpen} onclick={() => { menuOpen = !menuOpen; openGroup = ''; }}>
				<span class="avatar">{USER.initials}</span>
				<span class="account-name"><strong>{USER.name}</strong><span>{ws.tierLabel}</span></span>
				<span class="round-icon chevron" class:up={menuOpen}><Icon name="chevron" size={12} /></span>
			</button>
			{#if menuOpen}
				<div class="account-menu popover" transition:fly={{ y: -6, duration: appearance.motion ? 220 : 0, easing: cubicOut }}>
					<div class="menu-head"><strong>{ws.orgName}</strong><span>{USER.email}</span></div>
					<p class="menu-label">Demo workspace</p>
					{#each tiers as tier (tier.id)}
						<button class="menu-row" aria-pressed={ws.tier === tier.id} onclick={() => { ws.setTier(tier.id); menuOpen = false; }}><Icon name="layers" size={16}/><span>{tier.label}</span>{#if ws.tier === tier.id}<Icon name="check" size={14}/>{/if}</button>
					{/each}
					<div class="menu-divider"></div>
                    <p class="menu-label">Display currency</p><div class="currencies">{#each ['USD', 'INR', 'EUR'] as const as currency (currency)}<button aria-pressed={ws.currency === currency} onclick={() => ws.setCurrency(currency)}>{currency}</button>{/each}</div>
					<a class="menu-row" href="/settings/team"><Icon name="settings"/><span>Workspace settings</span></a>
					<a class="menu-row" href="{siteURL}/support"><Icon name="inbox"/><span>Support centre</span></a>
					{#if legalAvailable}<a class="menu-row" href="{siteURL}/account/privacy"><Icon name="shield"/><span>Privacy preferences</span></a>{/if}
					<a class="menu-row" href="/logout" data-sveltekit-reload><Icon name="logout"/><span>Sign out</span></a>
				</div>
			{/if}
		</div>
		<a class="brand" href="/dashboard" aria-label="ZENRES dashboard"><span>ZENRES</span><small>Business, in a better state.</small></a>
		<div class="tools">
			<button class="search glass-pill" onclick={() => (ws.paletteOpen = true)} aria-label="Search workspace"><span class="round-icon"><Icon name="search" size={15}/></span><span class="search-label">Search anything</span><kbd>⌘ K</kbd></button>
			<button class="tool glass-pill" class:unread={ws.unread > 0} onclick={() => (ws.bellOpen = true)} aria-label="Notifications, {ws.unread} unread"><Icon name="bell" size={18}/>{#if ws.unread}<span class="notification-dot"></span>{/if}</button>
			<button class="tool glass-pill" onclick={() => (appearanceOpen = true)} aria-label="Customize appearance" title="Customize appearance"><Icon name="settings" size={18}/></button>
		</div>
	</header>

	<div class="workspace-bar">
		<div class="welcome">
			<span>Hi, {USER.name.split(' ')[0]}!</span>
			<svelte:element this={page.url.pathname === '/dashboard' ? 'h1' : 'p'} class="welcome-title">{page.url.pathname === '/dashboard' ? 'Welcome back' : currentPage}</svelte:element>
		</div>
		<nav class="module-nav" aria-label="Workspace modules">
			{#each nav as group (group.group)}
				<div class="nav-group">
					<button class="module-pill glass-pill" class:active={group.items.some((i) => isActive(i.key))} class:expanded={openGroup === group.group} aria-expanded={openGroup === group.group} aria-controls={'nav-' + group.items[0].key.replaceAll('/', '-')} onclick={() => { openGroup = openGroup === group.group ? '' : group.group; menuOpen = false; }}>
						<span class="round-icon"><Icon name={group.items[0].icon} size={14}/></span><span>{group.label}</span><span class="round-icon chevron" class:up={openGroup === group.group}><Icon name="chevron" size={11}/></span>
					</button>
					{#if openGroup === group.group}
						<div class="nav-popover popover" id={'nav-' + group.items[0].key.replaceAll('/', '-')} transition:fly={{ y: -7, duration: appearance.motion ? 240 : 0, easing: cubicOut }}>
							<p class="menu-label">{group.group}</p>
							{#each group.items as item (item.key)}
								<a class="menu-row" class:selected={isActive(item.key)} href="/{item.key}" aria-current={isActive(item.key) ? 'page' : undefined}><span class="nav-icon"><Icon name={item.icon}/></span><span>{item.label}</span>{#if item.key === 'inbox' && ws.unread}<span class="count">{ws.unread}</span>{:else if isActive(item.key)}<Icon name="check" size={13}/>{/if}</a>
							{/each}
						</div>
					{/if}
				</div>
			{/each}
		</nav>
	</div>
	<main id="workspace-content" class="content" tabindex="-1">{@render children()}</main>
	<footer class="workspace-footer"><span class="status"><i></i>{ws.mode === 'live' ? 'Connected to your workspace' : 'Demo workspace · sample data'}</span><span class="footer-date">{date}</span>{#if legalAvailable}<a href="{siteURL}/legal/terms">Terms</a><a href="{siteURL}/legal/privacy">Privacy</a>{/if}<button onclick={() => (appearanceOpen = true)}><Icon name="sparkle" size={13}/> Make it yours</button></footer>
</div>
<AppearancePanel open={appearanceOpen} onclose={() => (appearanceOpen = false)}/>
<MailBellDrawer/>
<CommandPalette/>
<Toasts/>

<style>
	.shell { position: relative; isolation: isolate; min-height: 100dvh; padding: 0 clamp(20px, 4.2vw, 72px); }
	.masthead { max-width: 1600px; margin: auto; height: 116px; display: grid; grid-template-columns: 1fr auto 1fr; align-items: center; gap: 24px; position: relative; z-index: 90; }
	.glass-pill { display: inline-flex; align-items: center; border: 1.5px solid #ffffffd9; background: #fcfdffb8; border-radius: 100px; box-shadow: 0 2px 3px #52678805, inset 0 1px 1px #fff; backdrop-filter: blur(18px); transition: background 250ms, box-shadow 300ms, transform 350ms var(--ease-out); }
	.glass-pill:hover { background: #fcfdffed; box-shadow: 0 5px 16px #5267880b, inset 0 1px 1px #fff; transform: translateY(-2px); }
	.glass-pill:active { transform: scale(.97); }
	.account { justify-self: start; position: relative; }
	.account-button { padding: 4px 5px 4px 4px; gap: 9px; text-align: left; }
	.avatar { width: 36px; height: 36px; border-radius: 50%; display: grid; place-items: center; font-size: 12px; font-weight: 650; background: linear-gradient(130deg, #bbd9ee, #e3d3f1); color: #4e5b78; border: 2px solid white; }
	.account-name { display: flex; flex-direction: column; padding-right: 8px; }
	.account-name strong { font-size: 12px; font-weight: 650; line-height: 1.3; }
	.account-name > span { font-size: 10px; color: var(--ink-muted); }
	.round-icon { display: grid; place-items: center; width: 27px; height: 27px; border-radius: 50%; flex: none; background: #e1e7edab; border: 1px solid #ffffffb0; }
	.chevron { transform: rotate(90deg); transition: transform 300ms var(--ease-expo); }
	.chevron.up { transform: rotate(-90deg); }
	.brand { display: flex; flex-direction: column; align-items: center; gap: 1px; }
	.brand > span { color: #2c3137; font-size: 27px; font-weight: 650; letter-spacing: -.055em; }
	.brand small { font-size: 11px; color: #697587; }
	.tools { justify-self: end; display: flex; gap: 9px; align-items: center; }
	.tool { height: 39px; width: 39px; justify-content: center; position: relative; }
	.search { gap: 8px; padding: 4px 10px 4px 4px; font-size: 12px; height: 39px; }
	.search kbd { background: transparent; box-shadow: none; border: 0; font-family: var(--font-sans); margin-left: 5px; color: #8691a1; font-size: 10px; }
	.notification-dot { position: absolute; top: 8px; right: 8px; height: 6px; width: 6px; border-radius: 50%; border: 1.5px solid #fff; background: #8f9de4; }
	.workspace-bar { max-width: 1600px; margin: 14px auto 0; min-height: 106px; display: flex; align-items: center; justify-content: space-between; gap: 24px; position: relative; z-index: 80; }
	.welcome { flex-shrink: 0; }
	.welcome > span { font-size: 14px; color: #596574; }
	.welcome-title { font-size: clamp(25px, 2.5vw, 34px); font-weight: 450; line-height: 1.3; letter-spacing: -.04em; margin-top: 3px; }
	.module-nav { display: flex; gap: 8px; flex-wrap: wrap; justify-content: flex-end; }
	.nav-group { position: relative; }
	.module-pill { gap: 7px; padding: 4px; height: 39px; font-size: 12px; font-weight: 500; white-space: nowrap; }
	.module-pill > .chevron { margin-left: 3px; }
	.module-pill.active { background: linear-gradient(105deg, #83c8fa 0%, #aaaef2 55%, #c9a8ee 100%); border-color: #ffffffd6; color: #2c3852; box-shadow: 0 2px 6px #869acb15, inset 0 1px 1px #ffffff80; }
	.module-pill.active .round-icon { background: #ffffffdf; }
	.popover { position: absolute; top: calc(100% + 10px); background: #fcfdfff5; border: 1px solid #fff; box-shadow: 0 14px 50px #485a7c24; border-radius: 19px; padding: 9px; backdrop-filter: blur(30px); z-index: 100; }
	.nav-popover { min-width: 235px; right: 0; }
	.account-menu { width: 275px; left: 0; }
	.menu-label { font-size: 10px; text-transform: uppercase; letter-spacing: .1em; color: var(--ink-muted); padding: 9px 11px; }
	.menu-head { display: flex; flex-direction: column; gap: 4px; padding: 12px 11px; border-bottom: var(--hairline); }
	.menu-head strong { font-size: 14px; font-weight: 600; }
	.menu-head span { font-size: 12px; color: var(--ink-muted); }
	.menu-row { display: flex; width: 100%; align-items: center; gap: 10px; min-height: 41px; padding: 8px 11px; border-radius: 12px; font-size: 13px; text-align: left; transition: background 200ms, transform 250ms var(--ease-out); }
	.menu-row > span:not(.nav-icon) { flex: 1; }
	.menu-row:hover { background: #e9edf6; transform: translateX(2px); }
	.menu-row.selected { background: linear-gradient(100deg, #e1f1fc, #eee5f8); }
	.currencies { display: flex; gap: 6px; padding: 0 8px 10px; } .currencies button { flex: 1; padding: 6px; border-radius: 20px; font-size: 12px; background: #eef2f8; } .currencies button[aria-pressed="true"] { background: #dce8fa; }
	.menu-divider { border-top: var(--hairline); margin: 6px; }
	.nav-icon { display: grid; place-items: center; color: #728298; }
	.menu-row .count { flex: 0 !important; min-width: 20px; text-align: center; background: #e5e1f5; border-radius: 50%; font-size: 10px; }
	.content { max-width: 1600px; margin: 22px auto 0; min-height: calc(100dvh - 332px); outline: none; view-transition-name: workspace-content; }
	.workspace-footer { display: flex; align-items: center; justify-content: space-between; max-width: 1600px; margin: auto; padding: 28px 0 24px; font-size: 11px; color: #647184; gap: 16px; }
	.workspace-footer button, .status { display: flex; gap: 7px; align-items: center; }
	.status i { width: 5px; height: 5px; border-radius: 50%; background: #8096b9; }
	.skip-link { position: fixed; top: -70px; left: 24px; z-index: 1000; padding: 12px 20px; background: white; border-radius: 30px; }
	.skip-link:focus { top: 12px; }
	@media (min-width: 1600px) { .module-pill { font-size: 13px; gap: 10px; } }
	@media (max-width: 1150px) { .workspace-bar { flex-direction: column; align-items: flex-start; gap: 20px; margin-top: 12px; padding-bottom: 10px; } .module-nav { justify-content: flex-start; } .search-label, .search kbd { display: none; } .search { width: 39px; padding: 4px; justify-content: center; } }
	@media (max-width: 650px) { .shell { padding: 0 18px; } .masthead { height: 92px; gap: 8px; } .account-name { display: none; } .account-button { gap: 3px; } .account-button .chevron { width: 20px; height: 20px; } .avatar { width: 30px; height: 30px; } .brand > span { font-size: 23px; } .brand small { display: none; } .tools { gap: 5px; } .tool, .search { height: 34px; width: 34px; } .search .round-icon { background: transparent; border: 0; } .module-nav { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px; width: 100%; } .nav-group { min-width: 0; } .module-pill { width: 100%; gap: 4px; justify-content: space-between; font-size: 11px; } .module-pill .round-icon { width: 23px; height: 23px; } .module-pill > .chevron { width: 18px; height: 18px; margin-left: 0; } .nav-popover { left: 0; right: auto; min-width: 225px; } .nav-group:nth-child(even) .nav-popover { left: auto; right: 0; } .workspace-bar { gap: 17px; } .content { margin-top: 16px; } .footer-date { display: none; } }

    .masthead { animation: shell-arrive 700ms var(--ease-expo) both; }
    .module-pill { animation: shell-arrive 600ms var(--ease-expo) backwards; }
    .nav-group:nth-child(2) .module-pill { animation-delay: 40ms; }
    .nav-group:nth-child(3) .module-pill { animation-delay: 80ms; }
    .nav-group:nth-child(4) .module-pill { animation-delay: 120ms; }
    .nav-group:nth-child(5) .module-pill { animation-delay: 160ms; }
    .nav-group:nth-child(6) .module-pill { animation-delay: 200ms; }
    @keyframes shell-arrive { from { opacity: 0; transform: translateY(-7px); } }
</style>
