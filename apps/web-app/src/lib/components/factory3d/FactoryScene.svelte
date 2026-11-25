<script lang="ts">
	import { T, useTask, useThrelte } from '@threlte/core';
	import { OrbitControls, HTML, Edges, interactivity, ContactShadows } from '@threlte/extras';
	import { Color, PerspectiveCamera, Vector3 } from 'three';
	import type { OrbitControls as OrbitControlsT } from 'three/examples/jsm/controls/OrbitControls.js';
	import Equipment from './Equipment.svelte';
	import { PAL, EMERALD, TERRACOTTA, type TwinMode } from './palette';
	import { ws } from '$lib/stores/workspace.svelte';

	interface Props {
		mode: TwinMode;
		preset: string;
		overlay: 'workers' | 'skus' | 'emissions' | 'clean';
		onselect: (code: string) => void;
	}
	let { mode, preset, overlay, onselect }: Props = $props();

	interactivity();
	const { scene, invalidate } = useThrelte();
	const pal = $derived(PAL[mode]);
	$effect(() => {
		scene.background = new Color(pal.bg);
		invalidate();
	});

	let camera = $state<PerspectiveCamera>();
	let controls = $state<OrbitControlsT>();
	let hovered = $state<string | null>(null);
	const lift: Record<string, number> = $state({});

	const PRESETS: Record<string, { pos: [number, number, number]; target: [number, number, number] }> = {
		birds: { pos: [46, 42, 50], target: [0, 0, 2] },
		production: { pos: [-4, 26, 24], target: [-6, 2, -8] },
		packaging: { pos: [26, 18, 30], target: [9, 1, 11] },
		offices: { pos: [-36, 16, 28], target: [-23, 4, 11] },
		warehouse: { pos: [-8, 20, 34], target: [-8, 2, 11] },
		cad: { pos: [0.01, 78, 0.01], target: [0, 0, 0] }
	};
	let goal: { pos: Vector3; target: Vector3 } | null = null;
	$effect(() => {
		const p = PRESETS[preset] ?? PRESETS.birds;
		goal = { pos: new Vector3(...p.pos), target: new Vector3(...p.target) };
	});

	useTask((delta) => {
		if (goal && camera && controls) {
			const k = 1 - Math.pow(0.0025, delta);
			camera.position.lerp(goal.pos, k);
			controls.target.lerp(goal.target, k);
			controls.update();
			if (camera.position.distanceTo(goal.pos) < 0.05) goal = null;
		}
		for (const z of ws.zones) {
			const target = hovered === z.code || ws.selectedZone === z.code ? 0.35 : 0;
			const cur = lift[z.code] ?? 0;
			if (Math.abs(cur - target) > 0.001) lift[z.code] = cur + (target - cur) * Math.min(1, delta * 10);
		}
	});

	const batchesIn = (code: string) => ws.batches.filter((b) => b.zone_code === code && b.stage !== 'shipped' && b.stage !== 'scheduled');
	const skusIn = (code: string) => new Set(batchesIn(code).map((b) => b.product_id)).size;
	const tonnesIn = (code: string) => batchesIn(code).reduce((s, b) => s + (b.stage === 'in_production' ? b.produced_qty : b.packaged_qty), 0);
	const tco2hr = (z: (typeof ws.zones)[number]) => z.live_power_kw * 0.000727 + z.fuel_lph * 0.00268;

	const crates = $derived(Math.round(ws.batches.filter((b) => b.zone_code === 'PACK-1').reduce((s, b) => s + b.packaged_qty, 0) * 1.2));
	const stock = $derived(
		ws.inventory
			.filter((i) => i.category === 'raw_material' || i.category === 'finished_good')
			.slice(0, 4)
			.map((i) => ({ pct: Math.min(1, i.qty / Math.max(1, i.reorder * 2.2)), low: i.qty < i.reorder }))
	);

	function cursor(on: boolean) {
		document.body.style.cursor = on ? 'pointer' : '';
	}
</script>

<T.PerspectiveCamera makeDefault position={[46, 42, 50]} fov={32} near={0.5} far={600} bind:ref={camera}>
	<OrbitControls bind:ref={controls} enableDamping dampingFactor={0.08} maxPolarAngle={Math.PI / 2.15} minDistance={18} maxDistance={140} target={[0, 0, 2]} />
</T.PerspectiveCamera>

<T.AmbientLight intensity={mode === 'studio' ? 0.55 : 1.1} />
<T.HemisphereLight args={[mode === 'studio' ? '#bfe3d3' : '#ffffff', mode === 'studio' ? '#0a0e0d' : '#dce3df', mode === 'studio' ? 0.6 : 0.8]} />
<T.DirectionalLight position={[30, 50, 20]} intensity={mode === 'studio' ? 1.6 : 1.4} castShadow shadow.mapSize={[2048, 2048]} shadow.camera.left={-50} shadow.camera.right={50} shadow.camera.top={40} shadow.camera.bottom={-40} />
<T.DirectionalLight position={[-30, 20, -30]} intensity={0.35} color="#9cc4b4" />

