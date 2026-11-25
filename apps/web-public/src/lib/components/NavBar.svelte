<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import Logo from '$design/components/Logo.svelte';
	import { magnetic, scrambleText, scrollToTarget, lockScroll } from '$design/motion';
	import { NAV_LINKS } from '$lib/content';
	import { ui } from '$lib/stores/ui.svelte';

	let scrolled = $state(false);
	let hidden = $state(false);

	onMount(() => {
		let last = window.scrollY;
		const onScroll = () => {
			const y = window.scrollY;
			scrolled = y > 24;
			hidden = y > 320 && y > last && !ui.menuOpen;
			last = y;
		};
		onScroll();
		window.addEventListener('scroll', onScroll, { passive: true });
		return () => window.removeEventListener('scroll', onScroll);
	});

	function handleLink(e: MouseEvent, href: string) {
		ui.menuOpen = false;
		lockScroll(false);
		const [path, hash] = href.split('#');
		if (hash && (path === '' || path === '/') && page.url.pathname === '/') {
			e.preventDefault();
			scrollToTarget(`#${hash}`);
			history.replaceState(history.state, '', `#${hash}`);
		}
	}

	function toggleMenu() {
		ui.menuOpen = !ui.menuOpen;
		lockScroll(ui.menuOpen);
	}
</script>

