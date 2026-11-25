<script lang="ts">
	import { reveal } from '$design/motion';
	import SplitText from '$design/components/SplitText.svelte';
	import type { Snippet } from 'svelte';

	interface Props {
		index: string;
		label: string;
		title: string;
		serif?: string;
		id?: string;
		children?: Snippet;
	}
	let { index, label, title, serif = '', id, children }: Props = $props();
</script>

<div class="sh">
	<div class="sh-left">
		<span class="eyebrow" use:reveal><span class="idx">({index})</span> — {label}</span>
		<SplitText tag="h2" class="sh-title" lines={[[{ text: title + (serif ? ' ' : '') }, ...(serif ? [{ text: serif, serif: true, accent: true }] : [])]]} />
	</div>
	{#if children}
		<div class="sh-right" use:reveal={{ delay: 150 }}>{@render children()}</div>
	{/if}
	<span class="rule" use:reveal={{ delay: 100 }} data-reveal="line" {id}></span>
</div>

<style>
	.sh {
		position: relative;
		display: grid;
		grid-template-columns: 7fr 4fr;
		gap: var(--gap);
		align-items: end;
		padding-bottom: 40px;
		margin-bottom: 48px;
	}
	.idx {
		color: var(--brand-emerald);
	}
	.sh-left {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}
	:global(.sh-title) {
		font-size: var(--fs-h1);
		letter-spacing: var(--tracking-display);
		line-height: 1;
		max-width: 18ch;
	}
	.sh-right {
		font-size: var(--fs-body);
		line-height: 1.65;
		color: var(--ink-secondary);
		justify-self: end;
		max-width: 44ch;
	}
	.rule {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		height: 1px;
		background: var(--border-hairline);
		scroll-margin-top: 140px;
	}
	@media (max-width: 900px) {
		.sh {
			grid-template-columns: 1fr;
		}
		.sh-right {
			justify-self: start;
		}
	}
</style>
