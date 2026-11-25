<script lang="ts" generics="T extends string">
	interface Props {
		options: { value: T; label: string }[];
		value: T;
		label?: string;
		size?: 'sm' | 'md';
		onchange?: (v: T) => void;
	}
	let { options, value = $bindable(), label = 'Options', size = 'md', onchange }: Props = $props();

	let buttons: HTMLButtonElement[] = $state([]);
	let thumb = $state({ left: 3, width: 0 });

	$effect(() => {
		const i = options.findIndex((o) => o.value === value);
		const b = buttons[i];
		if (!b) return;
		const place = () => (thumb = { left: b.offsetLeft, width: b.offsetWidth });
		place();
		const ro = new ResizeObserver(place);
		ro.observe(b);
		return () => ro.disconnect();
	});

	function pick(v: T) {
		value = v;
		onchange?.(v);
	}
</script>

<div class="seg" class:sm={size === 'sm'} role="group" aria-label={label}>
	<span class="seg-thumb" style="left:{thumb.left}px;width:{thumb.width}px" aria-hidden="true"></span>
	{#each options as o, i (o.value)}
		<button bind:this={buttons[i]} type="button" aria-pressed={value === o.value} onclick={() => pick(o.value)}>{o.label}</button>
	{/each}
</div>

<style>
	.sm button {
		height: 26px;
		padding: 0 10px;
	}
</style>