<header class="nav" class:scrolled class:hidden class:open={ui.menuOpen}>
	<div class="container-wide inner">
		<a href="/" class="brand" aria-label="ZENRES home" onclick={() => { ui.menuOpen = false; lockScroll(false); }}>
			<Logo size={26} inverse={ui.menuOpen} />
		</a>

		<nav class="links" aria-label="Primary">
			{#each NAV_LINKS as link (link.href)}
				<a href={link.href} onclick={(e) => handleLink(e, link.href)} class:active={page.url.pathname === link.href}>
					<span use:scrambleText>{link.label}</span>
				</a>
			{/each}
		</nav>

		<div class="actions">
			<a href="/login" class="signin link-u">Sign In</a>
			<span use:magnetic={0.22}>
				<a href="/signup" class="btn btn-primary btn-sm">Start Free Workspace <span class="arrow">→</span></a>
			</span>
			<button class="burger" aria-label={ui.menuOpen ? 'Close menu' : 'Open menu'} aria-expanded={ui.menuOpen} onclick={toggleMenu}>
				<span></span><span></span>
			</button>
		</div>
	</div>
</header>

<div class="sheet" class:open={ui.menuOpen} aria-hidden={!ui.menuOpen}>
	<div class="container sheet-inner">
		<span class="eyebrow">Menu</span>
		<nav aria-label="Mobile">
			{#each NAV_LINKS as link, i (link.href)}
				<a href={link.href} style="--i:{i}" onclick={(e) => handleLink(e, link.href)} tabindex={ui.menuOpen ? 0 : -1}>
					<span class="mono">0{i + 1}</span>{link.label}
				</a>
			{/each}
		</nav>
		<div class="sheet-actions">
			<a href="/login" class="btn btn-lg" tabindex={ui.menuOpen ? 0 : -1} onclick={() => { ui.menuOpen = false; lockScroll(false); }}>Sign In</a>
			<a href="/signup" class="btn btn-primary btn-lg" tabindex={ui.menuOpen ? 0 : -1} onclick={() => { ui.menuOpen = false; lockScroll(false); }}>Start Free Workspace</a>
		</div>
	</div>
</div>

<style>
	.nav {
		position: fixed;
		inset: 0 0 auto 0;
		z-index: 100;
		height: 68px;
		transition:
			transform 600ms var(--ease-expo),
			background 300ms var(--ease-out),
			border-color 300ms var(--ease-out);
		border-bottom: 1px solid transparent;
	}
	.nav.scrolled {
		background: color-mix(in srgb, var(--surface-canvas) 82%, transparent);
		backdrop-filter: saturate(1.4) blur(14px);
		-webkit-backdrop-filter: saturate(1.4) blur(14px);
		border-bottom-color: var(--border-hairline);
	}
	.nav.hidden {
		transform: translateY(-100%);
	}
	.nav.open {
		background: transparent;
		border-bottom-color: transparent;
		backdrop-filter: none;
	}
	.inner {
		height: 100%;
		display: grid;
		grid-template-columns: 1fr auto 1fr;
		align-items: center;
		gap: 24px;
	}
	.links {
		display: flex;
		gap: 4px;
	}
	.links a {
		position: relative;
		padding: 8px 12px;
		font-size: var(--fs-sm);
		font-weight: 500;
		color: var(--ink-secondary);
		border-radius: var(--radius-sm);
		transition: color var(--dur-2) var(--ease-out), background var(--dur-2) var(--ease-out);
	}
	.links a:hover,
	.links a.active {
		color: var(--ink-primary);
		background: var(--surface-subtle);
	}
	.actions {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 18px;
	}
	.signin {
		font-size: var(--fs-sm);
		font-weight: 500;
		color: var(--ink-primary);
	}
	.burger {
		display: none;
		width: 40px;
		height: 40px;
		position: relative;
		border-radius: var(--radius-sm);
	}
	.burger span {
		position: absolute;
		left: 11px;
		right: 11px;
		height: 1.5px;
		background: currentColor;
		transition: transform 500ms var(--ease-expo), top 500ms var(--ease-expo);
	}
	.burger span:first-child {
		top: 16px;
	}
	.burger span:last-child {
		top: 23px;
	}
	.open .burger {
		color: var(--ink-inverse);
	}
	.open .burger span:first-child {
		top: 19.5px;
		transform: rotate(45deg);
	}
	.open .burger span:last-child {
		top: 19.5px;
		transform: rotate(-45deg);
	}

	.sheet {
		position: fixed;
		inset: 0;
		z-index: 90;
		background: var(--obsidian);
		color: var(--ink-inverse);
		clip-path: inset(0 0 100% 0);
		transition: clip-path 900ms var(--ease-expo);
		pointer-events: none;
	}
	.sheet.open {
		clip-path: inset(0 0 0 0);
		pointer-events: auto;
	}
	.sheet-inner {
		height: 100%;
		display: flex;
		flex-direction: column;
		justify-content: center;
		gap: 28px;
		padding-top: 80px;
	}
	.sheet .eyebrow {
		color: var(--vault-dim);
	}
	.sheet nav {
		display: flex;
		flex-direction: column;
	}
	.sheet nav a {
		display: flex;
		align-items: baseline;
		gap: 16px;
		padding: 10px 0;
		font-size: clamp(2.25rem, 9vw, 4rem);
		font-weight: 500;
		letter-spacing: -0.03em;
		line-height: 1.05;
		border-bottom: 1px solid var(--obsidian-line);
		opacity: 0;
		transform: translateY(30px);
		transition: opacity 700ms var(--ease-expo), transform 900ms var(--ease-expo);
		transition-delay: calc(var(--i) * 60ms + 200ms);
	}
	.sheet nav a .mono {
		font-size: 12px;
		color: var(--vault-phosphor);
		letter-spacing: 0.06em;
	}
	.sheet.open nav a {
		opacity: 1;
		transform: none;
	}
	.sheet-actions {
		display: flex;
		gap: 12px;
		flex-wrap: wrap;
	}
	.sheet-actions .btn:not(.btn-primary) {
		--btn-bg: transparent;
		--btn-fg: var(--ink-inverse);
		--btn-bd: var(--obsidian-line);
	}

	@media (max-width: 980px) {
		.links,
		.signin,
		.actions > span {
			display: none;
		}
		.burger {
			display: block;
		}
		.inner {
			grid-template-columns: 1fr auto;
		}
	}
	@media (min-width: 981px) {
		.sheet {
			display: none;
		}
	}
</style>
