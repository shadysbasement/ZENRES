<script lang="ts" module>
	export type HotspotId = 'factory' | 'container' | 'freight' | 'port';
</script>

<script lang="ts">
	import { T, useTask, useThrelte } from '@threlte/core';
	import { OrbitControls, HTML, RoundedBoxGeometry, Edges, interactivity } from '@threlte/extras';
	import { CanvasTexture, CatmullRomCurve3, InstancedMesh, Object3D, QuadraticBezierCurve3, Vector3, Group, NeutralToneMapping } from 'three';
	import type { OrbitControls as OrbitControlsT } from 'three/examples/jsm/controls/OrbitControls.js';
	import { C, studioEnvironment, qrCells, easeOutBack, easeOutExpo, clamp01 } from './shared';

	interface Props {
		play: boolean;
		active: boolean;
		selected: HotspotId | null;
		hovered: HotspotId | null;
		onhover: (id: HotspotId | null) => void;
		onselect: (id: HotspotId) => void;
	}
	let { play, active, selected, hovered, onhover, onselect }: Props = $props();

	interactivity();
	const { renderer, scene, autoRender, invalidate } = useThrelte();
	renderer.toneMapping = NeutralToneMapping;
	renderer.toneMappingExposure = 0.92;
	scene.environment = studioEnvironment(renderer);
	scene.environmentIntensity = 0.42;

	// Pause the render loop entirely while the hero is scrolled out of view.
	$effect(() => {
		autoRender.set(active);
		if (active) invalidate();
	});

	let controls = $state<OrbitControlsT>();
	let island = $state<Group>();
	let autoRotate = $state(true);
	let idleTimer: ReturnType<typeof setTimeout>;
	$effect(() => {
		const c = controls;
		if (!c) return;
		const start = () => {
			autoRotate = false;
			clearTimeout(idleTimer);
		};
		const end = () => {
			clearTimeout(idleTimer);
			idleTimer = setTimeout(() => (autoRotate = true), 3500);
		};
		c.addEventListener('start', start);
		c.addEventListener('end', end);
		return () => {
			c.removeEventListener('start', start);
			c.removeEventListener('end', end);
		};
	});

	// Pointer parallax on the whole island.
	let mx = 0,
		my = 0;
	$effect(() => {
		const move = (e: PointerEvent) => {
			mx = e.clientX / window.innerWidth - 0.5;
			my = e.clientY / window.innerHeight - 0.5;
		};
		window.addEventListener('pointermove', move, { passive: true });
		return () => window.removeEventListener('pointermove', move);
	});

	/* ---------------- geometry data ---------------- */

	const shadowTex = (() => {
		const c = document.createElement('canvas');
		c.width = c.height = 256;
		const g = c.getContext('2d')!;
		const grad = g.createRadialGradient(128, 128, 10, 128, 128, 128);
		grad.addColorStop(0, 'rgba(67, 92, 140, 0.30)');
		grad.addColorStop(0.55, 'rgba(67, 92, 140, 0.12)');
		grad.addColorStop(1, 'rgba(67, 92, 140, 0)');
		g.fillStyle = grad;
		g.fillRect(0, 0, 256, 256);
		return new CanvasTexture(c);
	})();

	const road = new CatmullRomCurve3(
		[
			new Vector3(-6.6, 0.06, 3.2),
			new Vector3(-1.5, 0.06, 3.9),
			new Vector3(3.2, 0.06, 3.4),
			new Vector3(6.2, 0.06, 1.2),
			new Vector3(6.0, 0.06, -1.0),
			new Vector3(3.4, 0.06, -0.6),
			new Vector3(0.8, 0.06, 0.2),
			new Vector3(-1.4, 0.06, 1.4),
			new Vector3(-4.6, 0.06, 1.6)
		],
		true,
		'catmullrom',
		0.35
	);
	const stream = new QuadraticBezierCurve3(new Vector3(-6.0, 3.9, -2.6), new Vector3(-4.2, 6.6, -2.4), new Vector3(-1.6, 5.0, -2.2));
	const cells = qrCells(9);
	const STREAM_N = 22;
	const dummy = new Object3D();
	let streamMesh = $state<InstancedMesh>();
	let qrMesh = $state<InstancedMesh>();

	$effect(() => {
		const m = qrMesh;
		if (!m) return;
		const s = 0.1;
		cells.forEach(([x, y], i) => {
			dummy.position.set((x - 4) * s, (4 - y) * s, 0.04);
			dummy.scale.set(1, 1, 1);
			dummy.rotation.set(0, 0, 0);
			dummy.updateMatrix();
			m.setMatrixAt(i, dummy.matrix);
		});
		m.instanceMatrix.needsUpdate = true;
	});

	/* ---------------- animation ---------------- */

	let t = $state(0);
	let introStart = -1;
	let rise = $state([0, 0, 0, 0, 0, 0]); // plinth, factory, container, road, port, floating
	let truckA = $state({ p: [0, 0, 0] as [number, number, number], r: 0 });
	let truckB = $state({ p: [0, 0, 0] as [number, number, number], r: 0 });
	const tmpP = new Vector3();
	const tmpT = new Vector3();

	function along(u: number) {
		road.getPointAt(((u % 1) + 1) % 1, tmpP);
		road.getTangentAt(((u % 1) + 1) % 1, tmpT);
		return { p: [tmpP.x, 0.32, tmpP.z] as [number, number, number], r: Math.atan2(tmpT.x, tmpT.z) };
	}

	useTask((dt) => {
		t += dt;
		if (play && introStart < 0) introStart = t;
		if (introStart >= 0) {
			const e = t - introStart;
			rise = rise.map((_, i) => clamp01((e - i * 0.12) / 1.1));
		}
		truckA = along(t * 0.035);
		truckB = along(t * 0.035 + 0.5);
		if (island) {
			// gentle idle sway (instead of full auto-rotation) keeps the composition framed
			const sway = autoRotate ? Math.sin(t * 0.22) * 0.28 : 0;
			island.rotation.y += (mx * 0.18 + sway - island.rotation.y) * 0.03;
			island.rotation.x += (my * 0.05 - island.rotation.x) * 0.04;
			island.position.y = Math.sin(t * 0.8) * 0.08;
		}
		if (streamMesh) {
			for (let i = 0; i < STREAM_N; i++) {
				const u = (t * 0.22 + i / STREAM_N) % 1;
				stream.getPoint(u, tmpP);
				const s = 0.06 + Math.sin(u * Math.PI) * 0.07;
				dummy.position.copy(tmpP);
				dummy.scale.setScalar(s / 0.1);
				dummy.updateMatrix();
				streamMesh.setMatrixAt(i, dummy.matrix);
			}
			streamMesh.instanceMatrix.needsUpdate = true;
		}
	});

	const up = (i: number, from = -2.2) => (1 - easeOutBack(rise[i])) * from;
	const isHot = (id: HotspotId) => hovered === id || selected === id;
	const halo = (id: HotspotId) => (isHot(id) ? 0.55 : 0.0);

	const hot = (id: HotspotId) => ({
		onpointerenter: (e: { stopPropagation: () => void }) => {
			e.stopPropagation();
			onhover(id);
			document.body.style.cursor = 'pointer';
		},
		onpointerleave: () => {
			onhover(null);
			document.body.style.cursor = '';
		},
		onclick: (e: { stopPropagation: () => void }) => {
			e.stopPropagation();
			onselect(id);
		}
	});

	let pinsShown = $state(false);
	$effect(() => {
		if (!play) return;
		const id = setTimeout(() => (pinsShown = true), 1900);
		return () => clearTimeout(id);
	});

	const PINS: { id: HotspotId; n: string; label: string; pos: [number, number, number] }[] = [
		{ id: 'factory', n: '01', label: 'Tally & carbon', pos: [-4.3, 4.7, -1.9] },
		{ id: 'container', n: '02', label: 'QR passports', pos: [1.4, 3.7, 1.15] },
		{ id: 'freight', n: '03', label: 'Return freight', pos: [5.9, 1.6, 2.2] },
		{ id: 'port', n: '04', label: 'Global buyers', pos: [5.4, 4.0, -2.9] }
	];
	const sawteeth = [0, 1, 2, 3];
	const ribs = Array.from({ length: 9 }, (_, i) => -1.05 + i * 0.26);
	const stacks: [number, number, number, string][] = [
		[4.3, 0.32, -3.4, C.blueSoft],
		[5.0, 0.32, -3.4, C.clay],
		[5.7, 0.32, -3.4, C.lilacSoft],
		[4.65, 0.92, -3.4, C.clay2],
		[5.35, 0.92, -3.4, C.blueSoft]
	];
	const trees: [number, number, number][] = [
		[-7.0, 0, -3.6],
		[-7.3, 0, -1.2],
		[-1.0, 0, -3.9],
		[2.6, 0, 4.4],
		[-3.6, 0, 4.4],
		[0.2, 0, -4.1]
	];
