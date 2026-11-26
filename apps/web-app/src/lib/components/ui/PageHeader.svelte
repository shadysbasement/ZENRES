<script lang="ts">
	import type { Snippet } from 'svelte';
	interface Props {
		eyebrow: string;
		title: string;
		serif?: string;
		lede?: string;
		actions?: Snippet;
	}
	let { eyebrow, title, serif, lede, actions }: Props = $props();
</script>

<header class="ph">
	<div class="ph-l">
		<span class="eyebrow">{eyebrow}</span>
		<h1>{title}{#if serif}{' '}<span class="serif">{serif}</span>{/if}</h1>
		{#if lede}<p>{lede}</p>{/if}
	</div>
	{#if actions}<div class="ph-r">{@render actions()}</div>{/if}
</header>

<style>
	.ph {
		display: flex;
		justify-content: space-between;
		align-items: flex-end;
		gap: 24px;
		margin-bottom: 26px;
		padding: 8px 0 12px;
		animation: ph 800ms var(--ease-expo);
	}
	@keyframes ph {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
	}
	.ph-l {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	h1 {
		font-size: clamp(1.6rem, 2.3vw, 2.15rem);
		font-weight: 450;
		letter-spacing: -0.03em;
		line-height: 1.05;
	}
	h1 .serif {
		color: inherit;
		font-style: normal;
	}
	p {
		max-width: 70ch;
		font-size: var(--fs-sm);
	}
	.ph-r {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
		justify-content: flex-end;
	}
	@media (max-width: 800px) {
		.ph {
			flex-direction: column;
			align-items: flex-start;
		}
	}
</style>
