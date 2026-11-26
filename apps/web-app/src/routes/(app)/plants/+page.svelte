<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HintWhisper from '$lib/components/HintWhisper.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import Drawer from '$lib/components/ui/Drawer.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { num } from '$lib/format';

	let adding = $state(false);
	let form = $state({ name: '', city: '', un_locode: '', grid: 'IN-WEST', supplier_domain: '' });

	function addPlant() {
		const max = ws.manifest.limits.max_plants;
		if (max !== null && ws.plants.length >= max) {
			// Silent hint instead of a padlock — the form simply explains the next step.
			ws.hint('plants', { kind: 'limit_hit', payload: 'add_plant', count: ws.plants.length });
			return;
		}
		adding = true;
	}

	function save() {
		if (form.supplier_domain && !form.supplier_domain.endsWith('deccanalloy.in')) {
			const h = ws.hint('plants-form', { kind: 'entity_pattern', payload: 'external_domain' });
			if (h) return;
		}
		ws.plants.push({ id: `plt_${Date.now()}`, name: form.name, city: form.city, state: '', un_locode: form.un_locode, grid_region: form.grid, meters: [], dims: { width: 60, depth: 40, height: 12 } });
		adding = false;
		ws.toast(`${form.name} added`);
	}
	const zonesOf = (id: string) => (id === 'plt_chakan' ? ws.zones : []);
</script>

<svelte:head><title>Plants & Zones — ZENRES</title></svelte:head>

<PageHeader eyebrow="ZENRES Business" title="Plants &" serif="factory zones" lede="Manufacturing installations, metering points and the 3D zone layout that powers the digital twin and per-zone emissions.">
	{#snippet actions()}
		<button class="btn btn-sm btn-primary" onclick={addPlant}><Icon name="plus" size={14} />Add plant</button>
	{/snippet}
</PageHeader>

<HintWhisper context="plants" />

<div class="plants">
	{#each ws.plants as p, i (p.id)}
		{@const zs = zonesOf(p.id)}
		<article class="panel plant fade-in" style="--i:{i}">
			<div class="ph">
				<div>
					<span class="eyebrow">{p.un_locode || 'UN/LOCODE pending'} · {p.grid_region}</span>
					<h2>{p.name}</h2>
					<p class="muted">{p.city}{p.state ? `, ${p.state}` : ''}</p>
				</div>
				<a class="btn btn-sm" href="/plants/{p.id}/zones">Zones & shifts <span class="arrow">→</span></a>
			</div>
			<div class="facts">
				<div><span class="eyebrow">Floor</span><span class="num">{p.dims.width} × {p.dims.depth} × {p.dims.height} m</span></div>
				<div><span class="eyebrow">Zones</span><span class="num">{zs.length}</span></div>
				<div><span class="eyebrow">On shift</span><span class="num">{zs.reduce((s, z) => s + z.active_workers, 0)}</span></div>
				<div><span class="eyebrow">Live load</span><span class="num">{num(zs.reduce((s, z) => s + z.live_power_kw, 0))} kW</span></div>
			</div>
			<div class="meters">
				<span class="eyebrow">Meters</span>
				{#each p.meters as m (m)}<span class="pill pill-outline">{m}</span>{:else}<span class="muted small">No meters mapped yet</span>{/each}
			</div>
			{#if zs.length}
				<div class="zones">
					{#each zs as z (z.id)}
						<a href="/dashboard?zone={z.code}" class="zone">
							<span class="mono code">{z.code}</span>
							<span class="zn">{z.name}</span>
							<span class="zbar"><span style="width:{(z.active_workers / z.shift_target_workers) * 100}%"></span></span>
							<span class="mono small">{z.active_workers}/{z.shift_target_workers}</span>
						</a>
					{/each}
				</div>
			{/if}
		</article>
	{/each}
</div>

<Drawer open={adding} title="Add manufacturing plant" eyebrow="Plants" onclose={() => (adding = false)}>
	<div class="form">
		<label class="field"><span class="label">Plant name</span><input class="input" bind:value={form.name} placeholder="Talegaon Rolling Mill" /></label>
		<label class="field"><span class="label">City / industrial area</span><input class="input" bind:value={form.city} /></label>
		<div class="two">
			<label class="field"><span class="label">UN/LOCODE</span><input class="input mono" bind:value={form.un_locode} placeholder="INPNQ" /></label>
			<label class="field"><span class="label">Grid region</span><select class="select" bind:value={form.grid}><option>IN-NORTH</option><option>IN-WEST</option><option>IN-SOUTH</option><option>IN-EAST</option><option>IN-NE</option></select></label>
		</div>
		<label class="field"><span class="label">Operator email domain <span class="muted">(optional)</span></span><input class="input" bind:value={form.supplier_domain} placeholder="deccanalloy.in" /></label>
		<HintWhisper context="plants-form" />
	</div>
	{#snippet footer()}
		<button class="btn btn-sm btn-ghost" onclick={() => (adding = false)}>Cancel</button>
		<button class="btn btn-sm btn-primary" disabled={!form.name} onclick={save}>Save plant</button>
	{/snippet}
</Drawer>

<style>
	.plants {
		display: grid;
		gap: 16px;
	}
	.ph {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 12px;
	}
	h2 {
		font-size: 1.4rem;
		margin-top: 6px;
	}
	.facts {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 1px;
		margin: 18px 0;
		background: var(--border-hairline);
		border: var(--hairline);
		border-radius: var(--radius-sm);
		overflow: hidden;
	}
	.facts > div {
		background: var(--surface-card);
		padding: 12px 14px;
		display: flex;
		flex-direction: column;
		gap: 4px;
		color: var(--ink-primary);
	}
	.meters {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}
	.small {
		font-size: 11.5px;
	}
	.zones {
		margin-top: 16px;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
		gap: 8px;
	}
	.zone {
		display: grid;
		grid-template-columns: auto 1fr;
		gap: 6px 10px;
		align-items: center;
		padding: 12px;
		border: var(--hairline);
		border-radius: var(--radius-sm);
		transition: border-color 200ms, transform 400ms var(--ease-expo);
	}
	.zone:hover {
		border-color: var(--brand-emerald);
		transform: translateY(-2px);
	}
	.code {
		font-size: 10.5px;
		color: var(--brand-emerald);
	}
	.zn {
		font-size: 12.5px;
		color: var(--ink-primary);
	}
	.zbar {
		height: 3px;
		background: var(--surface-sunken);
		border-radius: 2px;
		overflow: hidden;
	}
	.zbar span {
		display: block;
		height: 100%;
		background: var(--brand-emerald);
	}
	.form {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.two {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 10px;
	}
	@media (max-width: 700px) {
		.facts {
			grid-template-columns: 1fr 1fr;
		}
	}
</style>
