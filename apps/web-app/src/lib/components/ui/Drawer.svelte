<script lang="ts">
	import { tick, type Snippet } from 'svelte';
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
    let drawer: HTMLElement;
    let closeButton: HTMLButtonElement;
    $effect(() => {
        if (!open) return;
        const previous = document.activeElement as HTMLElement | null;
        const previousOverflow = document.documentElement.style.overflow;
        document.documentElement.style.overflow = 'hidden';
        tick().then(() => { if (open) closeButton?.focus(); });
        return () => {
            document.documentElement.style.overflow = previousOverflow;
            if (previous?.isConnected) previous.focus({ preventScroll: true });
        };
    });
    function onKey(e: KeyboardEvent) {
        if (!open) return;
        if (e.key === 'Escape') { e.preventDefault(); onclose(); }
        if (e.key !== 'Tab') return;
        const nodes = Array.from(drawer.querySelectorAll<HTMLElement>('button:not([disabled]), a[href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex="0"]')).filter((node) => node.getClientRects().length);
        const first = nodes[0], last = nodes[nodes.length - 1];
        if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last?.focus(); }
        else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first?.focus(); }
    }
</script>

<svelte:window onkeydown={onKey} />

<div class="scrim" class:open onclick={onclose} role="presentation"></div>
<div bind:this={drawer} tabindex="-1" role="dialog" aria-modal={open ? true : undefined} inert={!open} class="drawer" class:open style="--w:{width}px" aria-hidden={!open} aria-label={title}>
	<header>
		<div>
			{#if eyebrow}<span class="eyebrow">{eyebrow}</span>{/if}
			<h2>{title}</h2>
		</div>
		<button bind:this={closeButton} class="x" onclick={onclose} aria-label="Close">
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6"><path d="M6 6l12 12M18 6L6 18" /></svg>
		</button>
	</header>
	<div class="body">{#if open}{@render children()}{/if}</div>
	{#if footer && open}<footer>{@render footer()}</footer>{/if}
</div>

<style>
	.scrim {
		position: fixed;
		inset: 0;
		z-index: 200;
		background: rgba(75, 92, 120, 0.16);
		backdrop-filter: blur(3px);
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
		background: #fcfdfffa;
		backdrop-filter: blur(30px);
		border: 1px solid #fff;
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
