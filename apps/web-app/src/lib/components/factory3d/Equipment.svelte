<script lang="ts">
	import { T, useTask } from '@threlte/core';
	import { Edges, InstancedMesh, Instance } from '@threlte/extras';
	import type { Zone } from '$lib/demo/seed';
	import { EMERALD, TERRACOTTA, COBALT, type Pal } from './palette';

	interface Props {
		zone: Zone;
		pal: Pal;
		crates?: number;
		stock?: { pct: number; low: boolean }[];
		hot: boolean;
	}
	let { zone, pal, crates = 0, stock = [], hot }: Props = $props();

	const [w, h, d] = $derived(zone.size);
	let t = $state(0);
	useTask((delta) => {
		t += delta;
	});

	// Production frame layout
	const bays = $derived(Math.max(2, Math.round(w / 6)));
	const cols = $derived(
		Array.from({ length: bays + 1 }, (_, i) => -w / 2 + (i * w) / bays).flatMap((x) => [
			[x, -d / 2],
			[x, d / 2]
		])
	);
	const craneZ = $derived(Math.sin(t * 0.35) * (d / 2 - 2));
	const isA = $derived(zone.code === 'PROD-A');
	const machines = $derived(
		Array.from({ length: isA ? 3 : 6 }, (_, i) => {
			const n = isA ? 3 : 6;
			return { x: -w / 2 + 3 + (i * (w - 6)) / Math.max(1, n - 1), z: isA ? -1 : i % 2 ? 2.5 : -2.5 };
		})
	);
	const rollers = $derived(Array.from({ length: Math.floor((w - 4) / 0.8) }, (_, i) => -w / 2 + 2 + i * 0.8));
	const crateSlots = $derived(Array.from({ length: Math.min(crates, 24) }, (_, i) => ({ x: -w / 2 + 2 + (i % 6) * 1.5, z: d / 2 - 2.2 - Math.floor(i / 6) * 1.5 })));
	const racks = $derived(Array.from({ length: 4 }, (_, i) => -w / 2 + 2 + i * ((w - 4) / 3)));
	const desks = $derived(Array.from({ length: 6 }, (_, i) => ({ x: -w / 2 + 2.5 + (i % 3) * 3.2, z: -d / 2 + 3 + Math.floor(i / 3) * 3.5 })));
</script>