<!-- site floor + architectural grid -->
<T.Mesh rotation.x={-Math.PI / 2} position.y={-0.02} receiveShadow>
	<T.PlaneGeometry args={[140, 110]} />
	<T.MeshStandardMaterial color={pal.floor} roughness={1} />
</T.Mesh>
<T.GridHelper args={[140, 70, pal.grid, pal.grid]} position.y={0.001} />
<!-- plant boundary -->
<T.Mesh position={[0, 0.03, 1.5]}>
	<T.BoxGeometry args={[62, 0.05, 37]} />
	<T.MeshStandardMaterial color={pal.slab} transparent opacity={mode === 'studio' ? 0.4 : 0.7} />
	<Edges color={pal.edge} />
</T.Mesh>

<ContactShadows position.y={0.05} opacity={mode === 'studio' ? 0.5 : 0.25} scale={90} blur={2.4} far={14} />

{#each ws.zones as z (z.id)}
	{@const [w, h, d] = z.size}
	{@const hot = hovered === z.code || ws.selectedZone === z.code}
	<T.Group position={[z.pos[0], (lift[z.code] ?? 0), z.pos[2]]}>
		<!-- slab -->
		<T.Mesh position.y={0.15} receiveShadow>
			<T.BoxGeometry args={[w, 0.3, d]} />
			<T.MeshStandardMaterial color={pal.slab} roughness={0.9} />
			<Edges color={hot ? EMERALD : pal.edge} />
		</T.Mesh>
		<!-- zone envelope (hairline volume) -->
		<T.Mesh position.y={h / 2 + 0.3}>
			<T.BoxGeometry args={[w, h, d]} />
			<T.MeshBasicMaterial transparent opacity={hot ? 0.06 : 0} color={EMERALD} depthWrite={false} />
			<Edges color={hot ? EMERALD : pal.edge} threshold={15} />
		</T.Mesh>
		<T.Group position.y={0.3}>
			<Equipment zone={z} {pal} {hot} crates={z.type === 'packaging' ? crates : 0} stock={z.type === 'warehouse' ? stock : []} />
		</T.Group>
		<!-- hit volume -->
		<T.Mesh
			position.y={h / 2}
			onpointerenter={(e: { stopPropagation: () => void }) => {
				e.stopPropagation();
				hovered = z.code;
				cursor(true);
			}}
			onpointerleave={() => {
				if (hovered === z.code) hovered = null;
				cursor(false);
			}}
			onclick={(e: { stopPropagation: () => void }) => {
				e.stopPropagation();
				onselect(z.code);
			}}
		>
			<T.BoxGeometry args={[w, h + 0.6, d]} />
			<T.MeshBasicMaterial transparent opacity={0} depthWrite={false} />
		</T.Mesh>

		{#if overlay !== 'clean'}
			<HTML position={[0, h + 2.4, 0]} center pointerEvents="none" zIndexRange={[40, 0]}>
				<div class="pin" class:hot class:dark={mode === 'studio'}>
					<span class="code">{z.code}</span>
					{#if overlay === 'workers'}
						<span class="p"><b>{z.active_workers}</b> Workers</span>
					{:else if overlay === 'skus'}
						<span class="p">{skusIn(z.code)} SKUs • {tonnesIn(z.code).toFixed(1)}t</span>
					{:else}
						<span class="p" class:warm={tco2hr(z) > 0.3}>{tco2hr(z).toFixed(2)} tCO₂e/h</span>
					{/if}
				</div>
			</HTML>
		{/if}
	</T.Group>
{/each}

<style>
	.pin {
		display: flex;
		align-items: center;
		gap: 0;
		font-family: 'IBM Plex Mono', monospace;
		font-size: 10.5px;
		white-space: nowrap;
		border-radius: 3px;
		overflow: hidden;
		background: #fff;
		border: 1px solid #dce3df;
		color: #0e1412;
		box-shadow: 0 6px 18px -8px rgba(14, 20, 18, 0.4);
		transition: transform 300ms cubic-bezier(0.16, 1, 0.3, 1), border-color 300ms;
		user-select: none;
	}
	.pin.dark {
		background: rgba(10, 14, 13, 0.88);
		border-color: #24332d;
		color: #dfe7e3;
	}
	.pin.hot {
		transform: translateY(-4px) scale(1.06);
		border-color: #38b884;
	}
	.code {
		padding: 3px 6px;
		background: #0e1412;
		color: #38b884;
		letter-spacing: 0.04em;
	}
	.dark .code {
		background: #14533c;
		color: #e8f3ee;
	}
	.p {
		padding: 3px 7px;
	}
	.p b {
		font-weight: 600;
	}
	.warm {
		color: #c05a3c;
	}
</style>
