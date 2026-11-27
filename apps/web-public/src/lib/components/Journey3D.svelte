<script lang="ts">
	import { onMount } from 'svelte';
	import type { Component } from 'svelte';
	import { JOURNEY } from '$lib/content';
	import { webglAvailable } from '$lib/three/shared';
	import { prefersReducedMotion, reveal, scrollToTarget } from '$design/motion';
	import JourneyStrip from './JourneyStrip.svelte';

	let section = $state<HTMLElement>() as HTMLElement;
	let mode = $state<'pending' | '3d' | 'flat'>('pending');
	let Canvas: Component<Record<string, unknown>> | null = $state(null);
	let Scene: Component<Record<string, unknown>> | null = $state(null);
	let progress = $state(0);
	let active = $state(false);
	const step = $derived(Math.min(JOURNEY.length - 1, Math.round(progress * (JOURNEY.length - 1))));
	const s = $derived(JOURNEY[step]);
	const titles = JOURNEY.map((j) => j.title);

	onMount(() => {
		const wide = window.matchMedia('(min-width: 900px)').matches;
		if (!wide || prefersReducedMotion() || !webglAvailable()) {
			mode = 'flat';
			return;
		}
		mode = '3d';
		const io = new IntersectionObserver(([e]) => (active = e.isIntersecting), { rootMargin: '200px' });
		io.observe(section);
		// Load the 3D bundle only when the section approaches the viewport.
		const loader = new IntersectionObserver(
			([e]) => {
				if (!e.isIntersecting) return;
				loader.disconnect();
				Promise.all([import('@threlte/core'), import('$lib/three/JourneyScene.svelte')]).then(([core, sc]) => {
					Canvas = core.Canvas as unknown as Component<Record<string, unknown>>;
					Scene = sc.default as unknown as Component<Record<string, unknown>>;
				});
			},
			{ rootMargin: '900px' }
		);
		loader.observe(section);
		let raf = 0;
		const update = () => {
			raf = 0;
			const r = section.getBoundingClientRect();
			const total = r.height - window.innerHeight;
			progress = Math.min(1, Math.max(0, -r.top / Math.max(1, total)));
		};
		const onScroll = () => {
			if (!raf) raf = requestAnimationFrame(update);
		};
		update();
		window.addEventListener('scroll', onScroll, { passive: true });
		window.addEventListener('resize', onScroll);
		return () => {
			io.disconnect();
			loader.disconnect();
			cancelAnimationFrame(raf);
			window.removeEventListener('scroll', onScroll);
			window.removeEventListener('resize', onScroll);
		};
	});

	function goTo(i: number) {
		const top = section.getBoundingClientRect().top + window.scrollY;
		const total = section.offsetHeight - window.innerHeight;
		const y = top + (i / (JOURNEY.length - 1)) * total + 2;
		const lenis = (window as unknown as { __lenis?: { scrollTo: (y: number, o: unknown) => void } }).__lenis;
		if (lenis) lenis.scrollTo(y, { duration: 1.2 });
		else window.scrollTo({ top: y, behavior: 'smooth' });
	}
</script>

