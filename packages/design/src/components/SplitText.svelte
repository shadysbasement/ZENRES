<script lang="ts" module>
	export type Segment = { text: string; serif?: boolean; accent?: boolean };
</script>

<script lang="ts">
	import { inview } from '../motion/index.js';
	interface Props {
		/** Either a plain string (\n splits lines) or explicit lines of styled segments. */
		text?: string;
		lines?: Segment[][];
		tag?: string;
		class?: string;
		delay?: number;
		/** When true the parent controls `.is-in` (e.g. after a preloader finishes). */
		manual?: boolean;
		play?: boolean;
	}
	let { text = '', lines, tag = 'h2', class: klass = '', delay = 0, manual = false, play = false }: Props = $props();

	const resolved = $derived<Segment[][]>(lines ?? text.split('\n').map((l) => [{ text: l }]));

	const words = $derived.by(() => {
		let w = 0;
		return resolved.map((line) =>
			line.flatMap((seg) =>
				seg.text
					.split(/(\s+)/)
					.filter((s) => s.length)
					.map((t) => ({ t, space: /^\s+$/.test(t), serif: seg.serif, accent: seg.accent, i: /^\s+$/.test(t) ? -1 : w++ }))
			)
		);
	});

	function maybeInview(node: HTMLElement) {
		if (manual) return {};
		return inview(node, { delay });
	}
</script>

<svelte:element this={tag} class="split {klass}" class:is-in={manual && play} style="--delay:{delay}ms" use:maybeInview aria-label={resolved.map((l) => l.map((s) => s.text).join('')).join(' ')}>
	{#each words as line, li (li)}
		<span class="split-line" aria-hidden="true">
			{#each line as word, wi (wi)}
				{#if word.space}{' '}{:else}<span class="split-word"><span class="split-inner" class:serif={word.serif} class:accent={word.accent} style="--w:{word.i}">{word.t}</span></span>{/if}
			{/each}
		</span>
	{/each}
</svelte:element>

<style>
	.accent {
		color: var(--brand-emerald);
	}
</style>
