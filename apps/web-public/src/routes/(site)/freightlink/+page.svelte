<script lang="ts">
	import SplitText from '$design/components/SplitText.svelte';
	import CountUp from '$design/components/CountUp.svelte';
	import { reveal, magnetic } from '$design/motion';
	import FreightCorridor from '$lib/components/FreightCorridor.svelte';
	import { DISCLAIMERS } from '$lib/content';

	const onboarding = [
		['Business identity', 'Legal name, GSTIN, contact person and service corridors.'],
		['Vehicle registration (RC)', 'Registration number, vehicle type and rated payload per truck.'],
		['Insurance & permits', 'Valid insurance, national and state permits with expiry tracking.'],
		['Document verification', 'A checklist reviewed by Zevra — the verified badge is never for sale.']
	];
	const rules = [
		['Never auto-booked', 'Every booking requires explicit, authorised acceptance of an itemised quote.'],
		['Disclosed fees', 'The optional 2% commission is shown before booking — never buried in the freight rate.'],
		['Consent-based GPS', 'Live location is shared only when the carrier explicitly opts in for that trip.'],
		['No pay-to-rank', 'Basic verification and organic ranking never depend on paid promotion.']
	];
</script>

<svelte:head>
	<title>FreightLink — Return-load backhaul matching · Zevra</title>
	<meta name="description" content="FreightLink matches empty return truck capacity to load demand on Delhi–Mumbai and Ludhiana–Delhi corridors, with QR pickup and digital POD." />
</svelte:head>

<section class="hero container-wide">
	<span class="eyebrow" use:reveal>FreightLink — Zevra Trace</span>
	<SplitText tag="h1" class="f-title" lines={[[{ text: 'Every return leg' }], [{ text: 'is a load waiting', serif: true }], [{ text: 'to be matched.', serif: true, accent: true }]]} />
	<div class="hero-row">
		<p class="lede" use:reveal={{ delay: 200 }}>Transporters list empty return capacity for free. Shippers post load demand on the same corridor. Zevra’s deterministic matcher applies hard filters first, then ranks by detour, availability, quote and reliability.</p>
		<div class="ctas" use:reveal={{ delay: 300 }}>
			<span use:magnetic><a class="btn btn-primary btn-lg" href="/signup?role=transporter">Register as transporter — free</a></span>
			<a class="btn btn-lg" href="/signup?plan=growth">Ship with FreightLink</a>
		</div>
	</div>
	<dl class="stats">
		<div use:reveal><dt class="eyebrow">Pilot corridors</dt><dd><CountUp value={2} /></dd></div>
		<div use:reveal={{ delay: 80 }}><dt class="eyebrow">Journey stages</dt><dd><CountUp value={8} /></dd></div>
		<div use:reveal={{ delay: 160 }}><dt class="eyebrow">Carrier registration</dt><dd class="num">$0</dd></div>
		<div use:reveal={{ delay: 240 }}><dt class="eyebrow">Illustrative commission</dt><dd><CountUp value={2} suffix="%" /></dd></div>
	</dl>
</section>

<section class="container-wide section" id="corridors">
	<FreightCorridor />
</section>

<section class="container-wide section split">
	<div>
		<span class="eyebrow" use:reveal>Transporter onboarding</span>
		<h2 use:reveal={{ delay: 80 }}>Free to join. <span class="serif">Verified to win.</span></h2>
	</div>
	<ol class="steps" role="list">
		{#each onboarding as [t, b], i (t)}
			<li use:reveal={{ delay: i * 80 }}><span class="mono">0{i + 1}</span><div><h3>{t}</h3><p>{b}</p></div></li>
		{/each}
	</ol>
</section>

<section class="container-wide section">
	<div class="rules">
		{#each rules as [t, b], i (t)}
			<div class="rule card" use:reveal={{ delay: i * 80 }}>
				<span class="pill pill-cobalt">Integrity rule 0{i + 1}</span>
				<h3>{t}</h3>
				<p>{b}</p>
			</div>
		{/each}
	</div>
	<p class="fine">{DISCLAIMERS.freight} {DISCLAIMERS.commercial}</p>
</section>

<style>
	.hero {
		padding-top: 150px;
		padding-bottom: 40px;
		display: flex;
		flex-direction: column;
		gap: 28px;
	}
	:global(.f-title) {
		font-size: var(--fs-display);
		letter-spacing: var(--tracking-display);
		line-height: 0.96;
	}
	.hero-row {
		display: flex;
		justify-content: space-between;
		align-items: flex-end;
		gap: 32px;
		flex-wrap: wrap;
	}
	.ctas {
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
	}
	.stats {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		border-top: var(--hairline);
		margin-top: 20px;
	}
	.stats > div {
		padding: 22px 0;
		border-right: var(--hairline);
		padding-left: 20px;
	}
	.stats > div:first-child {
		padding-left: 0;
	}
	.stats > div:last-child {
		border-right: 0;
	}
	.stats dd {
		margin-top: 8px;
		font-family: var(--font-mono);
		font-size: 2.6rem;
		letter-spacing: -0.04em;
		color: var(--ink-primary);
	}
	.section {
		padding-block: 80px;
	}
	.split {
		display: grid;
		grid-template-columns: 5fr 7fr;
		gap: var(--gap);
	}
	.split h2 {
		margin-top: 16px;
		font-size: var(--fs-h1);
		letter-spacing: var(--tracking-display);
	}
	.split h2 .serif {
		color: var(--brand-emerald);
	}
	.steps {
		border-top: var(--hairline);
	}
	.steps li {
		display: grid;
		grid-template-columns: 60px 1fr;
		padding: 24px 0;
		border-bottom: var(--hairline);
	}
	.steps .mono {
		color: var(--brand-emerald);
		font-size: 13px;
	}
	.steps h3 {
		font-size: 1.3rem;
	}
	.steps p {
		margin-top: 6px;
		font-size: var(--fs-sm);
	}
	.rules {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: var(--gap);
	}
	.rule {
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 14px;
		align-items: flex-start;
	}
	.rule h3 {
		font-size: 1.25rem;
	}
	.rule p {
		font-size: var(--fs-sm);
	}
	.fine {
		margin-top: 20px;
		font-size: 11.5px;
		color: var(--ink-muted);
		max-width: 90ch;
	}
	@media (max-width: 1000px) {
		.rules {
			grid-template-columns: 1fr 1fr;
		}
		.split {
			grid-template-columns: 1fr;
		}
		.stats {
			grid-template-columns: 1fr 1fr;
		}
		.stats > div:nth-child(2) {
			border-right: 0;
		}
		.stats > div:nth-child(3) {
			padding-left: 0;
		}
	}
	@media (max-width: 560px) {
		.rules {
			grid-template-columns: 1fr;
		}
	}
</style>
