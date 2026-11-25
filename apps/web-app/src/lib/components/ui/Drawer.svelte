<script lang="ts">
	import type { Snippet } from 'svelte';
	interface Props {
		open: boolean;
		title: string;
		eyebrow?: string;
		width?: number;
		onclose: () => void;
		children: Snippet;
		footer?: Snippet;
	}
	let { open, title, eyebrow, width = 440, onclose, children, footer }: Props = $props();
</script>

<svelte:window onkeydown={(e) => open && e.key === 'Escape' && onclose()} />

<div class="scrim" class:open onclick={onclose} role="presentation"></div>
<aside class="drawer" class:open style="--w:{width}px" aria-hidden={!open} aria-label={title}>
	<header>
		<div>
			{#if eyebrow}<span class="eyebrow">{eyebrow}</span>{/if}
			<h2>{title}</h2>
		</div>
		<button class="x" onclick={onclose} aria-label="Close">
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6"><path d="M6 6l12 12M18 6L6 18" /></svg>
		</button>
	</header>
	<div class="body">{#if open}{@render children()}{/if}</div>
	{#if footer && open}<footer>{@render footer()}</footer>{/if}
</aside>

<style>
	.scrim {
		position: fixed;
		inset: 0;
		z-index: 200;
		background: rgba(14, 20, 18, 0.18);
		opacity: 0;
		pointer-events: none;
		transition: opacity 400ms var(--ease-out);
	}
	.scrim.open {
		opacity: 1;
		pointer-events: auto;
	}
	.drawer {
		position: fixed;
		top: 8px;
		right: 8px;
		bottom: 8px;
		z-index: 201;
		width: min(var(--w), calc(100vw - 16px));
		display: flex;
		flex-direction: column;
		background: var(--surface-card);
		border: var(--hairline);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-pop);
		transform: translateX(calc(100% + 24px));
		transition: transform 650ms var(--ease-expo);
	}
	.drawer.open {
		transform: none;
	}
	header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 12px;
		padding: 20px 22px 16px;
		border-bottom: var(--hairline);
	}
	h2 {
		margin-top: 4px;
		font-size: 1.3rem;
	}
	.x {
		width: 32px;
		height: 32px;
		display: grid;
		place-items: center;
		border-radius: var(--radius-sm);
		color: var(--ink-muted);
	}
	.x:hover {
		background: var(--surface-subtle);
		color: var(--ink-primary);
	}
	.body {
		flex: 1;
		overflow: auto;
		padding: 18px 22px;
	}
	footer {
		padding: 14px 22px;
		border-top: var(--hairline);
		display: flex;
		gap: 8px;
		justify-content: flex-end;
	}
</style>
