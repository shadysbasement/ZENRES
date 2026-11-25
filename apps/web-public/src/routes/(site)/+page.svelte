<script lang="ts">
	import Preloader from '$design/components/Preloader.svelte';
	import SplitText from '$design/components/SplitText.svelte';
	import Marquee from '$design/components/Marquee.svelte';
	import CountUp from '$design/components/CountUp.svelte';
	import { reveal, magnetic, parallax, scrollToTarget } from '$design/motion';
	import HeroWorkbench from '$lib/components/HeroWorkbench.svelte';
	import PillarTabs from '$lib/components/PillarTabs.svelte';
	import JourneyStrip from '$lib/components/JourneyStrip.svelte';
	import FreightCorridor from '$lib/components/FreightCorridor.svelte';
	import PricingMatrix from '$lib/components/PricingMatrix.svelte';
	import QrPassportDemo from '$lib/components/QrPassportDemo.svelte';
	import SectionHead from '$lib/components/SectionHead.svelte';
	import ScrollText from '$lib/components/ScrollText.svelte';
	import { ui } from '$lib/stores/ui.svelte';
	import { DISCLAIMERS } from '$lib/content';

	const sectors = ['Iron & Steel', 'Aluminium', 'Cement', 'Fertilisers', 'Hydrogen', 'General Manufacturing'];

	const quality = [
		{ name: 'Measured', cls: 'pill-emerald', body: 'From meters, bills and production records captured inside ZENRES.' },
		{ name: 'Supplier-Reported', cls: 'pill-cobalt', body: 'Declared by an upstream supplier with linked evidence.' },
		{ name: 'Estimated', cls: 'pill-terracotta', body: 'Modelled or default values — clearly flagged until replaced.' },
		{ name: 'Independently Verified', cls: 'pill-dark', body: 'Signed off by an accredited verifier or CA in the audit workspace.' }
	];

	const trust = [
		['Tenant isolation', 'PostgreSQL row-level security on every tenant-scoped table. Suppliers never see each other.'],
		['Argon2id credentials', 'Memory-hard password hashing, mandatory email verification and optional TOTP.'],
		['Passkeys & Proton', 'WebAuthn passkeys tuned for Proton Pass, with PGP-encrypted login links via WKD.'],
		['Tamper-evident audit', 'SHA-256 hash-chained evidence trail with selective off-chain anchoring.']
	];
</script>

<svelte:head>
	<title>ZENRES — CBAM proof, QR passports & return freight for Indian manufacturers</title>
	<meta name="description" content="ZENRES turns Tally records, meters and floor sheets into audit-ready EU CBAM evidence, verifiable dual-QR passports and revenue-generating backhaul freight." />
</svelte:head>

<Preloader ondone={() => (ui.ready = true)} />

