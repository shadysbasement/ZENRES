<script lang="ts">
	import { page } from '$app/state';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import FallbackSvgMap from '$lib/components/factory3d/FallbackSvgMap.svelte';
	import ZoneInspector from '$lib/components/factory3d/ZoneInspector.svelte';
	import { ws } from '$lib/stores/workspace.svelte';

	const plant = $derived(ws.plants.find((p) => p.id === page.params.id) ?? ws.plants[0]);
	const focus = page.url.searchParams.get('zone');
	const assign = (code: string) => ws.batches.filter((b) => b.zone_code === code && b.stage !== 'scheduled').map((b) => b.code);

	function bump(id: string, field: 'active_workers' | 'shift_target_workers', d: number) {
		const z = ws.zones.find((x) => x.id === id)!;
		ws.updateZone(id, { [field]: Math.max(0, z[field] + d) });
	}
</script>

<svelte:head><title>Zones — {plant.name} — ZENRES</title></svelte:head>

<PageHeader eyebrow="Plants · {plant.name}" title="Zones, shifts &" serif="line assignments" lede="Configure factory zones, their 3D bounding boxes and live shift rosters. Headcount edits broadcast to every open tab." />

<div class="grid2">
	<section class="panel">
		<div class="panel-head"><h2>Zone register</h2><span class="mono muted small">{ws.totalWorkers}/{ws.totalTarget} on shift</span></div>
		<div class="table-wrap">
			<table class="table">
				<thead><tr><th>Zone</th><th>Type</th><th>Supervisor</th><th class="num">Active</th><th class="num">Target</th><th>3D pos · size (m)</th><th>Assigned</th></tr></thead>
				<tbody>
					{#each ws.zones as z (z.id)}
						<tr class:focus={focus === z.code}>
							<td><button class="link-u strong" onclick={() => (ws.selectedZone = z.code)}>{z.code}</button><div class="muted small">{z.name}</div></td>
							<td><span class="pill">{z.type}</span></td>
							<td>{z.supervisor}</td>
							<td class="num"><span class="step"><button onclick={() => bump(z.id, 'active_workers', -1)}>−</button>{z.active_workers}<button onclick={() => bump(z.id, 'active_workers', 1)}>+</button></span></td>
							<td class="num">{z.shift_target_workers}</td>
							<td class="mono small">[{z.pos.join(', ')}] · {z.size.join('×')}</td>
							<td>{#each assign(z.code) as b (b)}<span class="pill pill-outline">{b}</span> {/each}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</section>
	<section class="panel map">
		<div class="panel-head"><h2>Blueprint</h2><span class="mono muted small">click a zone</span></div>
		<div class="mapbox"><FallbackSvgMap overlay="workers" dark={false} onselect={(c) => (ws.selectedZone = c)} /></div>
	</section>
</div>
<ZoneInspector />

<style>
	.grid2 {
		display: grid;
		grid-template-columns: 1.4fr 1fr;
		gap: 16px;
	}
	.small {
		font-size: 11px;
	}
	.strong {
		color: var(--ink-primary);
		font-weight: 600;
	}
	tr.focus {
		background: var(--brand-emerald-tint);
	}
	.step {
		display: inline-flex;
		align-items: center;
		gap: 8px;
	}
	.step button {
		width: 20px;
		height: 20px;
		border: var(--hairline);
		border-radius: 3px;
		color: var(--ink-muted);
	}
	.step button:hover {
		border-color: var(--brand-emerald);
		color: var(--brand-emerald);
	}
	.mapbox {
		border-radius: var(--radius-sm);
		overflow: hidden;
		border: var(--hairline);
		height: 360px;
	}
	@media (max-width: 1100px) {
		.grid2 {
			grid-template-columns: 1fr;
		}
	}
</style>
