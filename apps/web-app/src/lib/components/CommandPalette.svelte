<script lang="ts">
	import { goto } from '$app/navigation';
	import { tick } from 'svelte';
	import { fly, fade } from 'svelte/transition';
	import { appearance } from '$lib/stores/appearance.svelte';
	import Icon from './ui/Icon.svelte';
	import HintWhisper from './HintWhisper.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { NAV } from '$lib/nav';

	let query = $state('');
	let input: HTMLInputElement | undefined = $state();
	let sel = $state(0);

	type Item = { label: string; sub: string; href: string; icon: string; kind: string };

	const modules = $derived<Item[]>(
		NAV.flatMap((g) => g.items.filter((i) => ws.manifest.visible_modules.includes(i.key)).map((i) => ({ label: i.label, sub: g.group, href: '/' + i.key, icon: i.icon, kind: 'Page' })))
	);
	const records = $derived<Item[]>([
		...ws.products.map((p) => ({ label: `${p.sku} · ${p.name}`, sub: `CN ${p.cn}`, href: `/products/${p.id}`, icon: 'product', kind: 'SKU' })),
		...ws.batches.map((b) => ({ label: `Batch ${b.code}`, sub: b.stage.replace('_', ' '), href: `/trace/qr?batch=${b.code}`, icon: 'qr', kind: 'Batch' })),
		...ws.containers.map((c) => ({ label: `Container ${c.code}`, sub: c.route, href: `/trace/qr?container=${c.code}`, icon: 'box', kind: 'Container' })),
		...ws.zones.map((z) => ({ label: z.name, sub: `${z.code} · ${z.active_workers} workers`, href: `/dashboard?zone=${z.code}`, icon: 'plant', kind: 'Zone' })),
		...ws.notifications.filter((n) => n.category === 'buyer_agency').map((n) => ({ label: n.subject, sub: n.sender_name, href: `/inbox?thread=${n.id}`, icon: 'mail', kind: 'Email' }))
	]);
	const results = $derived.by(() => {
		const q = query.trim().toLowerCase();
		const all = [...modules, ...records];
		if (!q) return modules.slice(0, 8);
		return all.filter((i) => (i.label + ' ' + i.sub).toLowerCase().includes(q)).slice(0, 10);
	});

	// Silent intent matcher on every keystroke.
	$effect(() => {
		const q = query;
		if (q.trim().length >= 3) ws.hint('palette', { kind: 'search_query', payload: q });
		else ws.clearHint('palette');
	});
	$effect(() => {
		if (ws.paletteOpen) {
			query = '';
			sel = 0;
			tick().then(() => input?.focus());
		}
	});
	$effect(() => {
		void results;
		sel = 0;
	});

	function go(i: Item) {
		ws.paletteOpen = false;
		goto(i.href);
	}
	function onKey(e: KeyboardEvent) {
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			sel = Math.min(results.length - 1, sel + 1);
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			sel = Math.max(0, sel - 1);
		} else if (e.key === 'Enter' && results[sel]) go(results[sel]);
		else if (e.key === 'Escape') ws.paletteOpen = false;
	}
</script>

{#if ws.paletteOpen}
	<div class="scrim" transition:fade={{ duration: appearance.motion ? 180 : 0 }} onclick={() => (ws.paletteOpen = false)} role="presentation"></div>
	<div class="pal card" transition:fly={{ y: -12, duration: appearance.motion ? 240 : 0 }} role="dialog" aria-label="Command palette">
		<div class="pfield">
			<Icon name="search" size={17} />
			<input bind:this={input} bind:value={query} onkeydown={onKey} placeholder="Search pages, SKUs, batches, buyer emails…" aria-label="Search" />
			<kbd>esc</kbd>
		</div>
		<div class="hintwrap"><HintWhisper context="palette" /></div>
		<ul class="res" role="listbox">
			{#each results as r, i (r.href + r.label)}
				<li role="option" aria-selected={i === sel}>
					<button class:sel={i === sel} onclick={() => go(r)} onmouseenter={() => (sel = i)}>
						<span class="ic"><Icon name={r.icon} size={15} /></span>
						<span class="t"><span class="l">{r.label}</span><span class="s">{r.sub}</span></span>
						<span class="k mono">{r.kind}</span>
					</button>
				</li>
			{:else}
				<li class="none">No matches for “{query}”.</li>
			{/each}
		</ul>
		<div class="foot mono"><span><kbd>↑</kbd><kbd>↓</kbd> navigate</span><span><kbd>↵</kbd> open</span></div>
	</div>
{/if}

<style>
	.scrim {
		position: fixed;
		inset: 0;
		z-index: 300;
		background: rgba(14, 20, 18, 0.24);
		backdrop-filter: blur(2px);

	}
	@keyframes fade {
		from { opacity: 0; }
	}
	.pal {
		position: fixed;
		left: 50%;
		top: 14vh;
		z-index: 301;
		width: min(620px, calc(100vw - 32px));
		transform: translateX(-50%);
		box-shadow: var(--shadow-pop);
		overflow: hidden;

	}
	@keyframes pal {
		from { opacity: 0; transform: translateX(-50%) translateY(-10px) scale(0.98); }
	}
	.pfield {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 0 16px;
		height: 56px;
		border-bottom: var(--hairline);
		color: var(--ink-muted);
	}
	input {
		flex: 1;
		border: 0;
		outline: none;
		background: transparent;
		font-size: 15px;
		color: var(--ink-primary);
	}
	.hintwrap {
		padding: 0 12px;
	}
	.res {
		list-style: none;
		padding: 6px;
		margin: 0;
		max-height: 380px;
		overflow: auto;
	}
	.res button {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 9px 10px;
		border-radius: var(--radius-sm);
		text-align: left;
	}
	.res button.sel {
		background: var(--surface-subtle);
	}
	.ic {
		width: 30px;
		height: 30px;
		display: grid;
		place-items: center;
		border-radius: var(--radius-sm);
		border: var(--hairline);
		background: var(--surface-card);
		color: var(--ink-secondary);
	}
	.sel .ic {
		color: var(--brand-emerald);
		border-color: #cfe1d7;
	}
	.t {
		flex: 1;
		display: flex;
		flex-direction: column;
		min-width: 0;
	}
	.l {
		font-size: 13.5px;
		color: var(--ink-primary);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.s {
		font-size: 11.5px;
		color: var(--ink-muted);
		text-transform: capitalize;
	}
	.k {
		font-size: 10px;
		color: var(--ink-faint);
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}
	.none {
		padding: 24px;
		text-align: center;
		font-size: 13px;
		color: var(--ink-muted);
	}
	.foot {
		display: flex;
		gap: 16px;
		padding: 10px 16px;
		border-top: var(--hairline);
		font-size: 10.5px;
		color: var(--ink-muted);
	}
	.foot kbd {
		margin-right: 3px;
	}

	.pal { background: #fcfdfffa; border-radius: 24px; border-color: white; backdrop-filter: blur(30px); }
</style>
