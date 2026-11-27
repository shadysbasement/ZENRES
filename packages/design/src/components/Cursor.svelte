<script lang="ts">
	import { onMount } from 'svelte';

	let x = $state(-100);
	let y = $state(-100);
	let rx = $state(-100);
	let ry = $state(-100);
	let mode = $state<'idle' | 'link' | 'label' | 'text'>('idle');
	let label = $state('');
	let hidden = $state(true);
	let pressed = $state(false);
	let enabled = $state(false);

	onMount(() => {
		const fine = window.matchMedia('(hover: hover) and (pointer: fine)').matches;
		const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
		if (!fine || reduced) return;
		enabled = true;
		document.documentElement.classList.add('has-cursor');
		let raf = 0;
		const loop = () => {
			rx += (x - rx) * 0.16;
			ry += (y - ry) * 0.16;
			raf = requestAnimationFrame(loop);
		};
		raf = requestAnimationFrame(loop);
		const move = (e: PointerEvent) => {
			x = e.clientX;
			y = e.clientY;
			hidden = false;
			const t = e.target as HTMLElement | null;
			const labelled = t?.closest('[data-cursor]') as HTMLElement | null;
			if (labelled) {
				mode = 'label';
				label = labelled.dataset.cursor ?? '';
			} else if (t?.closest('a,button,[role="button"],[role="tab"],label,summary,select')) {
				mode = 'link';
			} else if (t?.closest('input,textarea,[contenteditable="true"]')) {
				mode = 'text';
			} else mode = 'idle';
		};
		const leave = () => (hidden = true);
		const down = () => (pressed = true);
		const up = () => (pressed = false);
		window.addEventListener('pointermove', move, { passive: true });
		document.addEventListener('pointerleave', leave);
		window.addEventListener('pointerdown', down);
		window.addEventListener('pointerup', up);
		return () => {
			cancelAnimationFrame(raf);
			document.documentElement.classList.remove('has-cursor');
			window.removeEventListener('pointermove', move);
			document.removeEventListener('pointerleave', leave);
			window.removeEventListener('pointerdown', down);
			window.removeEventListener('pointerup', up);
		};
	});
</script>

{#if enabled}
	<div class="cursor-dot" class:hidden class:hide-dot={mode === 'label' || mode === 'text'} style="transform:translate3d({x}px,{y}px,0)" aria-hidden="true"></div>
	<div
		class="cursor-ring"
		class:hidden={hidden || mode === 'text'}
		class:link={mode === 'link'}
		class:label={mode === 'label'}
		class:pressed
		style="transform:translate3d({rx}px,{ry}px,0)"
		aria-hidden="true"
	>
		<span class="ring"><span class="txt">{label}</span></span>
	</div>
{/if}

<style>
	.cursor-dot,
	.cursor-ring {
		position: fixed;
		top: 0;
		left: 0;
		z-index: 10000;
		pointer-events: none;
		will-change: transform;
	}
	.cursor-dot {
		width: 6px;
		height: 6px;
		margin: -3px 0 0 -3px;
		border-radius: 50%;
		background: var(--ink-primary);
		mix-blend-mode: difference;
		filter: invert(1);
		transition: opacity 0.2s;
	}
	.cursor-dot.hide-dot {
		opacity: 0;
	}
	.ring {
		position: absolute;
		left: 0;
		top: 0;
		width: 34px;
		height: 34px;
		margin: -17px 0 0 -17px;
		border-radius: 50%;
		border: 1px solid rgba(67, 108, 155, 0.35);
		display: grid;
		place-items: center;
		transition:
			width 0.5s var(--ease-expo),
			height 0.5s var(--ease-expo),
			margin 0.5s var(--ease-expo),
			background 0.4s var(--ease-out),
			border-color 0.4s var(--ease-out),
			opacity 0.3s;
	}
	.link .ring {
		width: 54px;
		height: 54px;
		margin: -27px 0 0 -27px;
		background: rgba(106, 184, 255, 0.12);
		border-color: rgba(130, 101, 168, 0.4);
	}
	.label .ring {
		width: 84px;
		height: 84px;
		margin: -42px 0 0 -42px;
		background: linear-gradient(135deg, #6ab8ff, #a98fcf);
		border-color: transparent;
	}
	.pressed .ring {
		width: 26px;
		height: 26px;
		margin: -13px 0 0 -13px;
	}
	.txt {
		font-family: var(--font-mono);
		font-size: 10px;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: #fff;
		opacity: 0;
		transition: opacity 0.3s;
	}
	.label .txt {
		opacity: 1;
	}
	.hidden .ring,
	.cursor-dot.hidden {
		opacity: 0;
	}
	:global(html.has-cursor),
	:global(html.has-cursor *) {
		cursor: none !important;
	}
	:global(html.has-cursor input),
	:global(html.has-cursor textarea) {
		cursor: text !important;
	}
</style>
