<script lang="ts">
	import { reveal } from '$design/motion';
	let { data } = $props();
	const authLabel = $derived(data.auth === 'operational' ? 'Responding' : data.auth === 'unavailable' ? 'Not responding' : 'Not configured');
</script>

<svelte:head>
	<title>Service status · ZENRES</title>
	<meta name="description" content="Current reachability check for the ZENRES public site and authentication service." />
</svelte:head>

<div class="container-wide status-page">
	<header use:reveal>
		<span class="eyebrow">Service status</span>
		<h1>Current <span>reachability</span></h1>
		<p>These checks show whether this website and the authentication API respond right now. They do not measure every login flow, data service or customer workflow.</p>
	</header>
	<div class="checks">
		<section class="card check" use:reveal={{ delay: 80 }}>
			<div><span class="eyebrow">Public website</span><h2>zenres.vercel.app</h2></div>
			<span class="state good"><span class="dot"></span>Responding</span>
		</section>
		<section class="card check" use:reveal={{ delay: 160 }}>
			<div><span class="eyebrow">Authentication API</span><h2>Account and PGP sign-in service</h2></div>
			<span class:good={data.auth === 'operational'} class:bad={data.auth === 'unavailable'} class="state"><span class="dot"></span>{authLabel}</span>
		</section>
	</div>
	<p class="checked">Checked by the website at {new Date(data.checkedAt).toLocaleString('en-IN', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'Asia/Kolkata' })} IST. Refresh for a new check.</p>
	<section class="notices"><h2>Maintenance & incident notices</h2>{#if !data.noticesAvailable}<p>Notices are temporarily unavailable.</p>{:else if data.announcements.length}{#each data.announcements as item (item.id)}<article class="card"><span class="eyebrow">{item.state} · {new Date(item.published_at).toLocaleString()}</span><h3>{item.title}</h3><p>{item.body}</p></article>{/each}{:else}<p>No published notices. This does not guarantee every service is operating.</p>{/if}</section>
	<div class="note card"><h2>What this does not cover</h2><p>The workspace still uses sample data. Core API, worker jobs, database backups and payment services are not represented by these checks. Notices are published manually; automated incident detection and alerting are not available.</p><a class="link-u" href="/features">See feature availability →</a></div>
</div>

<style>
	.status-page { padding-top: clamp(128px, 15vw, 190px); padding-bottom: 40px; }
	header { max-width: 800px; }
	h1 { margin-top: 20px; font-size: var(--fs-display); line-height: .98; letter-spacing: var(--tracking-display); color: var(--ink-primary); }
	h1 span { color: var(--brand-emerald); }
	header p { max-width: 64ch; margin-top: 24px; font-size: clamp(1rem, 2vw, 1.22rem); line-height: 1.55; }
	.checks { display: grid; grid-template-columns: 1fr 1fr; gap: var(--gap); margin-top: 60px; }
	.check { min-height: 175px; padding: 24px; display: flex; flex-direction: column; justify-content: space-between; gap: 24px; }
	.check h2 { margin-top: 10px; font-size: 1.25rem; color: var(--ink-primary); }
	.state { width: fit-content; display: inline-flex; align-items: center; gap: 8px; padding: 7px 11px; border-radius: 99px; background: var(--surface-subtle); color: var(--ink-secondary); font-size: var(--fs-sm); }
	.dot { width: 8px; height: 8px; border-radius: 50%; background: currentColor; }
	.good { color: #197451; background: #e5f5ec; }
	.bad { color: #a6404f; background: #fbe9eb; }
	.checked { margin-top: 14px; color: var(--ink-muted); font-size: var(--fs-xs); }
	.notices { margin-top: 55px; }.notices h2 { color: var(--ink-primary); font-size: 1.6rem; margin-bottom: 16px; }.notices article { padding: 22px; margin-bottom: 10px; }.notices h3 { color: var(--ink-primary); font-size: 1.2rem; margin: 8px 0; }.notices p { line-height: 1.5; font-size: var(--fs-sm); }
	.note { margin-top: 54px; padding: 26px; max-width: 800px; }
	.note h2 { color: var(--ink-primary); font-size: 1.25rem; }
	.note p { max-width: 65ch; margin: 10px 0 20px; line-height: 1.6; font-size: var(--fs-sm); }
	@media (max-width: 700px) { .checks { grid-template-columns: 1fr; } }
</style>
