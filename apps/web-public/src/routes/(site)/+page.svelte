<script lang="ts">
	import Preloader from '$design/components/Preloader.svelte';
	import SplitText from '$design/components/SplitText.svelte';
	import Marquee from '$design/components/Marquee.svelte';
	import CountUp from '$design/components/CountUp.svelte';
	import { reveal, magnetic, parallax, scrollToTarget } from '$design/motion';
	import HeroWorkbench from '$lib/components/HeroWorkbench.svelte';
	import HeroStage from '$lib/three/HeroStage.svelte';
	import PillarTabs from '$lib/components/PillarTabs.svelte';
	import Journey3D from '$lib/components/Journey3D.svelte';
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
		<div class="hero-copy">
			<div class="hero-meta">
				<span class="pill pill-outline live"><span class="dot dot-live text-emerald"></span>EU CBAM definitive regime — live since Jan 2026</span>
			</div>

			<SplitText
				tag="h1"
				class="hero-title"
				manual
				play={ui.ready}
				delay={150}
				lines={[
					[{ text: 'Turn factory records' }],
					[{ text: 'into ' }, { text: 'EU CBAM proof,', accent: true }],
					[{ text: 'verifiable QR passports' }],
					[{ text: '& revenue-generating' }],
					[{ text: 'return freight.', accent: true }]
				]}
			/>

			<p class="lede hero-lede">
				ZENRES connects to Tally.ERP 9 — it never replaces it — and turns your ledgers, utility bills and floor sheets into
				measured carbon intelligence, dual QR traceability and discovery by global B2B buyers.
			</p>
			<div class="hero-ctas">
				<span use:magnetic><a href="/signup" class="btn btn-primary btn-lg" data-cursor="Start">Create Workspace <span class="arrow">→</span></a></span>
				<button class="btn btn-lg btn-glass" onclick={() => scrollToTarget('#workbench', -90)}>Run the CBAM numbers</button>
			</div>
			<dl class="hero-stats">
				<div><dt>Connected pillars</dt><dd><CountUp value={4} /></dd></div>
				<div><dt>Steps, voucher → XML</dt><dd><CountUp value={8} /></dd></div>
				<div><dt>Suppliers per hub</dt><dd class="num">15–500</dd></div>
			</dl>
		</div>
		<div class="hero-stage">
			<HeroStage play={ui.ready} />
		</div>
	</div>
	<button class="scroll-cue" onclick={() => scrollToTarget('#workbench', -90)} aria-label="Scroll to the workbench">
		<span>Scroll</span><i aria-hidden="true"></i>
	</button>
</section>