<!-- ───────────────────────── HERO ───────────────────────── -->
<section class="hero" class:ready={ui.ready}>
	<div class="hero-grid" aria-hidden="true">
		{#each Array(12) as _, i (i)}<span style="--i:{i}"></span>{/each}
	</div>
	<div class="container-wide hero-inner">
		<div class="hero-meta">
			<span class="pill pill-outline"><span class="dot dot-live text-emerald"></span>EU CBAM definitive regime — live since Jan 2026</span>
			<span class="eyebrow meta-r">For Indian MSMEs, transporters & EU importers</span>
		</div>

		<SplitText
			tag="h1"
			class="hero-title"
			manual
			play={ui.ready}
			delay={150}
			lines={[
				[{ text: 'Turn factory records into ' }, { text: 'EU CBAM proof,', serif: true }],
				[{ text: 'verifiable QR passports, and' }],
				[{ text: 'revenue-generating ' }, { text: 'return freight.', serif: true, accent: true }]
			]}
		/>

		<div class="hero-lower">
			<div class="hero-copy">
				<p class="lede">
					ZENRES connects to Tally.ERP 9 — it never replaces it — and turns your ledgers, utility bills and floor sheets into
					measured carbon intelligence, dual QR traceability and discovery by global B2B buyers.
				</p>
				<div class="hero-ctas">
					<span use:magnetic><a href="/signup" class="btn btn-primary btn-lg" data-cursor="Start">Create Workspace <span class="arrow">→</span></a></span>
					<button class="btn btn-lg" onclick={() => scrollToTarget('#workbench', -100)}>Inspect Sample CBAM & Freight Flow</button>
				</div>
				<dl class="hero-stats">
					<div><dt class="eyebrow">Pillars</dt><dd><CountUp value={4} /></dd></div>
					<div><dt class="eyebrow">Journey steps</dt><dd><CountUp value={8} /></dd></div>
					<div><dt class="eyebrow">Supplier portal</dt><dd class="num">15–500</dd></div>
				</dl>
			</div>
			<div class="hero-bench" id="workbench">
				<HeroWorkbench />
			</div>
		</div>
	</div>
</section>

<!-- ───────────────────────── SECTOR MARQUEE ───────────────────────── -->
<section class="band" aria-label="CBAM sectors">
	<Marquee speed={46}>
		{#each sectors as s (s)}
			<span class="band-item">{s}</span><span class="band-sep" aria-hidden="true">✦</span>
		{/each}
	</Marquee>
</section>

<!-- ───────────────────────── MANIFESTO ───────────────────────── -->
<section class="manifesto container">
	<span class="eyebrow" use:reveal>(00) — Why ZENRES</span>
	<ScrollText
		text="Indian factories already keep meticulous records. The EU now asks for *proof.* ZENRES turns the ledgers, meters and floor sheets you trust into evidence your importer can file — and fills the empty truck on the way *home.*"
	/>
</section>

<!-- ───────────────────────── PILLARS ───────────────────────── -->
<section class="section container-wide" id="platform">
	<SectionHead index="01" label="The four connected pillars" title="One ledger." serif="Four connected pillars.">
		Business records, carbon accounting, traceability and buyer discovery share one tenant-isolated data model — so every number
		on a CBAM report can be traced back to a voucher, a meter or a batch.
	</SectionHead>
	<PillarTabs />
</section>

<!-- ───────────────────────── CBAM ───────────────────────── -->
<section class="section container-wide" id="cbam">
	<SectionHead index="02" label="CBAM & carbon intelligence" title="Default values cost more." serif="Measured data costs less.">
		Under the definitive regime, embedded emissions without verified actual data fall back to default values with date-effective
		markups. ZENRES makes measured data the path of least resistance.
	</SectionHead>
	<div class="cbam-grid">
		<div class="cbam-card card" use:reveal>
			<span class="eyebrow">Data-quality labels on every figure</span>
			<ul role="list" class="quality">
				{#each quality as q, i (q.name)}
					<li use:reveal={{ delay: i * 80 }}>
						<span class="pill {q.cls}">{q.name}</span>
						<p>{q.body}</p>
					</li>
				{/each}
			</ul>
		</div>
		<div class="cbam-side">
			<div class="cbam-stat card" use:reveal={{ delay: 80 }}>
				<span class="eyebrow">Rule versions</span>
				<p class="big-serif serif">Date-effective</p>
				<p>Every calculation is pinned to the EU CBAM rule set in force for its period — including IR (EU) 2025/2621 markup schedules.</p>
			</div>
			<div class="cbam-stat card dark" use:reveal={{ delay: 160 }}>
				<span class="eyebrow">Strict separation</span>
				<p class="big-serif serif">Logistics ≠ embedded</p>
				<p>FreightLink avoided-fuel estimates are reported separately and never netted against statutory CBAM embedded emissions.</p>
			</div>
		</div>
	</div>
	<p class="section-note">{DISCLAIMERS.reporting}</p>
</section>

<!-- ───────────────────────── JOURNEY ───────────────────────── -->
<JourneyStrip />

<!-- ───────────────────────── FREIGHTLINK ───────────────────────── -->
<section class="section container-wide" id="freightlink">
	<SectionHead index="04" label="FreightLink backhaul" title="Zero empty" serif="return kilometres.">
		Every outbound export truck comes home empty. FreightLink lists that return capacity, matches it to load demand on the same
		corridor and closes the loop with QR pickup and a digital POD.
	</SectionHead>
	<FreightCorridor />
</section>

<!-- ───────────────────────── PRICING ───────────────────────── -->
<section class="section container-wide" id="pricing">
	<SectionHead index="05" label="Transparent pricing" title="Three tiers." serif="No hidden walls.">
		Each workspace is complete for its tier. Higher-tier capabilities never appear as padlocks — only as a quiet suggestion when
		your own work calls for them.
	</SectionHead>
	<PricingMatrix />
</section>

<!-- ───────────────────────── TRACEABILITY ───────────────────────── -->
<section class="section container-wide" id="traceability">
	<SectionHead index="06" label="Dual QR passports" title="One container." serif="Every batch inside it.">
		Container ZV-001 links to production batches A-101, A-102 and A-103. Each scanner sees only what their role permits.
	</SectionHead>
	<QrPassportDemo />
</section>

<!-- ───────────────────────── TRUST ───────────────────────── -->
<section class="section container-wide trust">
	<div class="trust-grid">
		{#each trust as [title, body], i (title)}
			<div class="trust-item" use:reveal={{ delay: i * 90 }}>
				<span class="mono idx">0{i + 1}</span>
				<h3>{title}</h3>
				<p>{body}</p>
			</div>
		{/each}
	</div>
	<div class="trust-visual" aria-hidden="true">
		<div class="tv-inner" use:parallax={0.06}>
			<span class="mono">sha256 · 9f2c…e41a</span>
			<span class="mono">rls · org_id = current_setting('app.current_org_id')</span>
			<span class="mono">argon2id · t=3 m=64MiB p=4</span>
			<span class="mono">webauthn · proton pass</span>
		</div>
	</div>
</section>

<style>
	/* HERO */
	.hero {
		position: relative;
		padding-top: 128px;
		padding-bottom: 72px;
		overflow: hidden;
	}
	.hero-grid {
		position: absolute;
		inset: 0;
		display: grid;
		grid-template-columns: repeat(12, 1fr);
		max-width: calc(var(--max-w-wide) + var(--gutter) * 2);
		margin: 0 auto;
		padding-inline: var(--gutter);
		pointer-events: none;
	}
	.hero-grid span {
		border-left: 1px solid var(--border-hairline);
		opacity: 0.55;
		transform: scaleY(0);
		transform-origin: top;
		transition: transform 1.6s var(--ease-expo);
		transition-delay: calc(var(--i) * 45ms);
	}
	.hero-grid span:last-child {
		border-right: 1px solid var(--border-hairline);
	}
	.ready .hero-grid span {
		transform: scaleY(1);
	}
	.hero-inner {
		position: relative;
	}
	.hero-meta {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 16px;
		margin-bottom: 36px;
		opacity: 0;
		transform: translateY(12px);
		transition: opacity 1s var(--ease-expo) 0.1s, transform 1s var(--ease-expo) 0.1s;
	}
	.ready .hero-meta {
		opacity: 1;
		transform: none;
	}
	.hero-meta .pill {
		height: 28px;
		padding: 0 12px;
		background: var(--surface-card);
		text-transform: none;
		letter-spacing: 0;
		font-size: 12px;
		color: var(--ink-secondary);
	}
	:global(.hero-title) {
		font-size: clamp(2.6rem, 6.4vw, 6.6rem);
		font-weight: 500;
		letter-spacing: -0.042em;
		line-height: 0.98;
		color: var(--ink-primary);
	}
	:global(.hero-title .serif) {
		letter-spacing: -0.03em;
	}
	.hero-lower {
		display: grid;
		grid-template-columns: 5fr 7fr;
		gap: clamp(32px, 5vw, 80px);
		margin-top: 56px;
		align-items: start;
	}
	.hero-copy {
		display: flex;
		flex-direction: column;
		gap: 28px;
		padding-top: 8px;
	}
	.hero-copy,
	.hero-bench {
		opacity: 0;
		transform: translateY(28px);
		transition: opacity 1.2s var(--ease-expo), transform 1.2s var(--ease-expo);
	}
	.hero-copy {
		transition-delay: 0.55s;
	}
	.hero-bench {
		transition-delay: 0.75s;
	}
	.ready .hero-copy,
	.ready .hero-bench {
		opacity: 1;
		transform: none;
	}
	.hero-ctas {
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
	}
	.hero-stats {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		border-top: var(--hairline);
		padding-top: 20px;
		margin-top: 8px;
	}
	.hero-stats dd {
		margin-top: 6px;
		font-family: var(--font-mono);
		font-size: 1.9rem;
		color: var(--ink-primary);
		letter-spacing: -0.04em;
	}
	.hero-bench {
		scroll-margin-top: 100px;
	}

	/* BAND */
	.band {
		border-block: var(--hairline);
		padding-block: 26px;
		background: var(--surface-card);
	}
	.band-item {
		font-size: clamp(1.75rem, 3.4vw, 3rem);
		font-weight: 500;
		letter-spacing: -0.03em;
		color: var(--ink-primary);
		padding-inline: 28px;
		white-space: nowrap;
	}
	.band-item:nth-child(4n + 1) {
		font-family: var(--font-serif);
		font-style: italic;
		font-weight: 400;
	}
	.band-sep {
		color: var(--brand-emerald);
		font-size: 1.1rem;
	}

	/* MANIFESTO */
	.manifesto {
		padding-block: clamp(100px, 14vw, 200px);
		display: flex;
		flex-direction: column;
		gap: 28px;
		max-width: 1120px;
	}

	/* SECTIONS */
	.section {
		padding-block: clamp(80px, 10vw, 140px);
		scroll-margin-top: 40px;
	}
	.section-note {
		margin-top: 20px;
		font-size: 11.5px;
		color: var(--ink-muted);
	}

	/* CBAM */
	.cbam-grid {
		display: grid;
		grid-template-columns: 7fr 5fr;
		gap: var(--gap);
	}
	.cbam-card {
		padding: 32px;
	}
	.quality {
		margin-top: 22px;
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1px;
		background: var(--border-hairline);
		border: var(--hairline);
		border-radius: var(--radius-sm);
		overflow: hidden;
	}
	.quality li {
		background: var(--surface-card);
		padding: 22px;
		display: flex;
		flex-direction: column;
		gap: 12px;
		align-items: flex-start;
		min-height: 150px;
	}
	.quality p {
		font-size: var(--fs-sm);
		color: var(--ink-secondary);
	}
	.cbam-side {
		display: flex;
		flex-direction: column;
		gap: var(--gap);
	}
	.cbam-stat {
		flex: 1;
		padding: 28px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.cbam-stat p:not(.big-serif) {
		font-size: var(--fs-sm);
	}
	.big-serif {
		font-size: clamp(1.9rem, 3vw, 2.6rem);
		color: var(--ink-primary);
		line-height: 1.05;
	}
	.cbam-stat.dark {
		background: var(--obsidian);
		border-color: var(--obsidian);
		color: #a9b6b0;
	}
	.cbam-stat.dark .big-serif {
		color: var(--vault-phosphor);
	}
	.cbam-stat.dark .eyebrow {
		color: var(--vault-dim);
	}

	/* TRUST */
	.trust {
		display: grid;
		grid-template-columns: 8fr 4fr;
		gap: var(--gap);
		align-items: stretch;
	}
	.trust-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		border-top: var(--hairline);
	}
	.trust-item {
		padding: 28px 28px 28px 0;
		border-bottom: var(--hairline);
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.trust-item:nth-child(odd) {
		border-right: var(--hairline);
	}
	.trust-item:nth-child(even) {
		padding-left: 28px;
	}
	.trust-item .idx {
		font-size: 12px;
		color: var(--brand-emerald);
	}
	.trust-item h3 {
		font-size: 1.35rem;
	}
	.trust-item p {
		font-size: var(--fs-sm);
	}
	.trust-visual {
		position: relative;
		background: var(--obsidian);
		border-radius: var(--radius-md);
		overflow: hidden;
		min-height: 280px;
		background-image: linear-gradient(var(--obsidian-line) 1px, transparent 1px), linear-gradient(90deg, var(--obsidian-line) 1px, transparent 1px);
		background-size: 28px 28px;
	}
	.tv-inner {
		position: absolute;
		inset: 0;
		display: flex;
		flex-direction: column;
		justify-content: center;
		gap: 14px;
		padding: 28px;
	}
	.tv-inner span {
		font-size: 12px;
		color: var(--vault-phosphor);
		background: var(--vault-bg);
		border: 1px solid var(--vault-border);
		padding: 8px 10px;
		border-radius: 3px;
		width: fit-content;
	}
	.tv-inner span:nth-child(2) {
		margin-left: 18px;
		color: #8fb4a4;
	}
	.tv-inner span:nth-child(3) {
		margin-left: 8px;
	}

	@media (max-width: 1100px) {
		.hero-lower {
			grid-template-columns: 1fr;
		}
	}
	@media (max-width: 900px) {
		.cbam-grid,
		.trust {
			grid-template-columns: 1fr;
		}
		.meta-r {
			display: none;
		}
	}
	@media (max-width: 600px) {
		.quality,
		.trust-grid {
			grid-template-columns: 1fr;
		}
		.trust-item:nth-child(odd) {
			border-right: 0;
		}
		.trust-item:nth-child(even) {
			padding-left: 0;
		}
		.hero {
			padding-top: 104px;
		}
	}
</style>
