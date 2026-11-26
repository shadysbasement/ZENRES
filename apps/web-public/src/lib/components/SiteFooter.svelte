<script lang="ts">
	import Logo from '$design/components/Logo.svelte';
	import { reveal, magnetic } from '$design/motion';
	import SplitText from '$design/components/SplitText.svelte';
	import { DISCLAIMERS } from '$lib/content';
	import { LEGAL_LINKS } from '$lib/legal/policies';
	let { legalAvailable = false }: { legalAvailable?: boolean } = $props();

	const columns = [
		{ title: 'Platform', links: [['Features & availability', '/features'], ['ZENRES Business', '/#platform'], ['ZENRES Carbon', '/#cbam'], ['ZENRES Trace', '/#traceability'], ['ZENRES Connect', '/#platform']] },
		{ title: 'Logistics', links: [['FreightLink', '/freightlink'], ['Delhi ↔ Mumbai', '/freightlink#corridors'], ['Ludhiana ↔ Delhi', '/freightlink#corridors'], ['Free carrier onboarding', '/signup?role=transporter']] },
		{ title: 'Commercial', links: [['Pricing', '/pricing'], ['Plan comparison', '/pricing#compare'], ['EU Importer Hub', '/pricing#importer_hub'], ['Verify a QR passport', '/verify/c/01JD8ZV001CNTR7Q4M2X9KPWBA']] },
		{ title: 'Account', links: [['Sign in', '/login'], ['Create workspace', '/signup'], ['Continue with Proton', '/login?method=proton'], ['Support centre', '/support'], ['Service status', '/status'], ['Security', '/#security']] }
	];

	let clicks: number[] = [];
	function buildClick(e: MouseEvent) {
		if (!e.shiftKey) {
			clicks = [];
			return;
		}
		const now = performance.now();
		clicks = [...clicks.filter((t) => now - t < 2000), now];
		if (clicks.length >= 5) {
			clicks = [];
			window.dispatchEvent(new CustomEvent('zv:diag'));
		}
	}
</script>

