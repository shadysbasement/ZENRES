<script lang="ts">
	import { fly } from 'svelte/transition';
	import { appearance } from '$lib/stores/appearance.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
</script>

<div class="toasts" aria-live="polite">
	{#each ws.toasts as t (t.id)}
		<div class="toast {t.tone}" transition:fly={{ y: 12, duration: appearance.motion ? 280 : 0 }}><span class="dot"></span>{t.text}</div>
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
		width: min(600px, calc(100vw - 36px));
	}
	.toast {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 16px;
		background: #fcfdfff5;
		border: 1px solid white;
		backdrop-filter: blur(20px);
		color: var(--ink-primary);
		border-radius: 100px;
		font-size: var(--fs-xs);
		box-shadow: var(--shadow-pop);

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
