<script lang="ts">
	interface Props {
		name: string;
		size?: number;
		class?: string;
	}
	let { name, size = 16, class: klass = '' }: Props = $props();
	const P: Record<string, string> = {
		dashboard: 'M3 3h7v7H3zM14 3h7v4h-7zM14 11h7v10h-7zM3 14h7v7H3z',
		inbox: 'M3 13l3-8h12l3 8v6H3zM3 13h5l1.5 3h5L16 13h5',
		plant: 'M3 21V10l5 3V10l5 3V6l8 4v11zM7 17h2M12 17h2M17 17h2',
		product: 'M12 3l8 4.5v9L12 21l-8-4.5v-9zM12 12l8-4.5M12 12v9M12 12L4 7.5',
		inventory: 'M4 7h16v13H4zM4 7l2-3h12l2 3M9 11h6',
		finance: 'M4 19V5M4 19h16M8 15l4-4 3 3 5-6',
		tally: 'M4 6h16M4 12h16M4 18h10M18 16l2 2-2 2',
		inputs: 'M12 3v12M7 10l5 5 5-5M4 19h16',
		cbam: 'M12 3a9 9 0 1 0 9 9M12 3v9h9M12 3a9 9 0 0 1 9 9',
		exports: 'M14 3h7v7M21 3l-9 9M19 14v6H4V5h6',
		audit: 'M12 3l7 3v6c0 4.5-3 7.5-7 9-4-1.5-7-4.5-7-9V6zM9 12l2 2 4-4',
		simulator: 'M4 20h16M6 16l4-6 4 3 4-7M18 6h2v2',
		qr: 'M4 4h6v6H4zM14 4h6v6h-6zM4 14h6v6H4zM14 14h2v2h-2zM18 14h2v2M14 18h2M18 18h2v2h-2z',
		truck: 'M3 6h11v10H3zM14 10h4l3 3v3h-7M7 19a2 2 0 1 0 0-.01M17 19a2 2 0 1 0 0-.01',
		route: 'M6 19a2 2 0 1 0 0-.01M18 7a2 2 0 1 0 0-.01M6 17V9a3 3 0 0 1 3-3h7M18 9v6a3 3 0 0 1-3 3H8',
		carrier: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8M4 21c0-4 4-6 8-6s8 2 8 6',
		fleet: 'M3 8h8v7H3zM11 10h3l2 2v3h-5M13 4h8v7M5 18a1.5 1.5 0 1 0 0-.01M14 18a1.5 1.5 0 1 0 0-.01',
		connect: 'M5 12a2 2 0 1 0 0-.01M19 5a2 2 0 1 0 0-.01M19 19a2 2 0 1 0 0-.01M7 11l10-5M7 13l10 5',
		importer: 'M3 21h18M5 21V8l7-5 7 5v13M9 21v-6h6v6M9 11h.01M15 11h.01',
		billing: 'M3 6h18v12H3zM3 10h18M7 15h3',
		team: 'M9 11a3 3 0 1 0 0-6 3 3 0 0 0 0 6M17 11a2.5 2.5 0 1 0 0-5M3 20c0-3 3-5 6-5s6 2 6 5M16 15c2.5 0 5 1.5 5 4',
		bell: 'M6 16V11a6 6 0 0 1 12 0v5l2 2H4zM10 20a2 2 0 0 0 4 0',
		search: 'M11 18a7 7 0 1 0 0-14 7 7 0 0 0 0 14M20 20l-4-4',
		chevron: 'M9 6l6 6-6 6',
		close: 'M6 6l12 12M18 6L6 18',
		plus: 'M12 5v14M5 12h14',
		upload: 'M12 16V4M7 9l5-5 5 5M4 20h16',
		download: 'M12 4v12M7 11l5 5 5-5M4 20h16',
		check: 'M5 12l5 5 9-10',
		mail: 'M3 6h18v12H3zM3 7l9 6 9-6',
		clock: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18M12 7v5l3 2',
		alert: 'M12 3l10 18H2zM12 10v4M12 17h.01',
		zap: 'M13 2L4 14h7l-1 8 9-12h-7z',
		users: 'M9 11a3 3 0 1 0 0-6 3 3 0 0 0 0 6M3 20c0-3 3-5 6-5s6 2 6 5',
		box: 'M4 8l8-4 8 4v8l-8 4-8-4zM4 8l8 4 8-4M12 12v8',
		eye: 'M2 12s4-7 10-7 10 7 10 7-4 7-10 7S2 12 2 12M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6',
		link: 'M10 14a4 4 0 0 0 6 0l3-3a4 4 0 0 0-6-6l-1 1M14 10a4 4 0 0 0-6 0l-3 3a4 4 0 0 0 6 6l1-1',
		shield: 'M12 3l7 3v6c0 4.5-3 7.5-7 9-4-1.5-7-4.5-7-9V6z',
		key: 'M8 15a4 4 0 1 0 0-8 4 4 0 0 0 0 8M11 11h10M18 11v3M15 11v2',
		whatsapp: 'M4 20l1.5-4A8 8 0 1 1 8 18.5zM9 9c0 3 3 6 6 6l1-1.5-2-1-1 1c-1 0-2.5-1.5-2.5-2.5l1-1-1-2z',
		layers: 'M12 3l9 5-9 5-9-5zM3 13l9 5 9-5',
		settings: 'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6M19 12l2-1-2-4-2 1-2-1V5h-4l-.5 2-2 1-2-1-2 4 2 1v1l-2 1 2 4 2-1 2 1 .5 2h4l.5-2 2-1 2 1 2-4-2-1z',
		logout: 'M15 4h4v16h-4M10 8l-4 4 4 4M6 12h10',
		sparkle: 'M12 3v4M12 17v4M3 12h4M17 12h4M6 6l2.5 2.5M15.5 15.5L18 18M6 18l2.5-2.5M15.5 8.5L18 6'
	};
</script>

<svg class="icon {klass}" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
	<path d={P[name] ?? P.box} />
</svg>

<style>
	.icon {
		flex: none;
	}
</style>
