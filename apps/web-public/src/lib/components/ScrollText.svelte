<script lang="ts">
	import { onMount } from 'svelte';
	import { prefersReducedMotion } from '$design/motion';

	interface Props {
		text: string;
		/** words wrapped in *asterisks* render in the editorial serif */
		class?: string;
	}
	let { text, class: klass = '' }: Props = $props();

	const words = $derived(text.split(' ').map((w) => ({ t: w.replace(/\*/g, ''), serif: w.startsWith('*') || w.endsWith('*') })));
	let el: HTMLElement;
	let progress = $state(0);

	onMount(() => {
		if (prefersReducedMotion()) {
			progress = 1;
			return;
		}
		let raf = 0;
		const update = () => {
			raf = 0;
			const r = el.getBoundingClientRect();
			const vh = window.innerHeight;
			progress = Math.min(1, Math.max(0, (vh * 0.85 - r.top) / (r.height + vh * 0.35)));
		};
		const onScroll = () => {
			if (!raf) raf = requestAnimationFrame(update);
		};
		update();
		window.addEventListener('scroll', onScroll, { passive: true });
		return () => window.removeEventListener('scroll', onScroll);
	});
</script>

<p class="st {klass}" bind:this={el}>
	{#each words as w, i (i)}
		{@const local = Math.min(1, Math.max(0, progress * words.length * 1.15 - i))}
		<span class:serif={w.serif} style="opacity:{0.14 + local * 0.86}">{w.t}</span>{' '}
	{/each}
</p>

<style>
	.st {
		font-size: clamp(1.75rem, 3.6vw, 3.4rem);
		line-height: 1.16;
		letter-spacing: -0.03em;
		font-weight: 500;
		color: var(--ink-primary);
	}
	.st span {
		transition: opacity 200ms linear;
	}
	.st .serif {
		color: var(--brand-emerald);
	}
</style>
