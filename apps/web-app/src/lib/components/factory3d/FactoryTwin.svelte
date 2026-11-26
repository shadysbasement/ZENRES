<script lang="ts">
	import { onMount } from 'svelte';
	import type { Component } from 'svelte';
	import FallbackSvgMap from './FallbackSvgMap.svelte';
	import ZoneInspector from './ZoneInspector.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import type { TwinMode } from './palette';

	interface Props {
		initialZone?: string | null;
	}
	let { initialZone = null }: Props = $props();

	let mode = $state<TwinMode>('cad');
	let preset = $state('birds');
	let overlay = $state<'workers' | 'skus' | 'emissions' | 'clean'>('workers');
	let renderer = $state<'loading' | '3d' | '2d'>('loading');
	let lowPower = $state(false);
	let Canvas: Component<Record<string, unknown>> | null = $state(null);
	let Scene: Component<Record<string, unknown>> | null = $state(null);

	const presets = [
		['birds', "Bird's Eye"],
		['production', 'Production'],
		['packaging', 'Packaging'],
		['offices', 'Offices'],
		['warehouse', 'Warehouse'],
		['cad', 'Top-Down CAD']
	];

	function webgl2(): boolean {
		try {
			return !!document.createElement('canvas').getContext('webgl2');
		} catch {
			return false;
		}
	}

	onMount(async () => {
		lowPower = localStorage.getItem('zv-lowpower') === '1';
		if (initialZone) ws.selectedZone = initialZone;
		if (!webgl2() || lowPower) {
			renderer = '2d';
			return;
		}
		const [core, scene] = await Promise.all([import('@threlte/core'), import('./FactoryScene.svelte')]);
		Canvas = core.Canvas as unknown as Component<Record<string, unknown>>;
		Scene = scene.default as unknown as Component<Record<string, unknown>>;
		renderer = '3d';
	});

	function toggleLowPower() {
		lowPower = !lowPower;
		localStorage.setItem('zv-lowpower', lowPower ? '1' : '0');
		if (lowPower) renderer = '2d';
		else if (Canvas) renderer = '3d';
		else location.reload();
	}

	function select(code: string) {
		ws.selectedZone = code;
	}
	$effect(() => {
		if (preset === 'cad') mode = 'cad';
	});
</script>

