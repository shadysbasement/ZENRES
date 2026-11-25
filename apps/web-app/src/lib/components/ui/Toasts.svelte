<script lang="ts">
	import { ws } from '$lib/stores/workspace.svelte';
</script>

<div class="toasts" aria-live="polite">
	{#each ws.toasts as t (t.id)}
		<div class="toast {t.tone}"><span class="dot"></span>{t.text}</div>
	{/each}
</div>

<style>
	.toasts {
		position: fixed;
		left: 50%;
		bottom: 22px;
		transform: translateX(-50%);
		z-index: 500;
		display: flex;
		flex-direction: column;
		gap: 8px;
		align-items: center;
		pointer-events: none;
	}
	.toast {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 16px;
		background: var(--obsidian);
		color: var(--ink-inverse);
		border-radius: var(--radius-md);
		font-size: var(--fs-xs);
		box-shadow: var(--shadow-pop);
		animation: t-in 600ms var(--ease-expo);
	}
	.toast .dot {
		color: var(--vault-phosphor);
	}
	.toast.warn .dot {
		color: var(--status-terracotta);
	}
	.toast.err .dot {
		color: var(--status-crimson);
	}
	@keyframes t-in {
		from {
			opacity: 0;
			transform: translateY(14px) scale(0.96);
		}
	}
</style>
