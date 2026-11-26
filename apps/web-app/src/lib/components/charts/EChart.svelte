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
		color: ['#8CBDE9', '#B6A0DD', '#DFA6AF', '#60718B', '#8CBFAB', '#9CADCE', '#8B9AAF'],
		backgroundColor: 'transparent',
		textStyle: { fontFamily: 'Urbanist Variable, sans-serif', color: '#60718B' },
		grid: { left: 8, right: 8, top: 24, bottom: 4, containLabel: true },
		categoryAxis: {
			axisLine: { lineStyle: { color: '#E3EAF3' } },
			axisTick: { show: false },
			axisLabel: { color: '#7D8CA3', fontFamily: 'Urbanist Variable', fontSize: 10 },
			splitLine: { show: false }
		},
		valueAxis: {
			axisLine: { show: false },
			axisTick: { show: false },
			axisLabel: { color: '#8B9AAF', fontFamily: 'Urbanist Variable', fontSize: 10 },
			splitLine: { lineStyle: { color: '#EEF2F8', type: 'dashed' } }
		},
		tooltip: {
			backgroundColor: '#FCFDFF',
			borderWidth: 0,
			padding: [8, 12],
			textStyle: { color: '#52647E', fontFamily: 'Urbanist Variable', fontSize: 11 },
			extraCssText: 'border-radius:14px;box-shadow:0 8px 32px rgba(70,90,120,.12);border:1px solid white;'
		},
		legend: { textStyle: { color: '#60718B', fontSize: 11 }, icon: 'rect', itemWidth: 10, itemHeight: 3 }
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
			chart.setOption({ animationDuration: 1000, animationDurationUpdate: 650, animationEasing: 'cubicOut', animationEasingUpdate: 'cubicOut', ...option, animation: !window.matchMedia('(prefers-reduced-motion: reduce)').matches });
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
