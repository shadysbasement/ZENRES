<script lang="ts">
	import type { Snippet } from 'svelte';
	interface Props {
		label: string;
		value: string;
		delta?: string;
		tone?: 'up' | 'down' | 'flat';
		href?: string;
		i?: number;
		children?: Snippet;
	}
	let { label, value, delta, tone = 'flat', href, i = 0, children }: Props = $props();
</script>

<svelte:element this={href ? 'a' : 'div'} {href} class="kpi card" style="--i:{i}">
	<div class="top">
		<span class="eyebrow">{label}</span>
		{#if href}<span class="go" aria-hidden="true">→</span>{/if}
	</div>
	<div class="val num">{value}</div>
	{#if delta}<span class="delta {tone}">{delta}</span>{/if}
	{#if children}<div class="extra">{@render children()}</div>{/if}
</svelte:element>

<style>
	.kpi {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 16px 18px;
		min-width: 0;
		transition: border-color 300ms, transform 500ms var(--ease-expo), box-shadow 300ms;
		animation: k-in 800ms var(--ease-expo) both;
		animation-delay: calc(var(--i) * 60ms);
	}
	@keyframes k-in {
		from {
			opacity: 0;
			transform: translateY(12px);
		}
	}
	a.kpi:hover {
		border-color: var(--border-strong);
		transform: translateY(-2px);
		box-shadow: var(--shadow-float);
	}
	.top {
		display: flex;
		justify-content: space-between;
	}
	.go {
		color: var(--ink-faint);
		transition: transform 300ms var(--ease-out), color 300ms;
	}
	a.kpi:hover .go {
		transform: translateX(3px);
		color: var(--brand-emerald);
	}
	.val {
		font-size: 1.6rem;
		letter-spacing: -0.04em;
		color: var(--ink-primary);
		line-height: 1.1;
		white-space: nowrap;
	}
	.delta {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--ink-muted);
	}
	.delta.up {
		color: var(--brand-emerald);
	}
	.delta.down {
		color: var(--status-crimson);
	}
	.extra {
		margin-top: 6px;
	}
</style>
