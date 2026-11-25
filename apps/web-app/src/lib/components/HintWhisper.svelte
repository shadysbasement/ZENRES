<script lang="ts">
	import { ws } from '$lib/stores/workspace.svelte';
	interface Props {
		context: string;
	}
	let { context }: Props = $props();
	const hint = $derived(ws.hints[context]);
</script>

{#if hint}
	<div class="whisper" role="note" data-hint={hint.code}>
		<span class="mark" aria-hidden="true"></span>
		<p>{hint.message}</p>
		<a href={hint.learn_more} class="learn">Learn more →</a>
		<button class="dismiss" onclick={() => ws.dismissHint(hint.code)} aria-label="Dismiss suggestion">Dismiss</button>
	</div>
{/if}

<style>
	.whisper {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 9px 12px;
		margin: 10px 0;
		border-radius: var(--radius-sm);
		background: #f1f5f2;
		border: 1px solid #dfe8e2;
		font-size: var(--fs-xs);
		color: var(--ink-secondary);
		animation: whisper 600ms var(--ease-expo);
	}
	@keyframes whisper {
		from {
			opacity: 0;
			transform: translateY(-4px);
		}
	}
	.mark {
		width: 3px;
		align-self: stretch;
		border-radius: 2px;
		background: var(--brand-emerald-mid);
		opacity: 0.5;
	}
	p {
		flex: 1;
		line-height: 1.45;
	}
	.learn {
		color: var(--brand-emerald);
		font-weight: 500;
		white-space: nowrap;
	}
	.dismiss {
		font-size: 11px;
		color: var(--ink-muted);
		white-space: nowrap;
	}
	.dismiss:hover {
		color: var(--ink-primary);
	}
</style>