<footer class="footer">
	<section class="cta container-wide">
		<div class="cta-inner">
			<span class="eyebrow" use:reveal>(Next step)</span>
			<SplitText
				tag="h2"
				class="cta-title"
				lines={[[{ text: 'Be ready before your' }], [{ text: 'EU importer asks.', serif: true }]]}
			/>
			<div class="cta-actions" use:reveal={{ delay: 200 }}>
				<span use:magnetic><a href="/signup" class="btn btn-primary btn-lg" data-cursor="Start">Create Workspace <span class="arrow">→</span></a></span>
				<a href="/pricing" class="btn btn-lg">Compare plans</a>
			</div>
		</div>
	</section>

	<div class="container-wide">
		<div class="cols">
			<div class="brandcol">
				<Logo size={28} inverse />
				<p>Zero Emission Vision, Responsible Advancement. CBAM evidence, dual-QR traceability and return freight for Indian manufacturers and the EU importers who buy from them.</p>
			</div>
			{#each columns as col (col.title)}
				<div class="col">
					<span class="eyebrow">{col.title}</span>
					<ul role="list">
						{#each col.links as [label, href] (label)}
							<li><a {href} class="link-u">{label}</a></li>
						{/each}
					</ul>
				</div>
			{/each}
		</div>

		<div class="notes" id="security">
			<p><span class="mono">[1]</span> {DISCLAIMERS.reporting}</p>
			<p><span class="mono">[2]</span> {DISCLAIMERS.buyers}</p>
			<p><span class="mono">[3]</span> {DISCLAIMERS.anchoring}</p>
			<p><span class="mono">[4]</span> {DISCLAIMERS.commercial}</p>
		</div>

		<div class="giant" aria-hidden="true">
			{#each 'ZENRES'.split('') as ch, i (i)}<span style="--i:{i}" use:reveal={{ delay: i * 60 }}>{ch}</span>{/each}
		</div>

		<div class="bar">
			<span>© 2026 ZENRES. All rights reserved.</span>
			<span class="build" onclick={buildClick} role="presentation">BUILD 2026.11 // ZV-CORE</span>
		</div>
		{#if legalAvailable}<nav class="legal-links" aria-label="Legal information">{#each LEGAL_LINKS as link (link.slug)}<a href="/legal/{link.slug}">{link.label}</a>{/each}</nav>{/if}
	</div>
</footer>

<style>
	.footer {
		background: var(--obsidian);
		color: #a9b6b0;
		margin-top: 120px;
		overflow: hidden;
	}
	.cta {
		padding-block: clamp(96px, 14vw, 180px) clamp(72px, 9vw, 120px);
		border-bottom: 1px solid var(--obsidian-line);
	}
	.cta-inner {
		display: flex;
		flex-direction: column;
		gap: 28px;
	}
	.cta .eyebrow {
		color: var(--vault-phosphor);
	}
	:global(.cta-title) {
		color: var(--ink-inverse) !important;
		font-size: var(--fs-display);
		font-weight: 500;
		letter-spacing: var(--tracking-display);
		line-height: 0.98;
	}
	:global(.cta-title .serif) {
		color: var(--vault-phosphor);
	}
	.cta-actions {
		display: flex;
		gap: 12px;
		flex-wrap: wrap;
		margin-top: 12px;
	}
	.cta-actions .btn:not(.btn-primary) {
		--btn-bg: transparent;
		--btn-fg: var(--ink-inverse);
		--btn-bd: var(--obsidian-line);
	}
	.cta-actions .btn:not(.btn-primary):hover {
		--btn-bd: var(--vault-dim);
	}
	.cols {
		display: grid;
		grid-template-columns: 1.6fr repeat(4, 1fr);
		gap: 40px;
		padding-block: 72px 56px;
	}
	.brandcol p {
		margin-top: 20px;
		max-width: 34ch;
		font-size: var(--fs-sm);
		line-height: 1.6;
	}
	.col .eyebrow {
		color: var(--vault-dim);
	}
	.col ul {
		margin-top: 16px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.col a {
		font-size: var(--fs-sm);
		color: #d4ddd9;
	}
	.notes {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 12px 40px;
		padding-block: 28px;
		border-top: 1px solid var(--obsidian-line);
		font-size: var(--fs-xs);
		line-height: 1.6;
		color: #7f8f88;
	}
	.notes .mono {
		color: var(--vault-phosphor);
		margin-right: 6px;
	}
	.giant {
		display: flex;
		justify-content: space-between;
		font-weight: 600;
		font-size: clamp(6rem, 25vw, 26rem);
		letter-spacing: -0.05em;
		line-height: 0.78;
		color: var(--obsidian-2);
		-webkit-text-stroke: 1px var(--obsidian-line);
		padding-top: 40px;
		user-select: none;
	}
	.giant span {
		display: inline-block;
	}
	.bar {
		display: flex;
		justify-content: space-between;
		gap: 16px;
		padding-block: 22px 28px;
		border-top: 1px solid var(--obsidian-line);
		font-size: var(--fs-xs);
		color: #6c7b75;
	}
	.build {
		font-family: var(--font-mono);
		letter-spacing: 0.06em;
		user-select: none;
	}
	.legal-links { display: flex; flex-wrap: wrap; gap: 10px 20px; padding-bottom: 26px; font-size: 12px; }
	.legal-links a { color: #c9d5d4; }
	.legal-links a:hover { color: #fff; }
	@media (max-width: 1024px) {
		.cols {
			grid-template-columns: repeat(2, 1fr);
		}
		.brandcol {
			grid-column: 1 / -1;
		}
		.notes {
			grid-template-columns: 1fr;
		}
	}
	@media (max-width: 560px) {
		.bar {
			flex-direction: column;
		}
	}
</style>
