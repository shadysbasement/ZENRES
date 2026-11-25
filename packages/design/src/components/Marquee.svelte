<script lang="ts">
	import type { Snippet } from 'svelte';
	interface Props {
		speed?: number;
		reverse?: boolean;
		class?: string;
		children: Snippet;
	}
	let { speed = 40, reverse = false, class: klass = '', children }: Props = $props();
</script>

<div class="marquee {klass}" style="--speed:{speed}s" class:reverse>
	<div class="track">
		<div class="group">{@render children()}</div>
		<div class="group" aria-hidden="true">{@render children()}</div>
	</div>
</div>

<style>
	.marquee {
		overflow: hidden;
		mask-image: linear-gradient(90deg, transparent, #000 8%, #000 92%, transparent);
	}
	.track {
		display: flex;
		width: max-content;
		animation: zv-marquee var(--speed) linear infinite;
	}
	.reverse .track {
		animation-direction: reverse;
	}
	.marquee:hover .track {
		animation-play-state: paused;
	}
	.group {
		display: flex;
		align-items: center;
		flex: none;
	}
	@keyframes zv-marquee {
		to {
			transform: translate3d(-50%, 0, 0);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.track {
			animation: none;
		}
	}
</style>