<section class="twin" class:cad={mode === 'cad'}>
	<div class="tbar">
		<div class="title">
			<span class="dot dot-live"></span>
			<span class="t">Factory Digital Twin</span>
			<span class="mono sub">Chakan Extrusion Plant · {ws.totalWorkers}/{ws.totalTarget} on shift</span>
		</div>
		<div class="presets" role="tablist" aria-label="Camera presets">
			{#each presets as [id, label] (id)}
				<button role="tab" aria-selected={preset === id} onclick={() => (preset = id)}>{label}</button>
			{/each}
		</div>
	</div>

	<div class="stage">
		{#if renderer === '3d' && Canvas && Scene}
			<Canvas shadows>
				<Scene {mode} {preset} {overlay} onselect={select} />
			</Canvas>
		{:else if renderer === '2d'}
			<FallbackSvgMap {overlay} onselect={select} dark={mode === 'studio'} />
		{:else}
			<div class="boot mono"><span class="spin"></span>Compiling procedural geometry…</div>
		{/if}

		<div class="hud">
			<div class="group">
				<span class="eyebrow">Overlays</span>
				{#each [['workers', 'Workers'], ['skus', 'SKUs'], ['emissions', 'Emissions'], ['clean', 'Clean']] as [id, label] (id)}
					<button class:on={overlay === id} onclick={() => (overlay = id as typeof overlay)}>{label}</button>
				{/each}
			</div>
			<div class="group">
				<button class:on={mode === 'studio'} onclick={() => (mode = 'studio')}>Studio</button>
				<button class:on={mode === 'cad'} onclick={() => (mode = 'cad')}>CAD</button>
				<button class:on={lowPower} onclick={toggleLowPower} title="2D fallback for low-power devices">2D</button>
			</div>
		</div>
		<div class="legend mono">
			<span><i class="sw em"></i>Active line</span>
			<span><i class="sw tc"></i>Furnace / low stock</span>
			<span><i class="sw cb"></i>Container ZV-001</span>
			<span class="hint">Drag to orbit · scroll to zoom · click a zone</span>
		</div>
	</div>
</section>

<ZoneInspector />

<style>
	.twin {
		border-radius: var(--radius-lg);
		overflow: hidden;
		background: var(--obsidian);
		border: 1px solid var(--obsidian-line);
		color: var(--ink-inverse);
		transition: background 500ms, border-color 500ms, color 500ms;
	}
	.twin.cad {
		background: var(--surface-card);
		border-color: var(--border-hairline);
		color: var(--ink-primary);
	}
	.tbar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 16px;
		padding: 12px 14px 12px 18px;
		border-bottom: 1px solid var(--obsidian-line);
		flex-wrap: wrap;
	}
	.cad .tbar {
		border-color: var(--border-hairline);
	}
	.title {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.title .dot {
		color: var(--vault-phosphor);
	}
	.t {
		font-weight: 600;
		font-size: 14px;
	}
	.sub {
		font-size: 11px;
		color: var(--vault-dim);
	}
	.cad .sub {
		color: var(--ink-muted);
	}
	.presets {
		display: flex;
		gap: 2px;
		padding: 3px;
		border-radius: var(--radius-sm);
		background: var(--obsidian-2);
		flex-wrap: wrap;
	}
	.cad .presets {
		background: var(--surface-subtle);
	}
	.presets button,
	.group button {
		height: 26px;
		padding: 0 10px;
		border-radius: 3px;
		font-size: 11.5px;
		color: #8fa39a;
		transition: all 200ms;
		white-space: nowrap;
	}
	.presets button:hover,
	.group button:hover {
		color: var(--ink-inverse);
	}
	.cad .presets button:hover,
	.cad .group button:hover {
		color: var(--ink-primary);
	}
	.presets button[aria-selected='true'],
	.group button.on {
		background: var(--brand-emerald);
		color: #fff;
	}
	.stage {
		position: relative;
		height: clamp(420px, 58vh, 620px);
	}
	.boot {
		height: 100%;
		display: grid;
		place-items: center;
		color: var(--vault-dim);
		font-size: 12px;
	}
	.spin {
		display: inline-block;
		width: 12px;
		height: 12px;
		margin-right: 10px;
		border: 1.5px solid var(--obsidian-line);
		border-top-color: var(--vault-phosphor);
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
		vertical-align: -2px;
	}
	@keyframes spin {
		to { transform: rotate(360deg); }
	}
	.hud {
		position: absolute;
		left: 14px;
		top: 14px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		pointer-events: none;
	}
	.group {
		display: flex;
		align-items: center;
		gap: 2px;
		padding: 3px;
		border-radius: var(--radius-sm);
		background: rgba(10, 14, 13, 0.72);
		border: 1px solid var(--obsidian-line);
		backdrop-filter: blur(8px);
		pointer-events: auto;
		width: fit-content;
	}
	.cad .group {
		background: rgba(255, 255, 255, 0.85);
		border-color: var(--border-hairline);
	}
	.group .eyebrow {
		padding: 0 8px;
		color: var(--vault-dim);
		font-size: 9.5px;
	}
	.cad .group button {
		color: var(--ink-muted);
	}
	.legend {
		position: absolute;
		left: 14px;
		right: 14px;
		bottom: 12px;
		display: flex;
		gap: 16px;
		flex-wrap: wrap;
		font-size: 10.5px;
		color: #8fa39a;
		pointer-events: none;
	}
	.cad .legend {
		color: var(--ink-muted);
	}
	.legend .hint {
		margin-left: auto;
	}
	.sw {
		display: inline-block;
		width: 10px;
		height: 3px;
		margin-right: 6px;
		vertical-align: middle;
	}
	.sw.em { background: var(--vault-phosphor); }
	.sw.tc { background: var(--status-terracotta); }
	.sw.cb { background: var(--accent-cobalt-mid); }
	@media (max-width: 700px) {
		.legend .hint { display: none; }
	}
</style>
