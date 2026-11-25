<script lang="ts">
	import SplitText from '$design/components/SplitText.svelte';
	import { reveal } from '$design/motion';
	import PricingMatrix from '$lib/components/PricingMatrix.svelte';

	const faqs = [
		['Is pricing in USD?', 'Yes. All contract pricing is in USD, excluding applicable taxes. INR and EUR figures shown on this page are indicative conversions for convenience.'],
		['What does “2 months free” mean?', 'Annual billing is priced at ten times the monthly fee: $990, $2,490 and $4,990 per year respectively.'],
		['Do I lose features when my plan has limits?', 'No. Your workspace is complete for your tier. When your own work reaches a limit — a second plant, a sixth CN-code SKU — Zevra shows a quiet one-line note you can dismiss.'],
		['How does the EU Importer Hub supplier seat work?', '15 suppliers are included. Each additional active supplier is $25/month, and the charge is always shown for confirmation before an invite is sent.'],
		['Is FreightLink transport included in the subscription?', 'No. Freight transport is paid separately. The optional 2% FreightLink commission is illustrative and always disclosed before booking.'],
		['Can transporters join for free?', 'Yes. Basic carrier registration and verification is free, and organic ranking never depends on paid promotion.'],
		['What is the assisted setup sprint?', 'An optional $350 engagement where our team maps your Tally ledgers, meters and CN codes so your first CBAM report is ready faster.']
	];
	let open = $state<number | null>(0);
</script>

<svelte:head>
	<title>Pricing — Zevra</title>
	<meta name="description" content="MSME Starter $99/mo, MSME Growth $249/mo, EU Importer Hub $499/mo. Annual billing includes two months free." />
</svelte:head>

<section class="head container-wide">
	<span class="eyebrow" use:reveal>Pricing</span>
	<SplitText tag="h1" class="p-title" lines={[[{ text: 'Transparent tiers for' }], [{ text: 'every export journey.', serif: true, accent: true }]]} />
	<p class="lede" use:reveal={{ delay: 200 }}>Choose a complete workspace for where your business is today. Move up when your plants, SKUs or supplier network grow — never because a padlock told you to.</p>
</section>

<section class="container-wide body">
	<PricingMatrix expanded />
</section>

<section class="container faq">
	<span class="eyebrow" use:reveal>Questions</span>
	<div class="faq-list">
		{#each faqs as [q, a], i (q)}
			<div class="faq-item" use:reveal={{ delay: i * 40 }}>
				<button aria-expanded={open === i} onclick={() => (open = open === i ? null : i)}>
					<span>{q}</span><span class="plus" class:open={open === i}></span>
				</button>
				<div class="ans" class:open={open === i}><div><p>{a}</p></div></div>
			</div>
		{/each}
	</div>
</section>

<style>
	.head {
		padding-top: 150px;
		padding-bottom: 56px;
		display: flex;
		flex-direction: column;
		gap: 24px;
	}
	:global(.p-title) {
		font-size: var(--fs-display);
		letter-spacing: var(--tracking-display);
		line-height: 0.98;
	}
	.body {
		padding-bottom: 80px;
	}
	.faq {
		max-width: 880px;
		padding-bottom: 40px;
	}
	.faq-list {
		margin-top: 18px;
		border-top: var(--hairline);
	}
	.faq-item {
		border-bottom: var(--hairline);
	}
	.faq-item button {
		width: 100%;
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 16px;
		padding: 22px 0;
		text-align: left;
		font-size: 1.15rem;
		font-weight: 500;
		color: var(--ink-primary);
		letter-spacing: -0.01em;
	}
	.plus {
		position: relative;
		width: 12px;
		height: 12px;
		flex: none;
	}
	.plus::before,
	.plus::after {
		content: '';
		position: absolute;
		left: 0;
		right: 0;
		top: 50%;
		height: 1.5px;
		background: currentColor;
		transition: transform 500ms var(--ease-expo);
	}
	.plus::after {
		transform: rotate(90deg);
	}
	.plus.open::after {
		transform: rotate(0);
	}
	.ans {
		display: grid;
		grid-template-rows: 0fr;
		transition: grid-template-rows 600ms var(--ease-expo);
	}
	.ans.open {
		grid-template-rows: 1fr;
	}
	.ans > div {
		overflow: hidden;
	}
	.ans p {
		padding-bottom: 22px;
		max-width: 64ch;
	}
</style>
