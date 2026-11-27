<script lang="ts">
	import { T, useTask, useThrelte } from '@threlte/core';
	import { HTML, RoundedBoxGeometry, Edges } from '@threlte/extras';
	import { CanvasTexture, CatmullRomCurve3, Color, Fog, InstancedMesh, NeutralToneMapping, Object3D, Vector3, type PerspectiveCamera } from 'three';
	import { C, studioEnvironment, qrCells } from './shared';

	interface Props {
		/** 0..1 scroll progress through the pinned section */
		progress: number;
		active: boolean;
		titles: string[];
		onpick?: (i: number) => void;
	}
	let { progress, active, titles, onpick }: Props = $props();

	const { renderer, scene, autoRender, invalidate } = useThrelte();
	renderer.toneMapping = NeutralToneMapping;
	renderer.toneMappingExposure = 0.95;
	scene.environment = studioEnvironment(renderer);
	scene.environmentIntensity = 0.45;
	scene.fog = new Fog('#e3eaf2', 16, 44);

	$effect(() => {
		autoRender.set(active);
		if (active) invalidate();
	});

	/* ---------------- path & stations ---------------- */

	const N = 8;
	const GAP = 6.2;
	const stations = Array.from({ length: N }, (_, i) => new Vector3(i * GAP, 0, Math.sin(i * 0.95) * 2.6));
	const curve = new CatmullRomCurve3(
		[new Vector3(-5, 0, -1.5), ...stations, new Vector3((N - 1) * GAP + 5, 0, Math.sin(N * 0.95) * 2.6)],
		false,
		'catmullrom',
		0.4
	);
	// parameter (0..1 along the curve) of each station
	const stationU = stations.map((p) => {
		let best = 0,
			bestD = Infinity;
		for (let k = 0; k <= 600; k++) {
			const d = curve.getPointAt(k / 600).distanceToSquared(p);
			if (d < bestD) {
				bestD = d;
				best = k / 600;
			}
		}
		return best;
	});

	const DOTS = 260;
	const lastLit = { n: -1 };
	const dummy = new Object3D();
	let dotMesh = $state<InstancedMesh>();
	const cFaint = new Color('#c9d5e6');
	const cOn = new Color(C.blue);
	const cOn2 = new Color(C.lilac);
	const tmpC = new Color();

	$effect(() => {
		const m = dotMesh;
		if (!m) return;
		for (let k = 0; k < DOTS; k++) {
			const p = curve.getPointAt(k / (DOTS - 1));
			dummy.position.set(p.x, 0.03, p.z);
			dummy.updateMatrix();
			m.setMatrixAt(k, dummy.matrix);
			m.setColorAt(k, cFaint);
		}
		m.instanceMatrix.needsUpdate = true;
		if (m.instanceColor) m.instanceColor.needsUpdate = true;
		// The material compiled before instance colours existed — recompile with instancing colour.
		(m.material as { needsUpdate: boolean }).needsUpdate = true;
		lastLit.n = -1;
	});

	const shadowTex = (() => {
		const c = document.createElement('canvas');
		c.width = c.height = 128;
		const g = c.getContext('2d')!;
		const grad = g.createRadialGradient(64, 64, 4, 64, 64, 64);
		grad.addColorStop(0, 'rgba(67, 92, 140, 0.32)');
		grad.addColorStop(1, 'rgba(67, 92, 140, 0)');
		g.fillStyle = grad;
		g.fillRect(0, 0, 128, 128);
		return new CanvasTexture(c);
	})();

	const cells = qrCells(9);
	let qrMesh = $state<InstancedMesh>();
	$effect(() => {
		const m = qrMesh;
		if (!m) return;
		cells.forEach(([x, y], i) => {
			dummy.position.set((x - 4) * 0.13, (4 - y) * 0.13, 0.05);
			dummy.updateMatrix();
			m.setMatrixAt(i, dummy.matrix);
		});
		m.instanceMatrix.needsUpdate = true;
	});

	/* ---------------- animation ---------------- */

	let camera = $state<PerspectiveCamera>();
	let t = $state(0);
	let cam = 0; // smoothed curve parameter for the camera
	let packet = $state(new Vector3());
	let scales = $state(Array(N).fill(0.82));
	let activeIdx = $state(0);
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

	const smooth = (x: number) => x * x * (3 - 2 * x);
	const lookAt = new Vector3();
	const camPos = new Vector3();

	useTask((dt) => {
		t += dt;
		// Dwell at each station: progress → eased station parameter.
		const s = Math.min(N - 1, Math.max(0, progress * (N - 1)));
		const i = Math.min(N - 2, Math.floor(s));
		const f = s - i;
		const eased = f < 0.3 ? 0 : f > 0.7 ? 1 : smooth((f - 0.3) / 0.4);
		const u = stationU[i] + (stationU[i + 1] - stationU[i]) * (s >= N - 1 ? 1 : eased);
		cam += (u - cam) * Math.min(1, dt * 4);
		activeIdx = Math.round(i + eased);

		const target = curve.getPointAt(Math.min(1, Math.max(0, cam)));
		packet = new Vector3(target.x, 0.55 + Math.sin(t * 3) * 0.08, target.z);
		lookAt.set(target.x + 2.2, 0.9, target.z);
		camPos.set(target.x + 5.5 + mx * 1.6, 6.4 - my * 1.2, target.z + 13.5);
		if (camera) {
			camera.position.lerp(camPos, Math.min(1, dt * 3));
			camera.lookAt(lookAt);
		}
		scales = scales.map((sc, k) => sc + ((k === activeIdx ? 1 : 0.8) - sc) * Math.min(1, dt * 5));

		// Light up the travelled path.
		const lit = Math.round(cam * (DOTS - 1));
		if (dotMesh && lit !== lastLit.n) {
			for (let k = 0; k < DOTS; k++) {
				if (k <= lit) tmpC.copy(cOn).lerp(cOn2, k / DOTS);
				else tmpC.copy(cFaint);
				dotMesh.setColorAt(k, tmpC);
			}
			if (dotMesh.instanceColor) dotMesh.instanceColor.needsUpdate = true;
			lastLit.n = lit;
		}
	});

	const bob = (k: number) => (k === activeIdx ? Math.sin(t * 1.6) * 0.08 : 0);
	const bars = [0.8, 1.5, 1.1, 2.1, 1.7];
