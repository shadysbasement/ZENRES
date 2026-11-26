<script lang="ts">
	import { LEGAL_LINKS } from '$lib/legal/policies';
	let { data } = $props();
	const version = $derived(data.slug === 'terms' ? data.config.termsVersion : data.slug === 'privacy' ? data.config.privacyVersion : 'draft-1');
</script>

<svelte:head>
	<title>{data.policy.title} — ZENRES</title>
	{#if !data.config.published}<meta name="robots" content="noindex, nofollow" />{/if}
</svelte:head>

<article class="legal container">
	<a class="back" href="/">← ZENRES home</a>
	<p class="eyebrow">Legal {data.config.published ? 'information' : 'review draft'}</p>
	<h1>{data.policy.title}</h1>
	<p class="intro">{data.policy.intro}</p>
	<div class="metadata"><span>Version {version}</span><span>Effective date: {data.config.effectiveDate}</span></div>
	{#if !data.config.published}<p class="draft" role="status">Draft for legal review. This page is not an effective policy.</p>{/if}
	<nav class="policy-nav" aria-label="Legal documents">{#each LEGAL_LINKS as link (link.slug)}<a href="/legal/{link.slug}" aria-current={data.slug === link.slug ? 'page' : undefined}>{link.label}</a>{/each}</nav>
	<div class="sections">
		{#each data.policy.sections as section (section.heading)}<section><h2>{section.heading}</h2><p>{section.text}</p></section>{/each}
		<section><h2>Contact and legal details</h2><dl><div><dt>Legal entity</dt><dd>{data.config.entity}</dd></div><div><dt>Registered address</dt><dd>{data.config.address}</dd></div><div><dt>Support</dt><dd>{data.config.support}</dd></div><div><dt>Privacy requests</dt><dd>{data.config.privacy}</dd></div><div><dt>Jurisdiction</dt><dd>{data.config.jurisdiction}</dd></div><div><dt>Retention</dt><dd>{data.config.retention}</dd></div></dl></section>
	</div>
</article>

<style>
	.legal { max-width: 1000px; padding-block: clamp(84px, 10vw, 150px) 60px; }
	.back { display: inline-block; margin-bottom: 42px; font-size: 13px; color: var(--ink-muted); }
	h1 { max-width: 16ch; font-size: clamp(2.8rem, 6vw, 5.2rem); line-height: 1; letter-spacing: -.055em; font-weight: 600; margin: 12px 0 20px; }
	.intro { max-width: 72ch; font-size: clamp(17px, 1.8vw, 21px); line-height: 1.55; color: var(--ink-muted); }
	.metadata { display: flex; gap: 16px; flex-wrap: wrap; color: var(--ink-muted); font-size: 12px; margin-top: 24px; }
	.draft { margin-top: 28px; border: 1px solid #c9b3ed; background: #f5efff; color: #594677; padding: 14px 18px; border-radius: 12px; }
	.policy-nav { display: flex; flex-wrap: wrap; gap: 8px; margin: 45px 0; }
	.policy-nav a { border: var(--hairline); background: #fff9; border-radius: 30px; padding: 9px 14px; font-size: 12px; transition: transform .25s, background .25s; }
	.policy-nav a:hover { transform: translateY(-2px); background: #fff; }
	.policy-nav a[aria-current='page'] { background: #dbeaff; border-color: #a9cff6; }
	.sections { display: grid; gap: 30px; max-width: 78ch; }
	section { padding-bottom: 26px; border-bottom: var(--hairline); }
	section h2 { font-size: clamp(19px, 2vw, 24px); margin-bottom: 10px; letter-spacing: -.025em; }
	section p { color: var(--ink-muted); line-height: 1.75; }
	dl { display: grid; gap: 10px; }
	dl div { display: grid; grid-template-columns: minmax(130px, 180px) 1fr; gap: 14px; font-size: 14px; }
	dt { color: var(--ink-muted); }
	@media (max-width: 600px) { .legal { padding-inline: 20px; } dl div { grid-template-columns: 1fr; gap: 2px; } }
</style>
