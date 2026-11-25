<script lang="ts">
	import { onMount } from 'svelte';
	import { JOURNEY } from '$lib/content';
	import { prefersReducedMotion, reveal } from '$design/motion';

	let section: HTMLElement;
	let track: HTMLElement;
	let x = $state(0);
	let progress = $state(0);
	let height = $state('auto');
	let horizontal = $state(false);

	onMount(() => {
		const mq = window.matchMedia('(min-width: 961px)');
		let distance = 0;
		const measure = () => {
			horizontal = mq.matches && !prefersReducedMotion();
			if (!horizontal) {
				height = 'auto';
				x = 0;
				return;
			}
			distance = Math.max(0, track.scrollWidth - window.innerWidth + 80);
			height = `${window.innerHeight + distance}px`;
			update();
		};
		let raf = 0;
		const update = () => {
			raf = 0;
			if (!horizontal) return;
			const r = section.getBoundingClientRect();
			const total = r.height - window.innerHeight;
			progress = Math.min(1, Math.max(0, -r.top / Math.max(1, total)));
			x = -progress * distance;
		};
		const onScroll = () => {
			if (!raf) raf = requestAnimationFrame(update);
		};
		measure();
		const ro = new ResizeObserver(measure);
		ro.observe(track);
		window.addEventListener('resize', measure);
		window.addEventListener('scroll', onScroll, { passive: true });
		return () => {
			ro.disconnect();
			window.removeEventListener('resize', measure);
			window.removeEventListener('scroll', onScroll);
			cancelAnimationFrame(raf);
		};
	});

	const activeStep = $derived(Math.min(JOURNEY.length - 1, Math.floor(progress * JOURNEY.length)));
</script>

<section class="journey" bind:this={section} style="height:{height}" aria-labelledby="journey-title">
	<div class="pin" class:horizontal>
		<div class="container-wide head">
			<div>
				<span class="eyebrow" use:reveal>(03) — End-to-end customer journey</span>
				<h2 id="journey-title" use:reveal={{ delay: 80 }}>From a Tally voucher <span class="serif">to a CBAM submission.</span></h2>
			</div>
			<div class="counter mono" aria-hidden="true">
				<span class="cur">{String(activeStep + 1).padStart(2, '0')}</span><span class="sep">/</span><span>08</span>
				<div class="meter"><span style="transform:scaleX({horizontal ? progress : 1})"></span></div>
			</div>
		</div>

		<ol class="track" bind:this={track} style="transform:translate3d({x}px,0,0)" role="list">
			{#each JOURNEY as step, i (step.n)}
				<li class="step" class:active={horizontal && i <= activeStep} use:reveal={{ delay: horizontal ? 0 : i * 40 }}>
					<div class="step-top">
						<span class="n mono">{step.n}</span>
						<span class="pill pill-outline">{step.tag}</span>
					</div>
					<div class="line" aria-hidden="true"><span class="node"></span></div>
					<h3>{step.title}</h3>
					<p>{step.body}</p>
				</li>
			{/each}
		</ol>
	</div>
</section>

<style>
	.journey {
		position: relative;
	}
	.pin {
		padding-block: 120px 80px;
		overflow: hidden;
	}
	.pin.horizontal {
		position: sticky;
		top: 0;
		height: 100vh;
		display: flex;
		flex-direction: column;
		justify-content: center;
		padding-block: 0;
	}
	.head {
		display: flex;
		justify-content: space-between;
		align-items: flex-end;
		gap: 40px;
		margin-bottom: 64px;
	}
	.head h2 {
		margin-top: 18px;
		font-size: var(--fs-h1);
		letter-spacing: var(--tracking-display);
		max-width: 16ch;
	}
	.head h2 .serif {
		color: var(--brand-emerald);
	}
	.counter {
		text-align: right;
		font-size: 14px;
		color: var(--ink-muted);
		min-width: 160px;
	}
	.counter .cur {
		font-size: 3.5rem;
		color: var(--ink-primary);
		letter-spacing: -0.04em;
		line-height: 1;
	}
	.counter .sep {
		margin: 0 6px;
	}
	.meter {
		margin-top: 14px;
		height: 1px;
		background: var(--border-hairline);
	}
	.meter span {
		display: block;
		height: 100%;
		background: var(--brand-emerald);
		transform-origin: left;
	}
	.track {
		display: flex;
		gap: 0;
		padding-left: max(var(--gutter), calc((100vw - var(--max-w-wide)) / 2));
		padding-right: var(--gutter);
		will-change: transform;
		list-style: none;
		margin: 0;
	}
	.step {
		flex: none;
		width: clamp(300px, 27vw, 380px);
		padding-right: 36px;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.step-top {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	.n {
		font-size: 13px;
		color: var(--ink-faint);
		transition: color 400ms;
	}
	.line {
		position: relative;
		height: 1px;
		background: var(--border-hairline);
		margin-right: -36px;
	}
	.node {
		position: absolute;
		left: 0;
		top: -5px;
		width: 11px;
		height: 11px;
		border-radius: 50%;
		background: var(--surface-canvas);
		border: 1px solid var(--border-strong);
		transition: background 500ms, border-color 500ms, transform 500ms var(--ease-expo);
	}
	.step h3 {
		font-size: 1.5rem;
		letter-spacing: -0.02em;
		margin-top: 8px;
		color: var(--ink-faint);
		transition: color 500ms;
	}
	.step p {
		font-size: var(--fs-sm);
		line-height: 1.6;
		max-width: 32ch;
	}
	.step.active .node,
	.pin:not(.horizontal) .node {
		background: var(--brand-emerald);
		border-color: var(--brand-emerald);
		transform: scale(1.15);
	}
	.step.active h3,
	.pin:not(.horizontal) .step h3 {
		color: var(--ink-primary);
	}
	.step.active .n {
		color: var(--brand-emerald);
	}
	@media (max-width: 960px) {
		.head {
			flex-direction: column;
			align-items: flex-start;
		}
		.counter {
			display: none;
		}
		.track {
			flex-direction: column;
			padding-inline: var(--gutter);
			gap: 0;
		}
		.step {
			width: 100%;
			padding: 0 0 36px 28px;
			border-left: var(--hairline);
			position: relative;
		}
		.line {
			position: absolute;
			left: -6px;
			top: 2px;
			background: none;
		}
	}
</style>