{#if zone.type === 'production'}
	<!-- portal frame columns -->
	<InstancedMesh>
		<T.BoxGeometry args={[0.35, h, 0.35]} />
		<T.MeshStandardMaterial color={pal.body2} roughness={0.7} />
		{#each cols as [x, z], i (i)}
			<Instance position={[x, h / 2, z]} />
		{/each}
	</InstancedMesh>
	<!-- roof trusses -->
	{#each Array.from({ length: bays + 1 }, (_, i) => -w / 2 + (i * w) / bays) as x, i (i)}
		<T.Mesh position={[x, h, 0]}>
			<T.BoxGeometry args={[0.25, 0.5, d]} />
			<T.MeshStandardMaterial color={pal.body2} />
		</T.Mesh>
	{/each}
	<!-- gantry crane -->
	<T.Group position={[0, h - 1.2, craneZ]}>
		<T.Mesh>
			<T.BoxGeometry args={[w - 0.6, 0.45, 0.6]} />
			<T.MeshStandardMaterial color={COBALT} metalness={0.3} roughness={0.5} />
		</T.Mesh>
		<T.Mesh position={[Math.sin(t * 0.6) * (w / 3), -0.6, 0]}>
			<T.BoxGeometry args={[0.8, 0.7, 0.8]} />
			<T.MeshStandardMaterial color={pal.body} />
		</T.Mesh>
	</T.Group>
	<!-- machines -->
	{#each machines as m, i (i)}
		{#if isA}
			<T.Group position={[m.x, 0, m.z]}>
				<T.Mesh position={[0, 1.6, 0]} castShadow>
					<T.CylinderGeometry args={[1.5, 1.8, 3.2, 20]} />
					<T.MeshStandardMaterial color={pal.body} roughness={0.6} />
				</T.Mesh>
				<T.Mesh position={[0, 3.35, 0]}>
					<T.CylinderGeometry args={[1.1, 1.1, 0.3, 20]} />
					<T.MeshStandardMaterial color={TERRACOTTA} emissive={TERRACOTTA} emissiveIntensity={0.6 + Math.sin(t * 2 + i) * 0.25} />
				</T.Mesh>
				<T.Mesh position={[0, 5, 0]}>
					<T.CylinderGeometry args={[0.25, 0.25, 3, 8]} />
					<T.MeshStandardMaterial color={pal.body2} />
				</T.Mesh>
			</T.Group>
		{:else}
			<T.Mesh position={[m.x, 0.9, m.z]} castShadow>
				<T.BoxGeometry args={[2.2, 1.8, 1.6]} />
				<T.MeshStandardMaterial color={pal.body} roughness={0.5} />
				<Edges color={pal.edge} />
			</T.Mesh>
			<T.Mesh position={[m.x, 1.85, m.z]}>
				<T.BoxGeometry args={[0.5, 0.08, 0.5]} />
				<T.MeshStandardMaterial color={EMERALD} emissive={EMERALD} emissiveIntensity={(Math.sin(t * 3 + i) + 1) * 0.6} />
			</T.Mesh>
		{/if}
	{/each}
	<!-- casting line conveyor (Line A) -->
	{#if isA}
		<T.Mesh position={[0, 0.5, 4.5]}>
			<T.BoxGeometry args={[w - 4, 0.3, 1.4]} />
			<T.MeshStandardMaterial color={pal.body2} />
		</T.Mesh>
	{/if}
	<!-- active line status strip -->
	<T.Mesh position={[0, 0.2, d / 2 - 0.4]}>
		<T.BoxGeometry args={[w - 1, 0.08, 0.16]} />
		<T.MeshStandardMaterial color={EMERALD} emissive={EMERALD} emissiveIntensity={hot ? 2.2 : 1.1} />
	</T.Mesh>
	<!-- energy meter node -->
	<T.Mesh position={[w / 2 - 0.8, 1.4, -d / 2 + 0.8]}>
		<T.BoxGeometry args={[0.7, 1.2, 0.4]} />
		<T.MeshStandardMaterial color={pal.body} />
		<Edges color={EMERALD} />
	</T.Mesh>
{:else if zone.type === 'packaging'}
	<!-- conveyor with rollers -->
	<T.Mesh position={[0, 0.7, -2]}>
		<T.BoxGeometry args={[w - 3, 0.15, 1.6]} />
		<T.MeshStandardMaterial color={pal.body2} />
	</T.Mesh>
	<InstancedMesh>
		<T.CylinderGeometry args={[0.09, 0.09, 1.5, 8]} />
		<T.MeshStandardMaterial color={pal.edge} metalness={0.4} />
		{#each rollers as x, i (i)}
			<Instance position={[x, 0.82, -2]} rotation={[Math.PI / 2, 0, 0]} />
		{/each}
	</InstancedMesh>
	<!-- moving parcels on conveyor -->
	{#each [0, 1, 2] as k (k)}
		<T.Mesh position={[(((t * 1.2 + k * ((w - 4) / 3)) % (w - 4)) - (w - 4) / 2), 1.15, -2]}>
			<T.BoxGeometry args={[0.9, 0.6, 0.9]} />
			<T.MeshStandardMaterial color={pal.crate} />
		</T.Mesh>
	{/each}
	<!-- QR labelling stations -->
	{#each [-w / 4, w / 4] as x, i (i)}
		<T.Mesh position={[x, 1.3, -3.6]}>
			<T.BoxGeometry args={[1, 1.4, 0.6]} />
			<T.MeshStandardMaterial color={pal.body} />
			<Edges color={EMERALD} />
		</T.Mesh>
	{/each}
	<!-- palletised crates (instanced count = WIP) -->
	<InstancedMesh>
		<T.BoxGeometry args={[1.1, 0.9, 1.1]} />
		<T.MeshStandardMaterial color={pal.crate} roughness={0.8} />
		{#each crateSlots as c, i (i)}
			<Instance position={[c.x, 0.55, c.z]} />
		{/each}
	</InstancedMesh>
	<!-- container ZV-001 staging pad -->
	<T.Mesh position={[w / 2 - 4.5, 0.05, d / 2 - 3]}>
		<T.BoxGeometry args={[7, 0.1, 3.4]} />
		<T.MeshStandardMaterial color={pal.floor} />
	</T.Mesh>
	<T.Mesh position={[w / 2 - 4.5, 1.35, d / 2 - 3]} castShadow>
		<T.BoxGeometry args={[6.1, 2.6, 2.44]} />
		<T.MeshStandardMaterial color={COBALT} roughness={0.55} metalness={0.2} />
		<Edges color="#3F7396" />
	</T.Mesh>
	<!-- loading dock bays -->
	{#each [-1, 1] as s, i (i)}
		<T.Mesh position={[w / 2 + 0.1, 2, s * 3]}>
			<T.BoxGeometry args={[0.2, 3.6, 3]} />
			<T.MeshStandardMaterial color={pal.body2} />
			<Edges color={pal.edge} />
		</T.Mesh>
	{/each}
{:else if zone.type === 'office'}
	<!-- ground floor desks -->
	{#each desks as k, i (i)}
		<T.Mesh position={[k.x, 0.45, k.z]}>
			<T.BoxGeometry args={[2.2, 0.9, 1.1]} />
			<T.MeshStandardMaterial color={pal.body} />
		</T.Mesh>
	{/each}
	<!-- mezzanine -->
	<T.Mesh position={[0, 4, 0]}>
		<T.BoxGeometry args={[w, 0.25, d]} />
		<T.MeshStandardMaterial color={pal.slab} />
		<Edges color={pal.edge} />
	</T.Mesh>
	<!-- glass partitions -->
	{#each [[0, d / 2, w, 0], [0, -d / 2, w, 0], [w / 2, 0, d, Math.PI / 2], [-w / 2, 0, d, Math.PI / 2]] as [x, z, len, r], i (i)}
		<T.Mesh position={[x, 5.8, z]} rotation={[0, r, 0]}>
			<T.BoxGeometry args={[len, 3.4, 0.06]} />
			<T.MeshPhysicalMaterial color={pal.glass} transparent opacity={0.18} roughness={0.05} metalness={0} />
			<Edges color={pal.glass} />
		</T.Mesh>
	{/each}
	<!-- mezzanine desks: Management, Accounts/Tally, Compliance, QA lab -->
	{#each [[-2.5, -3], [2.5, -3], [-2.5, 3], [2.5, 3]] as [x, z], i (i)}
		<T.Mesh position={[x, 4.6, z]}>
			<T.BoxGeometry args={[2.6, 0.9, 1.4]} />
			<T.MeshStandardMaterial color={i === 1 ? COBALT : pal.body} />
		</T.Mesh>
	{/each}
{:else}
	<!-- warehouse racking aisles with fill-height tied to stock % -->
	{#each racks as x, i (i)}
		{@const s = stock[i] ?? { pct: 0.5, low: false }}
		<T.Group position={[x, 0, 0]}>
			{#each [-d / 2 + 1.5, d / 2 - 1.5] as z, j (j)}
				<T.Mesh position={[0, (h - 1) / 2, z]}>
					<T.BoxGeometry args={[0.12, h - 1, 0.12]} />
					<T.MeshStandardMaterial color={pal.edge} />
				</T.Mesh>
			{/each}
			{#each [1.6, 3.2, 4.8] as y, j (j)}
				<T.Mesh position={[0, y, 0]}>
					<T.BoxGeometry args={[1.4, 0.06, d - 3]} />
					<T.MeshStandardMaterial color={pal.body2} />
				</T.Mesh>
			{/each}
			<T.Mesh position={[0, (Math.max(0.08, s.pct) * (h - 1.5)) / 2, 0]}>
				<T.BoxGeometry args={[1.1, Math.max(0.08, s.pct) * (h - 1.5), d - 3.6]} />
				<T.MeshStandardMaterial color={s.low ? TERRACOTTA : pal.crate} roughness={0.85} transparent opacity={0.92} />
			</T.Mesh>
		</T.Group>
	{/each}
{/if}