</script>

<T.PerspectiveCamera makeDefault position={[25, 17, 21]} fov={26} near={0.5} far={200}>
	<OrbitControls
		bind:ref={controls}
		target={[1.3, 0.3, -0.3]}
		enableZoom={false}
		enablePan={false}
		enableDamping
		dampingFactor={0.06}
		rotateSpeed={0.55}
		minPolarAngle={0.72}
		maxPolarAngle={1.22}
	/>
</T.PerspectiveCamera>

<T.HemisphereLight args={['#ffffff', '#9fb2cf', 0.55]} />
<T.DirectionalLight
	position={[-7, 18, 9]}
	intensity={2.8}
	castShadow
	shadow.mapSize={[2048, 2048]}
	shadow.camera.left={-12}
	shadow.camera.right={12}
	shadow.camera.top={12}
	shadow.camera.bottom={-12}
	shadow.bias={-0.0004}
	shadow.radius={6}
/>
<T.DirectionalLight position={[-10, 6, -8]} intensity={0.55} color={C.lilacMid} />
<T.DirectionalLight position={[12, 5, 3]} intensity={0.45} color={C.blueMid} />

<!-- soft floor shadow: radial texture that fades to zero, so no visible plane bounds -->
<T.Mesh position.y={-1.7} rotation.x={-Math.PI / 2} renderOrder={-1}>
	<T.PlaneGeometry args={[30, 22]} />
	<T.MeshBasicMaterial map={shadowTex} transparent depthWrite={false} toneMapped={false} />