</script>

<T.PerspectiveCamera makeDefault position={[4.5, 5.2, 10.5]} fov={34} near={0.5} far={120} bind:ref={camera} />

<T.HemisphereLight args={['#ffffff', '#9fb2cf', 0.6]} />
<T.DirectionalLight position={[-6, 14, 10]} intensity={2.4} />
<T.DirectionalLight position={[14, 4, 6]} intensity={0.5} color={C.blueMid} />
<T.DirectionalLight position={[-10, 6, -8]} intensity={0.5} color={C.lilacMid} />

<!-- dotted track -->
<T.InstancedMesh args={[undefined, undefined, DOTS]} bind:ref={dotMesh}>
	<T.CylinderGeometry args={[0.06, 0.06, 0.04, 10]} />
	<T.MeshBasicMaterial toneMapped={false} />
</T.InstancedMesh>

<!-- travelling data packet -->
<T.Group position={[packet.x, packet.y, packet.z]}>
	<T.Mesh>
		<T.SphereGeometry args={[0.2, 24, 24]} />
		<T.MeshStandardMaterial color={C.blueMid} emissive={C.blueMid} emissiveIntensity={1.1} />
	</T.Mesh>
	<T.Mesh scale={1.6 + Math.sin(t * 4) * 0.25}>
		<T.SphereGeometry args={[0.2, 16, 16]} />
		<T.MeshBasicMaterial color={C.lilacMid} transparent opacity={0.22} depthWrite={false} />
	</T.Mesh>
</T.Group>

