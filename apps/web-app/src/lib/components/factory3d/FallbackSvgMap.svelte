<script lang="ts">
	import { ws } from '$lib/stores/workspace.svelte';
	interface Props {
		overlay: string;
		dark?: boolean;
		onselect: (code: string) => void;
	}
	let { overlay, dark = true, onselect }: Props = $props();
	let hovered = $state<string | null>(null);
	// Plant footprint 62 × 37 m → SVG units ×10
	const X = (x: number) => (x + 31) * 10;
	const Z = (z: number) => (z + 17) * 10;
</script>

<svg viewBox="0 0 620 370" class="map" class:dark role="img" aria-label="2D architectural blueprint of the plant">
	<defs>
		<pattern id="bp" width="20" height="20" patternUnits="userSpaceOnUse"><path d="M20 0H0V20" fill="none" stroke="currentColor" stroke-opacity=".08" /></pattern>
		<pattern id="hatch" width="6" height="6" patternUnits="userSpaceOnUse" patternTransform="rotate(45)"><line x1="0" y1="0" x2="0" y2="6" stroke="currentColor" stroke-opacity=".12" /></pattern>
	</defs>
	<rect width="620" height="370" fill="url(#bp)" />
	{#each ws.zones as z (z.id)}
		{@const x = X(z.pos[0] - z.size[0] / 2)}
		{@const y = Z(z.pos[2] - z.size[2] / 2)}
		{@const w = z.size[0] * 10}
		{@const h = z.size[2] * 10}
		<g
			class="zone"
			class:hot={hovered === z.code || ws.selectedZone === z.code}
			role="button"
			tabindex="0"
			aria-label={z.name}
			onpointerenter={() => (hovered = z.code)}
			onpointerleave={() => (hovered = null)}
			onclick={() => onselect(z.code)}
			onkeydown={(e) => e.key === 'Enter' && onselect(z.code)}
		>
			<rect {x} {y} width={w} height={h} rx="2" />
			<rect {x} {y} width={w} height={h} fill="url(#hatch)" class="h" />
			<text x={x + 10} y={y + 20} class="code">{z.code}</text>
			<text x={x + 10} y={y + 36} class="name">{z.name}</text>
			{#if overlay !== 'clean'}
				<text x={x + 10} y={y + h - 12} class="val">
					{overlay === 'workers' ? `${z.active_workers}/${z.shift_target_workers} workers` : overlay === 'skus' ? `${ws.batches.filter((b) => b.zone_code === z.code && b.stage !== 'scheduled').length} batches` : `${z.live_power_kw} kW`}
				</text>
			{/if}
		</g>
	{/each}
</svg>

<style>
	.map {
		width: 100%;
		height: 100%;
		color: var(--ink-primary);
		background: var(--surface-canvas);
		padding: 50px 20px 40px;
	}
	.map.dark {
		color: #dfe7e3;
		background: var(--obsidian);
	}
	.zone {
		cursor: pointer;
		outline: none;
	}
	.zone rect:first-child {
		fill: currentColor;
		fill-opacity: 0.03;
		stroke: currentColor;
		stroke-opacity: 0.35;
		transition: all 250ms;
	}
	.zone.hot rect:first-child,
	.zone:focus-visible rect:first-child {
		stroke: #38b884;
		stroke-opacity: 1;
		stroke-width: 1.5;
		fill: #38b884;
		fill-opacity: 0.08;
	}
	.code {
		font-family: var(--font-mono);
		font-size: 10px;
		fill: #38b884;
		letter-spacing: 0.06em;
	}
	.name {
		font-size: 11px;
		fill: currentColor;
		fill-opacity: 0.8;
	}
	.val {
		font-family: var(--font-mono);
		font-size: 11px;
		fill: currentColor;
	}
</style>
