<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { tweenNumber } from '../motion/index.js';

	interface Props {
		value: number;
		decimals?: number;
		prefix?: string;
		suffix?: string;
		duration?: number;
		locale?: string;
		class?: string;
	}
	let { value, decimals = 0, prefix = '', suffix = '', duration = 1400, locale = 'en-US', class: klass = '' }: Props = $props();

	let shown = $state(0);
	let el: HTMLSpanElement;
	let visible = false;
	let stop = () => {};

	const fmt = (v: number) =>
		v.toLocaleString(locale, { minimumFractionDigits: decimals, maximumFractionDigits: decimals });

	onMount(() => {
		const io = new IntersectionObserver(([e]) => {
			if (e.isIntersecting && !visible) {
				visible = true;
				stop = tweenNumber(0, value, duration, (v) => (shown = v));
				io.disconnect();
			}
		});
		io.observe(el);
		return () => {
			io.disconnect();
			stop();
		};
	});

	let last = untrack(() => value);
	$effect(() => {
		const target = value;
		if (visible && target !== last) {
			stop();
			stop = tweenNumber(shown, target, 700, (v) => (shown = v));
		}
		last = target;
	});
</script>

<span bind:this={el} class="num {klass}">{prefix}{fmt(shown)}{suffix}</span>
