<script lang="ts">
	import { reveal } from '$design/motion';

	const groups = [
		{
			label: 'Explore now',
			state: 'Interactive demo',
			intro: 'These screens run on sample factory data. You can try the workflows without creating a live customer record.',
			items: [
				['Factory workspace', 'Production, inventory, finance and a factory twin with sample records.'],
				['Carbon workspace', 'Activity capture, CBAM comparisons, XML previews and a calculation simulator.'],
				['Trace & freight', 'Linked QR passports, route matching and proof-of-delivery previews.'],
				['Buyer workspace', 'Supplier discovery, RFQ and importer portfolio previews.']
			]
		},
		{
			label: 'Built for pilot integration',
			state: 'Backend code present',
			intro: 'The repository contains service endpoints and tests. The public workspace is still using sample data while production connections are completed.',
			items: [
				['Tally CSV import', 'Parsing and duplicate protection are implemented; live customer sync needs deployment and validation.'],
				['CBAM and QR services', 'Calculation, export and QR resolver endpoints exist; regulatory and customer-data review remains.'],
				['PGP administrator sign-in', 'The deployed auth service verifies a pinned PGP key; the operations dashboard still shows sample records.']
			]
		},
		{
			label: 'Before customer launch',
			state: 'Not launched',
			intro: 'These require operating processes, partners or verified production infrastructure before ZENRES can offer them to customers.',
			items: [
				['Business verification', 'No live GSTIN, PAN, IEC or representative-verification workflow is connected.'],
				['Payments and subscriptions', 'Prices are illustrative plan information; there is no checkout or live billing.'],
				['Evidence storage and support', 'The demo does not upload documents to a secure vault or run a support-ticket service.'],
				['Marketplace transactions', 'Orders, settlements, inspections and international payments are future work.']
			]
		}
	];
</script>

<svelte:head>
	<title>Features & availability · ZENRES</title>
	<meta name="description" content="See which ZENRES features are interactive demos, which have pilot backend code, and which are still in development." />
</svelte:head>

<div class="container-wide feature-page">
	<header class="intro">
		<span class="eyebrow" use:reveal>Platform / Feature availability</span>
		<h1 use:reveal={{ delay: 80 }}>Explore the product.<br /><span>Know what is live.</span></h1>
		<p use:reveal={{ delay: 160 }}>ZENRES is an early product. The public workspace shows how business records, carbon evidence, traceability and freight could work together. This page marks the current availability of each part.</p>
		<div class="actions" use:reveal={{ delay: 220 }}>
			<a class="btn btn-primary btn-lg" href="https://zenres-app.vercel.app/dashboard">Open sample workspace <span aria-hidden="true">↗</span></a>
			<a class="btn btn-lg" href="/pricing">View proposed plans</a>
		</div>
	</header>

	<div class="groups">
		{#each groups as group, gi (group.label)}
			<section class="group" aria-labelledby={`group-${gi}`}>
				<div class="group-heading" use:reveal>
					<div><span class="eyebrow">0{gi + 1} / {group.label}</span><h2 id={`group-${gi}`}>{group.state}</h2></div>
					<p>{group.intro}</p>
				</div>
				<div class="grid">
					{#each group.items as [title, detail], i (title)}
						<article class="card feature" use:reveal={{ delay: i * 55 }}>
							<span class="number">0{i + 1}</span>
							<h3>{title}</h3>
							<p>{detail}</p>
						</article>
					{/each}
				</div>
			</section>
		{/each}
	</div>
	<p class="footnote">Feature status describes the current ZENRES implementation and can change as the pilot develops. Illustrative estimates and sample records are not customer results.</p>
</div>

<style>
	.feature-page { padding-top: clamp(128px, 15vw, 190px); padding-bottom: 40px; }
	.intro { max-width: 1020px; }
	h1 { margin-top: 20px; font-size: var(--fs-display); letter-spacing: var(--tracking-display); line-height: .98; color: var(--ink-primary); }
	h1 span { color: var(--brand-emerald); }
	.intro > p { max-width: 72ch; margin-top: 30px; font-size: clamp(1rem, 2vw, 1.25rem); line-height: 1.55; }
	.actions { display: flex; gap: 10px; flex-wrap: wrap; margin-top: 30px; }
	.groups { margin-top: clamp(72px, 9vw, 120px); }
	.group { padding-block: 46px 58px; border-top: var(--hairline); }
	.group-heading { display: flex; justify-content: space-between; align-items: end; gap: 30px; margin-bottom: 28px; }
	.group-heading > div { flex: 1; }
	.group-heading h2 { margin-top: 12px; color: var(--ink-primary); font-size: var(--fs-h1); letter-spacing: var(--tracking-display); }
	.group-heading p { max-width: 50ch; font-size: var(--fs-sm); line-height: 1.6; }
	.grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: var(--gap); }
	.feature { min-height: 210px; padding: 22px; display: flex; flex-direction: column; align-items: flex-start; transition: transform 400ms var(--ease-expo), box-shadow 400ms var(--ease-out); }
	.feature:hover { transform: translateY(-5px); box-shadow: 0 14px 38px #53668a12; }
	.number { font: 12px var(--font-mono); color: var(--brand-emerald); }
	.feature h3 { margin-top: 30px; font-size: 1.2rem; color: var(--ink-primary); }
	.feature p { margin-top: 8px; font-size: var(--fs-sm); line-height: 1.55; }
	.footnote { border-top: var(--hairline); padding-top: 24px; font-size: var(--fs-xs); color: var(--ink-muted); }
	@media (max-width: 1000px) { .grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
	@media (max-width: 650px) { .group-heading { display: block; } .group-heading p { margin-top: 12px; } .grid { grid-template-columns: 1fr; } .feature { min-height: 0; } }
</style>