{#if mode === 'flat'}
	<JourneyStrip />
{:else}
	<section class="j3d" bind:this={section} aria-labelledby="journey3d-title" style="--steps:{JOURNEY.length}">
		<div class="pin">
			<div class="scene" aria-hidden="true">
				{#if Canvas && Scene}
					<Canvas dpr={[1, 1.6]}>
						<Scene {progress} {active} {titles} onpick={goTo} />
					</Canvas>
				{/if}
			</div>
			<div class="veil" aria-hidden="true"></div>

			<header class="container-wide head">
				<div>
					<span class="eyebrow" use:reveal><span class="idx">(03)</span> — End-to-end customer journey</span>
					<h2 id="journey3d-title" use:reveal={{ delay: 80 }}>From a Tally voucher <span class="accent">to a CBAM submission.</span></h2>
				</div>
				<div class="counter" aria-hidden="true">
					<span class="cur">{String(step + 1).padStart(2, '0')}</span><span class="of">/ 08</span>
				</div>
			</header>

			<div class="container-wide foot">
				{#key step}
					<article class="card" aria-live="polite">
						<div class="card-top">
							<span class="num">{s.n}</span>
							<span class="pill">{s.tag}</span>
						</div>
						<h3>{s.title}</h3>
						<p>{s.body}</p>
					</article>
				{/key}

				<nav class="rail" aria-label="Journey steps">
					<div class="track"><span style="transform:scaleX({progress})"></span></div>
					<ol role="list">
						{#each JOURNEY as j, i (j.n)}
							<li>
								<button class:done={i < step} class:on={i === step} onclick={() => goTo(i)} aria-label="Step {j.n}: {j.title}">
									<i></i><span>{j.title}</span>
								</button>
							</li>
						{/each}
					</ol>
				</nav>
			</div>
		</div>
	</section>
{/if}

<style>
	.j3d {
		position: relative;
		height: calc(100vh + (var(--steps) - 1) * 70vh);
	}
	.pin {
		position: sticky;
		top: 0;
		height: 100vh;
		overflow: hidden;
	}
	.scene {
		position: absolute;
		inset: 0;
	}
	.veil {
		position: absolute;
		inset: 0;
		pointer-events: none;
		background:
			linear-gradient(180deg, var(--surface-canvas) 0%, transparent 26%, transparent 70%, var(--surface-canvas) 100%),
			linear-gradient(90deg, rgba(227, 234, 242, 0.85) 0%, transparent 32%);
	}
	.head {
		position: relative;
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 24px;
		padding-top: clamp(84px, 11vh, 120px);
		pointer-events: none;
	}
	.head .idx {
		color: var(--brand-emerald);
	}
	h2 {
		margin-top: 14px;
		font-size: clamp(2.2rem, 3.8vw, 3.8rem);
		font-weight: 600;
		letter-spacing: -0.04em;
		line-height: 1;
		max-width: 15ch;
	}
	.accent {
		background: linear-gradient(100deg, #436c9b 0%, #6a8fd6 45%, #8265a8 100%);
		-webkit-background-clip: text;
		background-clip: text;
		color: transparent;
	}
	.counter {
		display: flex;
		align-items: baseline;
		gap: 6px;
		font-variant-numeric: tabular-nums;
	}
	.cur {
		font-size: clamp(3rem, 6vw, 5.5rem);
		font-weight: 600;
		letter-spacing: -0.05em;
		line-height: 1;
		color: var(--ink-primary);
	}
	.of {
		font-size: 14px;
		color: var(--ink-muted);
	}
	.foot {
		position: absolute;
		left: 0;
		right: 0;
		bottom: clamp(24px, 5vh, 48px);
		margin: 0 auto;
		display: grid;
		grid-template-columns: minmax(0, 380px) 1fr;
		gap: 32px;
		align-items: end;
	}
	.card {
		padding: 22px 22px 20px;
		border-radius: var(--radius-lg);
		background: rgba(252, 253, 255, 0.7);
		border: 1px solid rgba(255, 255, 255, 0.95);
		box-shadow: 0 30px 70px -36px rgba(67, 92, 140, 0.6);
		backdrop-filter: blur(18px) saturate(1.5);
		-webkit-backdrop-filter: blur(18px) saturate(1.5);
		display: flex;
		flex-direction: column;
		gap: 10px;
		animation: card-in 700ms var(--ease-expo);
	}
	@keyframes card-in {
		from {
			opacity: 0;
			transform: translateY(14px);
		}
	}
	.card-top {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	.card .num {
		display: grid;
		place-items: center;
		width: 38px;
		height: 38px;
		border-radius: 50%;
		background: linear-gradient(135deg, #a4d4f7, #cbb5ed);
		font-weight: 700;
		font-size: 13px;
		color: #32435c;
	}
	h3 {
		font-size: 1.45rem;
		letter-spacing: -0.025em;
	}
	.card p {
		font-size: 14px;
		line-height: 1.6;
	}
	.rail {
		position: relative;
		padding: 14px 18px 12px;
		border-radius: 999px;
		background: rgba(252, 253, 255, 0.55);
		border: 1px solid rgba(255, 255, 255, 0.9);
		backdrop-filter: blur(14px);
		-webkit-backdrop-filter: blur(14px);
	}
	.track {
		position: absolute;
		left: 34px;
		right: 34px;
		top: 50%;
		height: 2px;
		margin-top: -1px;
		background: rgba(189, 203, 220, 0.6);
		border-radius: 2px;
		overflow: hidden;
	}
	.track span {
		display: block;
		height: 100%;
		background: linear-gradient(90deg, #6ab8ff, #cfa3f6);
		transform-origin: left;
	}
	.rail ol {
		position: relative;
		display: flex;
		justify-content: space-between;
		list-style: none;
		margin: 0;
		padding: 0;
	}
	.rail button {
		position: relative;
		display: grid;
		place-items: center;
		width: 32px;
		height: 32px;
	}
	.rail i {
		width: 12px;
		height: 12px;
		border-radius: 50%;
		background: var(--surface-card);
		border: 2px solid var(--border-strong);
		transition: all 400ms var(--ease-expo);
	}
	.rail .done i {
		background: #8fbff0;
		border-color: #8fbff0;
	}
	.rail .on i {
		width: 16px;
		height: 16px;
		background: linear-gradient(135deg, #6ab8ff, #cfa3f6);
		border-color: #fff;
		box-shadow: 0 0 0 5px rgba(164, 212, 247, 0.4);
	}
	.rail button span {
		position: absolute;
		bottom: calc(100% + 10px);
		left: 50%;
		transform: translate(-50%, 4px);
		padding: 5px 10px;
		border-radius: 8px;
		background: var(--ink-primary);
		color: #fff;
		font-size: 11.5px;
		font-weight: 500;
		white-space: nowrap;
		opacity: 0;
		pointer-events: none;
		transition: all 300ms var(--ease-out);
	}
	.rail button:hover span,
	.rail button:focus-visible span {
		opacity: 1;
		transform: translate(-50%, 0);
	}
</style>
