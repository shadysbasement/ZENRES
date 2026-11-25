<script lang="ts">
	import '$design/styles/fonts.css';
	import '$design/styles/tokens.css';
	import '../app.css';
	import { onNavigate } from '$app/navigation';

	let { children } = $props();

	onNavigate((navigation) => {
		if (!document.startViewTransition || navigation.from?.url.pathname === navigation.to?.url.pathname) return;
		return new Promise((resolve) => {
			document.startViewTransition(async () => {
				resolve();
				await navigation.complete;
			});
		});
	});
</script>

{@render children()}

<style>
	:global(::view-transition-old(root)) {
		animation: 220ms cubic-bezier(0.65, 0, 0.35, 1) both zv-out;
	}
	:global(::view-transition-new(root)) {
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