</T.Mesh>

<T.Group bind:ref={island}>
	<!-- plinth -->
	<T.Group position.y={up(0, -1)} scale={0.92 + 0.08 * easeOutExpo(rise[0])}>
		<T.Mesh position.y={-0.45} receiveShadow castShadow>
			<RoundedBoxGeometry args={[17, 0.9, 11.6]} radius={0.42} smoothness={6} />
			<T.MeshStandardMaterial color={C.plinth} roughness={0.9} />
			<Edges color="#c3d1e4" thresholdAngle={30} />
		</T.Mesh>
		<T.Mesh position.y={-1.0} receiveShadow>
			<RoundedBoxGeometry args={[16.4, 0.3, 11]} radius={0.15} />
			<T.MeshStandardMaterial color={C.plinthEdge} roughness={1} />
		</T.Mesh>
		<!-- water channel by the port -->
		<T.Mesh position={[5.4, 0.01, -4.75]} receiveShadow>
			<T.BoxGeometry args={[6.6, 0.04, 1.6]} />
			<T.MeshPhysicalMaterial color={C.blueSoft} roughness={0.08} metalness={0} clearcoat={1} clearcoatRoughness={0.05} />
		</T.Mesh>
		<!-- trees -->
		{#each trees as [x, y, z], i (i)}
			<T.Group position={[x, y, z]} scale={0.85 + (i % 3) * 0.18}>
				<T.Mesh position.y={0.78} castShadow>
					<T.ConeGeometry args={[0.42, 1.25, 7]} />
					<T.MeshStandardMaterial color={i % 2 ? '#bcd9ee' : '#d5c6ef'} roughness={0.75} flatShading />
				</T.Mesh>
				<T.Mesh position.y={0.12}>
					<T.CylinderGeometry args={[0.06, 0.08, 0.25, 6]} />
					<T.MeshStandardMaterial color={C.clay3} />
				</T.Mesh>
			</T.Group>
		{/each}
	</T.Group>

	<!-- road -->
	<T.Mesh position.y={up(3, -0.5)} receiveShadow scale.y={0.12}>
		<T.TubeGeometry args={[road, 220, 0.34, 8, true]} />
		<T.MeshStandardMaterial color={C.road} roughness={0.95} />
	</T.Mesh>

	<!-- 01 FACTORY · Tally + carbon -->
	<T.Group position={[-4.3, up(1), -1.9]} scale={isHot('factory') ? 1.03 : 1} {...hot('factory')}>
		<T.Mesh position.y={0.9} castShadow receiveShadow>
			<RoundedBoxGeometry args={[5.2, 1.8, 3.6]} radius={0.08} />
			<T.MeshStandardMaterial color={C.clay} roughness={0.8} />
			<Edges color="#b3c2d8" thresholdAngle={25} />
		</T.Mesh>
		<!-- sawtooth roof with solar + north-light glazing -->
		{#each sawteeth as i (i)}
			<T.Group position={[-1.95 + i * 1.3, 1.8, 0]}>
				<T.Mesh position={[0.18, 0.36, 0]} rotation.z={0.5} castShadow>
					<T.BoxGeometry args={[1.42, 0.07, 3.5]} />
					<T.MeshStandardMaterial color={C.solar} roughness={0.35} metalness={0.4} />
				</T.Mesh>
				<T.Mesh position={[0.79, 0.34, 0]}>
					<T.BoxGeometry args={[0.04, 0.68, 3.5]} />
					<T.MeshPhysicalMaterial color={C.glass} transmission={0.85} roughness={0.12} thickness={0.2} ior={1.4} />
				</T.Mesh>
			</T.Group>
		{/each}
		<!-- glazed office annex -->
		<T.Mesh position={[3.25, 0.7, 0.9]} castShadow>
			<RoundedBoxGeometry args={[1.5, 1.4, 1.8]} radius={0.06} />
			<T.MeshPhysicalMaterial color={C.glass} transmission={0.7} roughness={0.08} thickness={0.6} ior={1.45} iridescence={0.6} iridescenceIOR={1.3} />
		</T.Mesh>
		<!-- chimney with measured-data band -->
		<T.Mesh position={[-2.0, 2.4, -1.15]} castShadow>
			<T.CylinderGeometry args={[0.26, 0.32, 3.2, 24]} />
			<T.MeshStandardMaterial color={C.clay2} roughness={0.7} />
		</T.Mesh>
		<T.Mesh position={[-2.0, 3.65, -1.15]}>
			<T.CylinderGeometry args={[0.275, 0.275, 0.22, 24]} />
			<T.MeshStandardMaterial color={C.lilacMid} emissive={C.lilacMid} emissiveIntensity={0.35 + Math.sin(t * 2) * 0.15} />
		</T.Mesh>
		<!-- loading dock -->
		<T.Mesh position={[1.2, 0.35, 2.05]} castShadow>
			<T.BoxGeometry args={[2.2, 0.7, 0.5]} />
			<T.MeshStandardMaterial color={C.clay2} />
			<Edges color="#b3c2d8" thresholdAngle={25} />
		</T.Mesh>
		<T.Mesh position.y={0.02} rotation.x={-Math.PI / 2} scale={1 + Math.sin(t * 2.4) * 0.03}>
			<T.RingGeometry args={[3.3, 3.42, 96]} />
			<T.MeshBasicMaterial color={C.blueMid} transparent opacity={halo('factory')} depthWrite={false} />
		</T.Mesh>
	</T.Group>

	<!-- measured data stream → Registry XML card -->
	<T.InstancedMesh args={[undefined, undefined, STREAM_N]} bind:ref={streamMesh} visible={rise[5] > 0.5}>
		<T.SphereGeometry args={[0.1, 12, 12]} />
		<T.MeshStandardMaterial color={C.blueMid} emissive={C.blueMid} emissiveIntensity={0.6} />
	</T.InstancedMesh>
	<T.Group position={[-1.2, 5.0 + Math.sin(t * 1.1) * 0.12 + up(5, -1.5), -2.2]} rotation.y={0.65}>
		<T.Mesh castShadow>
			<RoundedBoxGeometry args={[1.9, 1.3, 0.06]} radius={0.03} />
			<T.MeshPhysicalMaterial color="#ffffff" transmission={0.6} roughness={0.25} thickness={0.3} iridescence={0.5} />
		</T.Mesh>
		{#each [0.32, 0.1, -0.12, -0.34] as y, i (i)}
			<T.Mesh position={[-0.12 + (i === 0 ? 0 : 0.12), y, 0.04]}>
				<T.BoxGeometry args={[i === 0 ? 1.3 : 1.05 - i * 0.12, 0.07, 0.01]} />
				<T.MeshBasicMaterial color={i === 0 ? C.blue : C.lilacSoft} />
			</T.Mesh>
		{/each}
	</T.Group>

	<!-- 02 CONTAINER ZV-001 + floating QR passport -->
	<T.Group position={[1.4, up(2), 1.15]} rotation.y={-0.12} scale={isHot('container') ? 1.04 : 1} {...hot('container')}>
		<T.Mesh position.y={0.62} castShadow receiveShadow>
			<RoundedBoxGeometry args={[2.6, 1.1, 1.1]} radius={0.05} />
			<T.MeshStandardMaterial color={C.blueSoft} roughness={0.55} />
			<Edges color="#7fb0dc" thresholdAngle={25} />
		</T.Mesh>
		{#each ribs as x (x)}
			<T.Mesh position={[x, 0.62, 0.56]}>
				<T.BoxGeometry args={[0.05, 0.98, 0.03]} />
				<T.MeshStandardMaterial color="#8fc2ec" />
			</T.Mesh>
		{/each}
		{#each [-0.55, 0.55] as z (z)}
			<T.Mesh position={[-0.9 + (z > 0 ? 0 : 0.4), 0.18, z * 1.3]} castShadow>
				<T.BoxGeometry args={[0.6, 0.36, 0.6]} />
				<T.MeshStandardMaterial color={C.clay3} />
			</T.Mesh>
		{/each}
		<T.Group position={[0, 2.25 + Math.sin(t * 1.6) * 0.14, 0]} rotation.y={t * 0.6}>
			<T.Mesh>
				<RoundedBoxGeometry args={[1.12, 1.12, 0.06]} radius={0.04} />
				<T.MeshPhysicalMaterial color="#ffffff" transmission={0.75} roughness={0.12} thickness={0.25} iridescence={1} iridescenceIOR={1.35} iridescenceThicknessMinimum={200} iridescenceThicknessMaximum={600} />
			</T.Mesh>
			<T.InstancedMesh args={[undefined, undefined, cells.length]} bind:ref={qrMesh}>
				<T.BoxGeometry args={[0.085, 0.085, 0.03]} />
				<T.MeshStandardMaterial color={C.ink} roughness={0.4} />
			</T.InstancedMesh>
		</T.Group>
		<T.Mesh position.y={0.02} rotation.x={-Math.PI / 2}>
			<T.RingGeometry args={[1.85, 1.95, 72]} />
			<T.MeshBasicMaterial color={C.blueMid} transparent opacity={halo('container')} depthWrite={false} />
		</T.Mesh>
	</T.Group>

	<!-- 03 FREIGHTLINK trucks on the backhaul loop -->
	{#each [truckA, truckB] as tk, i (i)}
		<T.Group position={[tk.p[0], tk.p[1] + up(3, -1), tk.p[2]]} rotation.y={tk.r} scale={isHot('freight') ? 1.55 : 1.35} {...hot('freight')}>
			<T.Mesh position={[0, 0.12, 0.62]} castShadow>
				<RoundedBoxGeometry args={[0.56, 0.56, 0.5]} radius={0.08} />
				<T.MeshStandardMaterial color={i ? C.lilac : C.blue} roughness={0.45} />
			</T.Mesh>
			<T.Mesh position={[0, 0.2, 0.86]}>
				<T.BoxGeometry args={[0.46, 0.22, 0.03]} />
				<T.MeshPhysicalMaterial color={C.glass} roughness={0.05} metalness={0.2} />
			</T.Mesh>
			<T.Mesh position={[0, 0.18, -0.2]} castShadow>
				<RoundedBoxGeometry args={[0.6, 0.7, 1.15]} radius={0.05} />
				<T.MeshStandardMaterial color={i ? C.lilacSoft : C.clay} roughness={0.6} />
			</T.Mesh>
		</T.Group>
	{/each}

	<!-- 04 PORT · global buyers -->
	<T.Group position={[0, up(4), 0]} scale={1} {...hot('port')}>
		<T.Group position={[5.4, 0, -2.6]} scale={isHot('port') ? 1.04 : 1}>
			{#each [-1.3, 1.3] as x (x)}
				<T.Mesh position={[x, 1.5, -0.9]} castShadow>
					<T.BoxGeometry args={[0.18, 3.0, 0.18]} />
					<T.MeshStandardMaterial color={C.lilac} roughness={0.5} />
				</T.Mesh>
			{/each}
			<T.Mesh position={[0, 3.05, -1.4]} castShadow>
				<T.BoxGeometry args={[3.0, 0.22, 2.6]} />
				<T.MeshStandardMaterial color={C.lilacSoft} roughness={0.55} />
				<Edges color="#a98fcf" thresholdAngle={25} />
			</T.Mesh>
			<T.Mesh position={[Math.sin(t * 0.7) * 0.9, 2.75, -1.6]} castShadow>
				<T.BoxGeometry args={[0.5, 0.32, 0.5]} />
				<T.MeshStandardMaterial color={C.lilac} />
			</T.Mesh>
			<T.Mesh position={[Math.sin(t * 0.7) * 0.9, 1.9, -1.6]}>
				<T.CylinderGeometry args={[0.015, 0.015, 1.4, 6]} />
				<T.MeshStandardMaterial color={C.ink} />
			</T.Mesh>
			<T.Mesh position.y={0.02} rotation.x={-Math.PI / 2}>
				<T.RingGeometry args={[2.3, 2.4, 72]} />
				<T.MeshBasicMaterial color={C.lilacMid} transparent opacity={halo('port')} depthWrite={false} />
			</T.Mesh>
		</T.Group>
		{#each stacks as [x, y, z, col], i (i)}
			<T.Mesh position={[x, y, z]} castShadow receiveShadow>
				<RoundedBoxGeometry args={[0.66, 0.58, 1.3]} radius={0.04} />
				<T.MeshStandardMaterial color={col} roughness={0.6} />
			</T.Mesh>
		{/each}
		<!-- ship -->
		<T.Group position={[5.2, 0.05 + Math.sin(t * 1.3) * 0.05, -4.75]} rotation.z={Math.sin(t * 0.9) * 0.02}>
			<T.Mesh castShadow>
				<RoundedBoxGeometry args={[4.6, 0.5, 1.15]} radius={0.2} />
				<T.MeshStandardMaterial color={C.blue} roughness={0.5} />
			</T.Mesh>
			{#each [-1.3, -0.6, 0.1, 0.8] as x, i (x)}
				<T.Mesh position={[x, 0.48, 0]} castShadow>
					<T.BoxGeometry args={[0.6, 0.45, 0.9]} />
					<T.MeshStandardMaterial color={[C.clay, C.blueSoft, C.lilacSoft, C.clay2][i]} />
				</T.Mesh>
			{/each}
			<T.Mesh position={[1.65, 0.6, 0]} castShadow>
				<T.BoxGeometry args={[0.5, 0.7, 0.8]} />
				<T.MeshStandardMaterial color={C.clay} />
			</T.Mesh>
		</T.Group>
	</T.Group>

	<!-- hotspot pins -->
	{#each PINS as h (h.id)}
		<HTML position={h.pos} center zIndexRange={[30, 0]}>
			<button
				class="pin"
				class:hot={isHot(h.id)}
				class:show={pinsShown}
				onpointerenter={() => onhover(h.id)}
				onpointerleave={() => onhover(null)}
				onclick={() => onselect(h.id)}
			>
				<span class="n">{h.n}</span><span class="l">{h.label}</span>
			</button>
		</HTML>
	{/each}
</T.Group>

<style>
	.pin {
		display: inline-flex;
		align-items: center;
		gap: 0;
		padding: 3px;
		border-radius: 999px;
		background: rgba(252, 253, 255, 0.72);
		border: 1px solid rgba(255, 255, 255, 0.9);
		box-shadow: 0 8px 24px -10px rgba(67, 92, 140, 0.45);
		backdrop-filter: blur(10px) saturate(1.4);
		-webkit-backdrop-filter: blur(10px) saturate(1.4);
		font-family: 'Urbanist Variable', sans-serif;
		color: #2c3137;
		cursor: pointer;
		white-space: nowrap;
		opacity: 0;
		transform: translateY(8px) scale(0.9);
		transition:
			opacity 700ms cubic-bezier(0.16, 1, 0.3, 1),
			transform 500ms cubic-bezier(0.16, 1, 0.3, 1),
			box-shadow 300ms;
	}
	.pin.show {
		opacity: 1;
		transform: none;
	}
	.n {
		display: grid;
		place-items: center;
		width: 24px;
		height: 24px;
		border-radius: 50%;
		background: linear-gradient(135deg, #a4d4f7, #cbb5ed);
		font-size: 10.5px;
		font-weight: 700;
		color: #32435c;
	}
	.l {
		max-width: 0;
		overflow: hidden;
		font-size: 12px;
		font-weight: 600;
		transition: max-width 500ms cubic-bezier(0.16, 1, 0.3, 1), padding 500ms cubic-bezier(0.16, 1, 0.3, 1);
	}
	.pin:hover .l,
	.pin.hot .l {
		max-width: 140px;
		padding: 0 10px 0 8px;
	}
	.pin.show.hot {
		transform: translateY(-3px);
		box-shadow: 0 12px 30px -10px rgba(130, 101, 168, 0.55);
	}
</style>