{#each stations as p, k (k)}
	{@const on = k === activeIdx}
	<T.Group position={[p.x, 0, p.z]}>
		<!-- shadow + pedestal -->
		<T.Mesh position.y={0.005} rotation.x={-Math.PI / 2}>
			<T.PlaneGeometry args={[5.2, 5.2]} />
			<T.MeshBasicMaterial map={shadowTex} transparent depthWrite={false} toneMapped={false} />
		</T.Mesh>
		<T.Mesh position.y={0.17}>
			<T.CylinderGeometry args={[1.55, 1.65, 0.34, 64]} />
			<T.MeshStandardMaterial color={C.clay2} roughness={0.85} />
		</T.Mesh>
		<T.Mesh position.y={0.345} rotation.x={-Math.PI / 2}>
			<T.RingGeometry args={[1.38, 1.5, 64]} />
			<T.MeshBasicMaterial color={on ? C.blueMid : '#d3deec'} toneMapped={false} />
		</T.Mesh>

		<T.Group position.y={0.34 + bob(k)} scale={scales[k]}>
			{#if k === 0}
				<!-- 01 Tally ledger stack -->
				{#each [0, 1, 2, 3, 4] as j (j)}
					<T.Mesh position={[0, 0.14 + j * 0.24, 0]} rotation.y={j * 0.12 - 0.25}>
						<RoundedBoxGeometry args={[1.9, 0.18, 1.35]} radius={0.05} />
						<T.MeshStandardMaterial color={j % 2 ? C.blueSoft : C.clay} roughness={0.6} />
					</T.Mesh>
				{/each}
				{#each [0, 1, 2] as j (j)}
					<T.Mesh position={[-0.15, 1.24, -0.32 + j * 0.3]} rotation.y={0.23}>
						<T.BoxGeometry args={[j === 0 ? 1.2 : 0.9, 0.02, 0.08]} />
						<T.MeshBasicMaterial color={j === 0 ? C.blue : C.lilacSoft} toneMapped={false} />
					</T.Mesh>
				{/each}
			{:else if k === 1}
				<!-- 02 utility meter -->
				<T.Mesh position.y={0.9}>
					<RoundedBoxGeometry args={[1.5, 1.8, 0.9]} radius={0.12} />
					<T.MeshStandardMaterial color={C.clay} roughness={0.7} />
					<Edges color="#b3c2d8" thresholdAngle={25} />
				</T.Mesh>
				<T.Mesh position={[0, 1.15, 0.46]}>
					<T.CircleGeometry args={[0.46, 48]} />
					<T.MeshPhysicalMaterial color={C.glass} roughness={0.05} clearcoat={1} />
				</T.Mesh>
				<T.Mesh position={[0, 1.15, 0.48]} rotation.z={t * (on ? 3 : 0.6)}>
					<T.TorusGeometry args={[0.3, 0.04, 8, 48, Math.PI * 1.4]} />
					<T.MeshStandardMaterial color={C.blueMid} emissive={C.blueMid} emissiveIntensity={0.8} />
				</T.Mesh>
				<T.Mesh position={[0, 0.45, 0.46]}>
					<T.BoxGeometry args={[0.9, 0.18, 0.02]} />
					<T.MeshBasicMaterial color={C.ink} />
				</T.Mesh>
			{:else if k === 2}
				<!-- 03 carbon calculation bars -->
				{#each bars as h, j (j)}
					{@const hh = h * (on ? 1 : 0.55)}
					<T.Mesh position={[-0.9 + j * 0.45, hh / 2, 0]}>
						<RoundedBoxGeometry args={[0.34, hh, 0.34]} radius={0.06} />
						<T.MeshStandardMaterial color={new Color(C.blueSoft).lerp(new Color(C.lilacSoft), j / 4)} roughness={0.45} />
					</T.Mesh>
				{/each}
			{:else if k === 3}
				<!-- 04 product / batch QR -->
				<T.Group position.y={1.3} rotation.y={Math.sin(t * 0.8) * 0.4}>
					<T.Mesh>
						<RoundedBoxGeometry args={[1.5, 1.5, 0.08]} radius={0.06} />
						<T.MeshPhysicalMaterial color="#ffffff" transmission={0.7} roughness={0.12} thickness={0.3} iridescence={1} iridescenceIOR={1.35} />
					</T.Mesh>
					<T.InstancedMesh args={[undefined, undefined, cells.length]} bind:ref={qrMesh}>
						<T.BoxGeometry args={[0.11, 0.11, 0.04]} />
						<T.MeshStandardMaterial color={C.ink} />
					</T.InstancedMesh>
				</T.Group>
				<T.Mesh position.y={0.25}>
					<T.CylinderGeometry args={[0.08, 0.12, 0.5, 12]} />
					<T.MeshStandardMaterial color={C.clay3} />
				</T.Mesh>
			{:else if k === 4}
				<!-- 05 container / shipment -->
				<T.Mesh position.y={0.62} rotation.y={-0.3}>
					<RoundedBoxGeometry args={[2.4, 1.15, 1.15]} radius={0.05} />
					<T.MeshStandardMaterial color={C.blueSoft} roughness={0.55} />
					<Edges color="#7fb0dc" thresholdAngle={25} />
				</T.Mesh>
				<T.Mesh position={[0.18, 0.66, 0.62]} rotation.y={-0.3}>
					<T.PlaneGeometry args={[0.5, 0.5]} />
					<T.MeshBasicMaterial color="#ffffff" />
				</T.Mesh>
				<T.Mesh position.y={1.45 + Math.sin(t * 2) * 0.05} rotation.y={t}>
					<T.OctahedronGeometry args={[0.18]} />
					<T.MeshStandardMaterial color={C.lilac} emissive={C.lilacMid} emissiveIntensity={0.5} />
				</T.Mesh>
			{:else if k === 5}
				<!-- 06 hash-anchored chain -->
				<T.Group position.y={0.9} rotation.y={t * 0.35}>
					{#each [-1, 0, 1] as j (j)}
						<T.Mesh position={[j * 0.85, Math.sin(t * 2 + j) * 0.06, 0]} rotation={[0.4, 0.6 + j * 0.4, 0]}>
							<RoundedBoxGeometry args={[0.55, 0.55, 0.55]} radius={0.08} />
							<T.MeshPhysicalMaterial color={j === 0 ? C.lilacSoft : C.blueSoft} roughness={0.2} clearcoat={1} />
						</T.Mesh>
					{/each}
					{#each [-0.42, 0.42] as x (x)}
						<T.Mesh position={[x, 0, 0]} rotation.y={Math.PI / 2}>
							<T.TorusGeometry args={[0.16, 0.035, 10, 32]} />
							<T.MeshStandardMaterial color={C.blue} emissive={C.blue} emissiveIntensity={on ? 0.6 : 0.1} />
						</T.Mesh>
					{/each}
				</T.Group>
			{:else if k === 6}
				<!-- 07 buyer discovery globe -->
				<T.Group position.y={1.25} rotation.y={t * 0.3}>
					<T.Mesh>
						<T.IcosahedronGeometry args={[0.85, 2]} />
						<T.MeshPhysicalMaterial color={C.glass} transmission={0.6} roughness={0.15} thickness={0.6} />
					</T.Mesh>
					<T.Mesh>
						<T.IcosahedronGeometry args={[0.87, 2]} />
						<T.MeshBasicMaterial color={C.blue} wireframe transparent opacity={0.35} />
					</T.Mesh>
					{#each [0, 1, 2] as j (j)}
						<T.Mesh position={[Math.cos(t * 1.2 + j * 2.1) * 1.25, Math.sin(t * 0.9 + j) * 0.3, Math.sin(t * 1.2 + j * 2.1) * 1.25]}>
							<T.SphereGeometry args={[0.1, 16, 16]} />
							<T.MeshStandardMaterial color={j === 1 ? C.lilac : C.blueMid} emissive={j === 1 ? C.lilacMid : C.blueMid} emissiveIntensity={0.6} />
						</T.Mesh>
					{/each}
				</T.Group>
			{:else}
				<!-- 08 CBAM XML export pack -->
				{#each [0, 1, 2] as j (j)}
					<T.Mesh position={[j * 0.18 - 0.18, 0.95 + j * 0.06, -j * 0.12]} rotation={[-0.12, -0.35 + j * 0.18, 0]}>
						<RoundedBoxGeometry args={[1.35, 1.75, 0.05]} radius={0.04} />
						<T.MeshStandardMaterial color={j === 2 ? '#ffffff' : C.clay2} roughness={0.5} />
					</T.Mesh>
				{/each}
				<T.Mesh position={[0.62, 1.7, 0.3]} rotation.y={-0.2}>
					<T.CylinderGeometry args={[0.24, 0.24, 0.08, 32]} />
					<T.MeshStandardMaterial color={C.blue} emissive={C.blue} emissiveIntensity={on ? 0.5 : 0.1} />
				</T.Mesh>
			{/if}
		</T.Group>

		<HTML position={[0, -0.05, 2.05]} center zIndexRange={[20, 0]}>
			<button class="tag" class:on onclick={() => onpick?.(k)}>
				<span>{String(k + 1).padStart(2, '0')}</span>{titles[k]}
			</button>
		</HTML>
	</T.Group>
{/each}

<style>
	.tag {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 4px 12px 4px 4px;
		border-radius: 999px;
		background: rgba(252, 253, 255, 0.55);
		border: 1px solid rgba(255, 255, 255, 0.85);
		backdrop-filter: blur(8px);
		-webkit-backdrop-filter: blur(8px);
		font-family: 'Urbanist Variable', sans-serif;
		font-size: 12px;
		font-weight: 600;
		color: #687585;
		white-space: nowrap;
		cursor: pointer;
		transition: all 500ms cubic-bezier(0.16, 1, 0.3, 1);
	}
	.tag span {
		display: grid;
		place-items: center;
		width: 22px;
		height: 22px;
		border-radius: 50%;
		background: #e1e7f0;
		font-size: 10px;
	}
	.tag.on {
		color: #2c3137;
		background: rgba(252, 253, 255, 0.9);
		box-shadow: 0 10px 26px -12px rgba(67, 92, 140, 0.5);
		transform: translateY(-2px);
	}
	.tag.on span {
		background: linear-gradient(135deg, #a4d4f7, #cbb5ed);
		color: #32435c;
	}
</style>