<!-- ───────────────────────── SECTOR MARQUEE ───────────────────────── -->
<section class="band" aria-label="CBAM sectors">
	<Marquee speed={46}>
		{#each sectors as s (s)}
			<span class="band-item">{s}</span><span class="band-sep" aria-hidden="true">✦</span>
		{/each}
	</Marquee>
</section>

<!-- ───────────────────────── WORKBENCH ───────────────────────── -->
<section class="section container-wide workbench" id="workbench">
	<div class="wb-grid">
		<div class="wb-copy">
			<span class="eyebrow" use:reveal><span class="idx">(00)</span> — Live workbench</span>
			<SplitText tag="h2" class="sh-title" lines={[[{ text: 'Run the numbers ' }], [{ text: 'before you sign up.', accent: true }]]} />
			<p use:reveal={{ delay: 120 }}>
				Slide your annual EU export volume to see how measured emissions change your illustrative CBAM exposure — or switch to
				FreightLink and watch an empty Mumbai return leg get matched.
			</p>
			<ul role="list" class="wb-points" use:reveal={{ delay: 200 }}>
				<li><span></span>Default values vs measured actuals, per CN code</li>
				<li><span></span>Avoided fuel reported separately from CBAM</li>
				<li><span></span>Every figure labelled as an illustrative estimate</li>
			</ul>
		</div>
		<div class="wb-bench" use:reveal={{ delay: 120 }}>
			<HeroWorkbench />
		</div>
	</div>
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
<Journey3D />

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
		min-height: max(760px, 100svh);
		padding-top: 96px;
		padding-bottom: 40px;
		display: flex;
		align-items: center;
		overflow: hidden;
	}
	.hero::before {
		content: '';
		position: absolute;
		inset: 0;
		background:
			radial-gradient(900px 600px at 78% 42%, rgba(164, 212, 247, 0.35), transparent 70%),
			radial-gradient(700px 520px at 92% 78%, rgba(203, 181, 237, 0.32), transparent 70%),
			radial-gradient(600px 400px at 8% 10%, rgba(255, 255, 255, 0.7), transparent 70%);
		pointer-events: none;
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
		mask-image: linear-gradient(180deg, transparent, #000 20%, #000 70%, transparent);
	}
	.hero-grid span {
		border-left: 1px solid rgba(189, 203, 220, 0.45);
		transform: scaleY(0);
		transform-origin: top;
		transition: transform 1.6s var(--ease-expo);
		transition-delay: calc(var(--i) * 45ms);
	}
	.hero-grid span:last-child {
		border-right: 1px solid rgba(189, 203, 220, 0.45);
	}
	.ready .hero-grid span {
		transform: scaleY(1);
	}
	.hero-inner {
		position: relative;
		display: grid;
		grid-template-columns: minmax(0, 6fr) minmax(0, 6fr);
		gap: clamp(16px, 3vw, 48px);
		align-items: center;
	}
	.hero-copy {
		position: relative;
		z-index: 2;
		display: flex;
		flex-direction: column;
		gap: 26px;
	}
	.hero-meta {
		opacity: 0;
		transform: translateY(12px);
		transition: opacity 1s var(--ease-expo) 0.1s, transform 1s var(--ease-expo) 0.1s;
	}
	.ready .hero-meta {
		opacity: 1;
		transform: none;
	}
	.hero-meta .live {
		height: 30px;
		padding: 0 14px;
		background: rgba(252, 253, 255, 0.7);
		border: 1px solid rgba(255, 255, 255, 0.95);
		box-shadow: 0 6px 20px -12px rgba(67, 92, 140, 0.4);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		text-transform: none;
		letter-spacing: 0;
		font-size: 12.5px;
		font-weight: 500;
		color: var(--ink-secondary);
	}
	:global(.hero-title) {
		font-size: clamp(2.4rem, 3.7vw, 4.1rem);
		font-weight: 600;
		letter-spacing: -0.045em;
		line-height: 1;
		color: var(--ink-primary);
	}
	:global(.hero-title .accent),
	:global(.sh-title .accent),
	:global(.cta-title .accent) {
		background: linear-gradient(100deg, #436c9b 0%, #6a8fd6 45%, #8265a8 100%);
		-webkit-background-clip: text;
		background-clip: text;
		color: transparent;
		padding-right: 0.04em;
	}
	.hero-lede {
		max-width: 46ch;
		font-size: 1.08rem;
	}
	.hero-lede,
	.hero-ctas,
	.hero-stats {
		opacity: 0;
		transform: translateY(20px);
		transition: opacity 1.2s var(--ease-expo), transform 1.2s var(--ease-expo);
	}
	.hero-lede {
		transition-delay: 0.55s;
	}
	.hero-ctas {
		transition-delay: 0.68s;
	}
	.hero-stats {
		transition-delay: 0.8s;
	}
	.ready .hero-lede,
	.ready .hero-ctas,
	.ready .hero-stats {
		opacity: 1;
		transform: none;
	}
	.hero-ctas {
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
	}
	:global(.btn-glass) {
		--btn-bg: rgba(252, 253, 255, 0.62);
		--btn-bd: rgba(255, 255, 255, 0.95);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		box-shadow: 0 8px 24px -16px rgba(67, 92, 140, 0.5);
	}
	.hero-stats {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.hero-stats > div {
		display: flex;
		flex-direction: column-reverse;
		gap: 2px;
		padding: 10px 16px;
		border-radius: var(--radius-md);
		background: rgba(252, 253, 255, 0.55);
		border: 1px solid rgba(255, 255, 255, 0.9);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
	}
	.hero-stats dt {
		font-size: 11.5px;
		color: var(--ink-muted);
	}
	.hero-stats dd {
		font-size: 1.45rem;
		font-weight: 600;
		letter-spacing: -0.03em;
		color: var(--ink-primary);
	}
	.hero-stage {
		position: relative;
		height: clamp(520px, 74vh, 780px);
		margin-right: calc(var(--gutter) * -1);
	}
	.scroll-cue {
		position: absolute;
		left: 50%;
		bottom: 18px;
		transform: translateX(-50%);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--ink-muted);
		opacity: 0;
		transition: opacity 1s var(--ease-out) 1.4s;
	}
	.ready .scroll-cue {
		opacity: 1;
	}
	.scroll-cue i {
		position: relative;
		width: 1px;
		height: 34px;
		background: rgba(104, 117, 133, 0.25);
		overflow: hidden;
	}
	.scroll-cue i::after {
		content: '';
		position: absolute;
		left: 0;
		top: -40%;
		width: 1px;
		height: 40%;
		background: var(--brand-emerald);
		animation: cue 1.8s var(--ease-in-out) infinite;
	}
	@keyframes cue {
		to {
			top: 100%;
		}
	}

	/* WORKBENCH */
	.wb-grid {
		display: grid;
		grid-template-columns: 5fr 7fr;
		gap: clamp(32px, 5vw, 80px);
		align-items: center;
	}
	.wb-copy {
		display: flex;
		flex-direction: column;
		gap: 22px;
	}
	.wb-copy .idx {
		color: var(--brand-emerald);
	}
	.wb-copy :global(.sh-title) {
		font-size: clamp(2.2rem, 3.5vw, 3.6rem);
		max-width: none;
	}
	.wb-copy p {
		max-width: 44ch;
		font-size: 1.02rem;
	}
	.wb-points {
		display: flex;
		flex-direction: column;
		gap: 10px;
		font-size: 14px;
		color: var(--ink-primary);
	}
	.wb-points li {
		display: flex;
		gap: 10px;
		align-items: center;
	}
	.wb-points span {
		width: 18px;
		height: 18px;
		flex: none;
		border-radius: 50%;
		background:
			url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='9' height='7' fill='none'%3E%3Cpath stroke='%2332435c' stroke-width='1.6' d='m1 3.5 2.2 2.2L8 1'/%3E%3C/svg%3E") center no-repeat,
			linear-gradient(135deg, #a4d4f7, #cbb5ed);
	}
	.wb-bench :global(.wb) {
		box-shadow: 0 40px 90px -40px rgba(67, 92, 140, 0.55);
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
		.hero-inner,
		.wb-grid {
			grid-template-columns: 1fr;
		}
		.hero {
			min-height: 0;
			padding-top: 112px;
		}
		.hero-stage {
			height: clamp(420px, 62vh, 600px);
			margin: 0 calc(var(--gutter) * -1);
		}
		.scroll-cue {
			display: none;
		}
	}
	@media (max-width: 900px) {
		.cbam-grid,
		.trust {
			grid-template-columns: 1fr;
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
