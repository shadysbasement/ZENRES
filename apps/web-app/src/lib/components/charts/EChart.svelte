<script lang="ts">
	import { onMount } from 'svelte';
	import type { EChartsCoreOption, ECharts } from 'echarts/core';

	interface Props {
		option: EChartsCoreOption;
		height?: number;
		class?: string;
	}
	let { option, height = 220, class: klass = '' }: Props = $props();

	let el: HTMLDivElement;
	let chart: ECharts | null = null;

	const theme = {
		color: ['#14533C', '#1E4D6B', '#C05A3C', '#47544F', '#38B884', '#3F7396', '#9AA7A1'],
		backgroundColor: 'transparent',
		textStyle: { fontFamily: 'Switzer, sans-serif', color: '#47544F' },
		grid: { left: 8, right: 8, top: 24, bottom: 4, containLabel: true },
		categoryAxis: {
			axisLine: { lineStyle: { color: '#DCE3DF' } },
			axisTick: { show: false },
			axisLabel: { color: '#6E7D77', fontFamily: 'IBM Plex Mono', fontSize: 10 },
			splitLine: { show: false }
		},
		valueAxis: {
			axisLine: { show: false },
			axisTick: { show: false },
			axisLabel: { color: '#9AA7A1', fontFamily: 'IBM Plex Mono', fontSize: 10 },
			splitLine: { lineStyle: { color: '#EEF2F0', type: 'dashed' } }
		},
		tooltip: {
			backgroundColor: '#0E1412',
			borderWidth: 0,
			padding: [8, 12],
			textStyle: { color: '#F7F8F7', fontFamily: 'IBM Plex Mono', fontSize: 11 },
			extraCssText: 'border-radius:6px;box-shadow:0 12px 32px -12px rgba(14,20,18,.4);'
		},
		legend: { textStyle: { color: '#47544F', fontSize: 11 }, icon: 'rect', itemWidth: 10, itemHeight: 3 }
	};

	onMount(() => {
		let ro: ResizeObserver | undefined;
		let disposed = false;
		(async () => {
			const echarts = await import('echarts/core');
			const [{ LineChart, BarChart, ScatterChart, PieChart }, comps, { SVGRenderer }] = await Promise.all([
				import('echarts/charts'),
				import('echarts/components'),
				import('echarts/renderers')
			]);
			echarts.use([LineChart, BarChart, ScatterChart, PieChart, comps.GridComponent, comps.TooltipComponent, comps.LegendComponent, comps.MarkLineComponent, comps.MarkAreaComponent, comps.DatasetComponent, SVGRenderer]);
			if (disposed) return;
			echarts.registerTheme('zevra', theme);
			chart = echarts.init(el, 'zevra', { renderer: 'svg' });
			chart.setOption({ animationDuration: 1100, animationEasing: 'cubicOut', ...option });
			ro = new ResizeObserver(() => chart?.resize());
			ro.observe(el);
		})();
		return () => {
			disposed = true;
			ro?.disconnect();
			chart?.dispose();
		};
	});

	$effect(() => {
		const o = option;
		chart?.setOption(o, { notMerge: false });
	});
</script>

<div bind:this={el} class="chart {klass}" style="height:{height}px"></div>

<style>
	.chart {
		width: 100%;
	}
</style>
