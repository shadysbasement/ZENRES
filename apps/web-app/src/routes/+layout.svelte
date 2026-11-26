<script lang="ts">
	import '$design/styles/fonts.css';
	import '$design/styles/tokens.css';
	import '../app.css';
	import { onNavigate } from '$app/navigation';

	let { children } = $props();
	let activeTransition: ViewTransition | undefined;

	onNavigate((navigation) => {
		if (window.matchMedia('(prefers-reduced-motion: reduce)').matches || !document.startViewTransition || navigation.from?.url.pathname === navigation.to?.url.pathname) return;
		activeTransition?.skipTransition();
		return new Promise((resolve) => {
			const transition = document.startViewTransition(async () => {
				resolve();
				await navigation.complete;
			});
			activeTransition = transition;
			transition.finished.catch(() => {}).finally(() => {
				if (activeTransition === transition) activeTransition = undefined;
			});
		});
	});
</script>

{@render children()}

<style>
	:global(:root) { view-transition-name: none; }
	:global(::view-transition) { pointer-events: none; }
	:global(::view-transition-old(workspace-content)) {
		animation: 220ms cubic-bezier(0.65, 0, 0.35, 1) both zv-out;
	}
	:global(::view-transition-new(workspace-content)) {
		animation: 520ms cubic-bezier(0.16, 1, 0.3, 1) both zv-in;
	}
	@keyframes -global-zv-out {
		to {
			opacity: 0;
		}
	}
	@keyframes -global-zv-in {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
	}
</style>
