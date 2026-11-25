<script lang="ts">
	interface Props {
		size?: number;
		inverse?: boolean;
		wordmark?: boolean;
		class?: string;
	}
	let { size = 26, inverse = false, wordmark = true, class: klass = '' }: Props = $props();
</script>

<span class="zv-logo {klass}" class:inverse style="--s:{size}px">
	<svg class="mark" viewBox="0 0 28 28" width={size} height={size} aria-hidden="true">
		<rect x="0" y="0" width="28" height="28" rx="5" class="tile" />
		<path d="M8 8.5h11.5L8.5 19.5H20" class="z" />
		<rect x="19.25" y="5.25" width="3.5" height="3.5" rx="0.6" class="node" />
	</svg>
	{#if wordmark}
		<span class="word">ZEVRA</span>
	{/if}
	<span class="sr-only">Zevra</span>
</span>

<style>
	.zv-logo {
		display: inline-flex;
		align-items: center;
		gap: calc(var(--s) * 0.42);
		color: var(--ink-primary);
	}
	.tile {
		fill: var(--ink-primary);
		transition: fill var(--dur-2) var(--ease-out);
	}
	.z {
		fill: none;
		stroke: var(--surface-canvas);
		stroke-width: 2.4;
		stroke-linecap: square;
		stroke-linejoin: miter;
	}
	.node {
		fill: var(--vault-phosphor);
	}
	.word {
		font-family: var(--font-sans);
		font-weight: 650;
		font-size: calc(var(--s) * 0.66);
		letter-spacing: 0.16em;
		line-height: 1;
		margin-right: -0.16em;
	}
	.inverse {
		color: var(--ink-inverse);
	}
	.inverse .tile {
		fill: var(--ink-inverse);
	}
	.inverse .z {
		stroke: var(--obsidian);
	}
</style>
