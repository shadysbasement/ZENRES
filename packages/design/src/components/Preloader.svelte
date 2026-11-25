<script lang="ts">
	import { onMount } from 'svelte';
	import { prefersReducedMotion } from '../motion/index.js';

	interface Props {
		/** Large word revealed letter by letter. */
		word?: string;
		/** Rotating status lines shown bottom-right. */
		status?: string[];
		/** Minimum display time in ms. */
		minDuration?: number;
		/** Session key: preloader shows once per tab session. Set to '' to always show. */
		sessionKey?: string;
		tagline?: string;
		ondone?: () => void;
	}
	let {
		word = 'ZEVRA',
		status = ['Loading CN-code engine', 'Calibrating emission factors', 'Linking QR passports', 'Mapping freight corridors'],
		minDuration = 2100,
		sessionKey = 'zv-preloaded',
		tagline = 'Zero Emission Vision — Responsible Advancement',
		ondone
	}: Props = $props();

	let progress = $state(0);
	let phase = $state<'in' | 'count' | 'out' | 'gone'>('in');
	let statusIndex = $state(0);

	const pad = (n: number) => String(Math.round(n)).padStart(3, '0');

	onMount(() => {
		const skip = prefersReducedMotion() || (sessionKey && sessionStorage.getItem(sessionKey) === '1');
		if (skip) {
			phase = 'gone';
			ondone?.();
			return;
		}
		document.documentElement.style.overflow = 'hidden';
		const fontsReady = (document as Document & { fonts?: FontFaceSet }).fonts?.ready ?? Promise.resolve();
		let loaded = false;
		Promise.all([fontsReady, new Promise((r) => (document.readyState === 'complete' ? r(0) : window.addEventListener('load', r, { once: true })))]).then(
			() => (loaded = true)
		);

		const start = performance.now();
		let raf = 0;
		requestAnimationFrame(() => (phase = 'count'));
		const tick = (now: number) => {
			const t = Math.min(1, (now - start) / minDuration);
			// eased progress that waits at 92% until assets are ready
			const eased = 1 - Math.pow(1 - t, 3);
			const target = loaded ? eased * 100 : Math.min(eased * 100, 92);
			progress += (target - progress) * 0.16;
			statusIndex = Math.min(status.length - 1, Math.floor((progress / 100) * status.length));
			if (t >= 1 && loaded && progress > 99.4) {
				progress = 100;
				finish();
				return;
			}
			raf = requestAnimationFrame(tick);
		};
		raf = requestAnimationFrame(tick);

		function finish() {
			setTimeout(() => {
				phase = 'out';
				if (sessionKey) sessionStorage.setItem(sessionKey, '1');
				document.documentElement.style.overflow = '';
				setTimeout(() => ondone?.(), 380);
				setTimeout(() => (phase = 'gone'), 1300);
			}, 220);
		}
		return () => {
			cancelAnimationFrame(raf);
			document.documentElement.style.overflow = '';
		};
	});
</script>

{#if phase !== 'gone'}
	<div class="preloader" class:count={phase === 'count'} class:out={phase === 'out'} role="progressbar" aria-valuemin={0} aria-valuemax={100} aria-valuenow={Math.round(progress)} aria-label="Loading Zevra">
		<div class="grid-lines" aria-hidden="true">
			{#each Array(6) as _, i (i)}<span style="--i:{i}"></span>{/each}
		</div>
		<div class="top">
			<span class="mono">{tagline}</span>
			<span class="mono">BUILD 2026.11 // ZV-CORE</span>
		</div>
		<div class="word" aria-hidden="true">
			{#each word.split('') as ch, i (i)}
				<span class="mask"><span class="ch" style="--i:{i}">{ch}</span></span>
			{/each}
		</div>
		<div class="bottom">
			<div class="counter mono">{pad(progress)}<span class="pct">%</span></div>
			<div class="status mono">
				{#each status as s, i (s)}
					<span class:active={i === statusIndex} class:done={i < statusIndex}>{i < statusIndex ? '✓' : '·'} {s}</span>
				{/each}
			</div>
		</div>
		<div class="track" aria-hidden="true"><span style="transform:scaleX({progress / 100})"></span></div>
	</div>
{/if}

<style>
	.preloader {
		position: fixed;
		inset: 0;
		z-index: 9999;
		background: var(--obsidian);
		color: var(--ink-inverse);
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		padding: clamp(20px, 3vw, 40px);
		clip-path: inset(0 0 0 0);
		transition: clip-path 1.05s var(--ease-expo);
	}
	.preloader.out {
		clip-path: inset(0 0 100% 0);
	}
	.grid-lines {
		position: absolute;
		inset: 0;
		display: grid;
		grid-template-columns: repeat(6, 1fr);
		pointer-events: none;
	}
	.grid-lines span {
		border-left: 1px solid var(--obsidian-line);
		transform: scaleY(0);
		transform-origin: top;
		transition: transform 1.4s var(--ease-expo);
		transition-delay: calc(var(--i) * 60ms);
	}
	.count .grid-lines span {
		transform: scaleY(1);
	}
	.top,
	.bottom {
		position: relative;
		display: flex;
		justify-content: space-between;
		gap: 24px;
	}
	.top .mono {
		font-size: 11px;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--vault-dim);
		opacity: 0;
		transition: opacity 0.8s var(--ease-out) 0.2s;
	}
	.count .top .mono {
		opacity: 1;
	}
	.word {
		position: relative;
		display: flex;
		justify-content: center;
		font-family: var(--font-sans);
		font-weight: 600;
		font-size: clamp(5rem, 22vw, 22rem);
		letter-spacing: -0.04em;
		line-height: 0.82;
	}
	.mask {
		display: inline-block;
		overflow: hidden;
		padding: 0.04em 0.01em;
	}
	.ch {
		display: inline-block;
		transform: translate3d(0, 105%, 0);
		transition: transform 1.1s var(--ease-expo);
		transition-delay: calc(var(--i) * 70ms + 120ms);
	}
	.count .ch {
		transform: translate3d(0, 0, 0);
	}
	.out .ch {
		transform: translate3d(0, -105%, 0);
		transition-duration: 0.7s;
		transition-delay: calc(var(--i) * 35ms);
	}
	.mask:last-child .ch {
		color: var(--vault-phosphor);
	}
	.counter {
		font-size: clamp(2.5rem, 6vw, 5rem);
		font-weight: 400;
		line-height: 1;
		letter-spacing: -0.04em;
	}
	.pct {
		font-size: 0.4em;
		color: var(--vault-dim);
		margin-left: 4px;
		vertical-align: top;
	}
	.status {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 4px;
		font-size: 11px;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		align-self: flex-end;
	}
	.status span {
		color: var(--obsidian-line);
		transition: color 0.4s var(--ease-out);
	}
	.status span.done {
		color: var(--vault-dim);
	}
	.status span.active {
		color: var(--vault-phosphor);
	}
	.track {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		height: 2px;
		background: var(--obsidian-2);
	}
	.track span {
		display: block;
		height: 100%;
		background: var(--vault-phosphor);
		transform-origin: left;
	}
	@media (max-width: 640px) {
		.status {
			display: none;
		}
		.top .mono:last-child {
			display: none;
		}
	}
</style>
