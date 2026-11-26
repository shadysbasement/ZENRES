<script lang="ts">
	import { appearance, BACKGROUNDS } from '$lib/stores/appearance.svelte';
</script>

<div class="workspace-backdrop" aria-hidden="true" class:animate={appearance.ambient && appearance.motion}>
	{#each BACKGROUNDS as background (background.id)}
		<div class="scene {background.id}" class:visible={appearance.background === background.id}>
			{#if background.id === 'silk'}
				<svg class="silk-art" viewBox="0 0 1600 920" preserveAspectRatio="xMidYMin slice" fill="none">
					<defs>
						<linearGradient id="silk-face" x1="250" y1="700" x2="1400" y2="170" gradientUnits="userSpaceOnUse"><stop stop-color="#8cc8f5"/><stop offset=".38" stop-color="#b5d7f4"/><stop offset=".68" stop-color="#efedff"/><stop offset="1" stop-color="#bea6ee"/></linearGradient>
						<linearGradient id="silk-fold" x1="300" y1="700" x2="1300" y2="130" gradientUnits="userSpaceOnUse"><stop stop-color="#72afe3"/><stop offset=".5" stop-color="#dbedfc"/><stop offset="1" stop-color="#c8b0ed"/></linearGradient>
						<linearGradient id="silk-edge" x1="50" y1="600" x2="1510" y2="200" gradientUnits="userSpaceOnUse"><stop stop-color="#a0d6fd"/><stop offset=".5" stop-color="#fcfdff"/><stop offset="1" stop-color="#dfc9f8"/></linearGradient>
						<filter id="silk-shadow" x="-25%" y="-40%" width="150%" height="180%"><feGaussianBlur stdDeviation="24"/></filter>
					</defs>
					<path d="M-160 700C160 900 590 150 1010 350S1440 600 1770 100" stroke="#93b4d7" stroke-width="130" opacity=".15" filter="url(#silk-shadow)"/>
					<g class="ribbon">
						<path d="M-160 640C100 870 550 145 940 325S1330 610 1730 110L1740 260C1370 710 1180 680 870 490S120 980-160 800Z" fill="url(#silk-face)"/>
						<path d="M-160 640C100 870 550 145 940 325S1330 610 1730 110C1350 800 1210 575 915 415S90 925-160 640Z" fill="url(#silk-fold)" opacity=".8"/>
						<path d="M-160 640C100 870 550 145 940 325S1330 610 1730 110" stroke="url(#silk-edge)" stroke-width="2"/>
						{#each Array(13) as _, i (i)}
							<path d="M-160 640C100 870 550 145 940 325S1330 610 1730 110" transform="translate(0 {i * 6})" stroke="url(#silk-edge)" stroke-width=".7" opacity={.24 - i * .012}/>
						{/each}
					</g>
				</svg>
			{:else if background.id === 'contour'}
				<svg viewBox="0 0 1600 950" preserveAspectRatio="xMidYMin slice" fill="none">
					{#each Array(26) as _, i (i)}
						<path d="M-200 850C300 850 360 160 920 280S1300 550 1800 140" transform="translate({i * 3} {i * 14 - 140})" stroke="#91a8c7" stroke-width="1" opacity={.36 - i * .009}/>
					{/each}
				</svg>
			{/if}
		</div>
	{/each}
</div>

<style>
	.workspace-backdrop { position: fixed; inset: 0; pointer-events: none; z-index: -1; overflow: hidden; background: #dae3ee; }
	.scene { position: absolute; inset: 0; opacity: 0; transition: opacity 1s var(--ease-out); }
	.scene.visible { opacity: 1; }
	svg { width: 100%; height: 100%; }
	.silk { background: linear-gradient(125deg, #e4eaf2 0%, #dce4ef 45%, #e5e1f0 100%); }
	.silk-art { opacity: .85; }
	.aurora { background: radial-gradient(ellipse at 10% 80%, #b1dcf4 0%, transparent 58%), radial-gradient(ellipse at 90% 20%, #d7c1ee 0%, transparent 55%), radial-gradient(ellipse at 50% 55%, #edf4ff 0%, #dce5f0 80%); background-size: 115% 115%; }
	.rose { background: radial-gradient(ellipse at 80% 20%, #edcede 0%, transparent 60%), radial-gradient(ellipse at 20% 75%, #c6daf3 0%, transparent 65%), #e6e5ef; background-size: 115% 115%; }
	.contour { background: linear-gradient(120deg, #e5edf3, #dce3f0); }
	.mist { background: linear-gradient(120deg, #e5ebf2, #dfe6ef); }
	.animate .visible .ribbon { transform-origin: center; animation: silk-drift 20s ease-in-out infinite alternate; }
	.animate .visible.aurora, .animate .visible.rose { animation: light-drift 22s ease-in-out infinite alternate; }
	@keyframes silk-drift { from { transform: translateY(-12px) rotate(-1deg); } to { transform: translateY(12px) rotate(1deg); } }
	@keyframes light-drift { from { background-position: 0% 0%; } to { background-position: 100% 100%; } }
	@media (prefers-reduced-motion: reduce) { .scene { transition: none; } .animate .scene, .animate .ribbon { animation: none !important; } }
</style>
