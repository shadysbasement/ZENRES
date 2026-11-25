<script lang="ts">
	import { page } from '$app/state';
	import { api, DEMO, APP_URL } from '$lib/auth';
	import { TIERS } from '$lib/content';

	const roles = [
		{ value: 'msme', label: 'Indian MSME Manufacturer', body: 'Factories exporting CBAM goods to the EU.' },
		{ value: 'eu_importer', label: 'EU Importer / Broker', body: 'Importers, brokers and trading houses managing suppliers.' },
		{ value: 'transporter', label: 'Logistics Transporter', body: 'Carriers listing return capacity on FreightLink.' },
		{ value: 'buyer', label: 'International Buyer', body: 'Sourcing verified, carbon-transparent suppliers.' }
	] as const;
	const sectors = ['Steel', 'Aluminium', 'Cement', 'Fertilisers', 'Hydrogen', 'General Manufacturing'];
	const countries = [
		['IN', 'India'],
		['DE', 'Germany'],
		['NL', 'Netherlands'],
		['FR', 'France'],
		['IT', 'Italy'],
		['ES', 'Spain'],
		['BE', 'Belgium'],
		['PL', 'Poland'],
		['GB', 'United Kingdom'],
		['US', 'United States']
	];

	const qp = page.url.searchParams;
	let step = $state(0);
	let dir = $state(1);
	let role = $state<string>(qp.get('role') ?? 'msme');
	let orgName = $state('');
	let country = $state(qp.get('role') === 'eu_importer' ? 'DE' : 'IN');
	let taxId = $state('');
	let sector = $state('Aluminium');
	let plan = $state<string>(qp.get('plan') ?? (qp.get('role') === 'eu_importer' ? 'importer_hub' : 'growth'));
	let sandbox = $state(true);
	let busy = $state(false);
	let error = $state('');

	$effect(() => {
		if (role === 'transporter') plan = 'transporter_free';
		else if (plan === 'transporter_free') plan = 'growth';
	});

	const taxLabel = $derived(country === 'IN' ? 'GSTIN' : 'EORI number');
	const taxValid = $derived(
		!taxId || (country === 'IN' ? /^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z][1-9A-Z]Z[0-9A-Z]$/.test(taxId.toUpperCase()) : /^[A-Z]{2}[0-9A-Z]{1,15}$/.test(taxId.toUpperCase()))
	);
	const canNext = $derived(step === 0 ? !!role : step === 1 ? orgName.trim().length > 1 && taxValid : !!plan);

	function go(n: number) {
		dir = n > step ? 1 : -1;
		step = n;
	}

	async function finish() {
		busy = true;
		error = '';
		const payload = { role, org_name: orgName.trim(), country_code: country, gstin_or_eori: taxId.toUpperCase() || null, sector, plan, sandbox };
		const r = await api<{ redirect: string }>('onboarding/complete', payload);
		busy = false;
		if (r.ok) {
			window.location.assign(r.data.redirect);
			return;
		}
		if (r.offline || (DEMO && r.status === 401)) {
			const tier = plan === 'transporter_free' ? 'starter' : plan;
			window.location.assign(`${APP_URL}/dashboard?demo_tier=${tier}&org=${encodeURIComponent(orgName.trim())}`);
			return;
		}
		error = 'Could not complete onboarding. Please try again.';
	}

	const steps = ['Role', 'Organisation', 'Plan'];
</script>

<svelte:head><title>Set up your organisation — ZENRES</title></svelte:head>

<div class="ob">
	<div class="progress" aria-label="Onboarding progress">
		{#each steps as s, i (s)}
			<button class="ps" class:active={i === step} class:done={i < step} onclick={() => i < step && go(i)} disabled={i > step}>
				<span class="mono">0{i + 1}</span>{s}
			</button>
		{/each}
		<span class="track"><span style="transform:scaleX({(step + 1) / steps.length})"></span></span>
	</div>

	{#key step}
		<div class="stepbody" style="--dir:{dir}">
			{#if step === 0}
				<h1>What best describes <span class="serif">your organisation?</span></h1>
				<div class="roles">
					{#each roles as r (r.value)}
						<label class="opt" class:sel={role === r.value}>
							<input type="radio" name="role" value={r.value} bind:group={role} />
							<span class="radio" aria-hidden="true"></span>
							<span><span class="opt-t">{r.label}</span><span class="opt-b">{r.body}</span></span>
						</label>
					{/each}
				</div>
			{:else if step === 1}
				<h1>Tell us about <span class="serif">your organisation.</span></h1>
				<div class="form">
					<label class="field">
						<span class="label">Organisation name</span>
						<input class="input" bind:value={orgName} placeholder={role === 'eu_importer' ? 'Rhein Metall Import GmbH' : 'Deccan Alloy Works Pvt. Ltd.'} required />
					</label>
					<div class="two">
						<label class="field">
							<span class="label">Country</span>
							<select class="select" bind:value={country}>
								{#each countries as [code, name] (code)}<option value={code}>{name}</option>{/each}
							</select>
						</label>
						<label class="field">
							<span class="label">{taxLabel} <span class="muted">(optional for sandbox)</span></span>
							<input class="input mono" bind:value={taxId} aria-invalid={!taxValid} placeholder={country === 'IN' ? '27AAACD1234F1Z5' : 'DE123456789012'} />
							{#if !taxValid}<span class="hint text-crimson">That {taxLabel} format doesn’t look right.</span>{/if}
						</label>
					</div>
					<div class="field">
						<span class="label">Primary sector</span>
						<div class="chips">
							{#each sectors as s (s)}
								<button type="button" class="chip" class:sel={sector === s} onclick={() => (sector = s)}>{s}</button>
							{/each}
						</div>
					</div>
				</div>
			{:else}
				<h1>Choose how you <span class="serif">want to start.</span></h1>
				<div class="plans">
					{#if role === 'transporter'}
						<label class="opt sel">
							<input type="radio" checked />
							<span class="radio" aria-hidden="true"></span>
							<span><span class="opt-t">Free Transporter Registration</span><span class="opt-b">List return capacity, get verified and receive matched loads. Organic ranking never depends on payment.</span></span>
							<span class="price mono">$0</span>
						</label>
					{:else}
						{#each TIERS as t (t.id)}
							<label class="opt" class:sel={plan === t.id}>
								<input type="radio" name="plan" value={t.id} bind:group={plan} />
								<span class="radio" aria-hidden="true"></span>
								<span><span class="opt-t">{t.name} {#if t.flagship}<span class="pill pill-emerald">Flagship</span>{/if}</span><span class="opt-b">{t.coverage}</span></span>
								<span class="price mono">${t.monthly}<span class="muted">/mo</span></span>
							</label>
						{/each}
						<label class="sandbox">
							<input type="checkbox" bind:checked={sandbox} />
							<span>Start with a 14-day sandbox preview seeded with pilot data — no card required.</span>
						</label>
					{/if}
				</div>
			{/if}
		</div>
	{/key}

	{#if error}<p class="err">{error}</p>{/if}

	<div class="nav">
		{#if step > 0}<button class="btn btn-ghost" onclick={() => go(step - 1)}>← Back</button>{:else}<span></span>{/if}
		{#if step < 2}
			<button class="btn btn-dark" disabled={!canNext} onclick={() => go(step + 1)}>Continue <span class="arrow">→</span></button>
		{:else}
			<button class="btn btn-primary" disabled={busy || !canNext} onclick={finish}>{busy ? 'Creating workspace…' : 'Enter workspace'} <span class="arrow">→</span></button>
		{/if}
	</div>
</div>

<style>
	.ob {
		display: flex;
		flex-direction: column;
		gap: 28px;
	}
	.progress {
		position: relative;
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		padding-bottom: 12px;
	}
	.ps {
		display: flex;
		gap: 8px;
		align-items: baseline;
		font-size: var(--fs-xs);
		font-weight: 500;
		color: var(--ink-faint);
		text-align: left;
	}
	.ps .mono {
		font-size: 10px;
	}
	.ps.active,
	.ps.done {
		color: var(--ink-primary);
	}
	.ps.active .mono,
	.ps.done .mono {
		color: var(--brand-emerald);
	}
	.track {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		height: 1px;
		background: var(--border-hairline);
	}
	.track span {
		display: block;
		height: 100%;
		background: var(--brand-emerald);
		transform-origin: left;
		transition: transform 800ms var(--ease-expo);
	}
	.stepbody {
		display: flex;
		flex-direction: column;
		gap: 22px;
		animation: slide 700ms var(--ease-expo);
	}
	@keyframes slide {
		from {
			opacity: 0;
			transform: translateX(calc(var(--dir) * 24px));
		}
	}
	h1 {
		font-size: clamp(1.8rem, 3vw, 2.4rem);
		letter-spacing: -0.03em;
	}
	h1 .serif {
		color: var(--brand-emerald);
	}
	.roles,
	.plans {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.opt {
		position: relative;
		display: grid;
		grid-template-columns: auto 1fr auto;
		gap: 14px;
		align-items: center;
		padding: 16px;
		border: var(--hairline);
		border-radius: var(--radius-md);
		background: var(--surface-card);
		cursor: pointer;
		transition: border-color 250ms, box-shadow 250ms;
	}
	.opt:hover {
		border-color: var(--border-strong);
	}
	.opt.sel {
		border-color: var(--brand-emerald);
		box-shadow: 0 0 0 3px var(--brand-emerald-tint);
	}
	.opt input {
		position: absolute;
		opacity: 0;
		pointer-events: none;
	}
	.radio {
		width: 16px;
		height: 16px;
		border-radius: 50%;
		border: 1.5px solid var(--border-strong);
		transition: border 250ms;
	}
	.opt.sel .radio {
		border: 5px solid var(--brand-emerald);
	}
	.opt-t {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: var(--fs-sm);
		font-weight: 500;
		color: var(--ink-primary);
	}
	.opt-b {
		display: block;
		margin-top: 2px;
		font-size: var(--fs-xs);
		color: var(--ink-muted);
	}
	.price {
		font-size: var(--fs-sm);
		color: var(--ink-primary);
	}
	.form {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.two {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 12px;
	}
	.chips {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.chip {
		height: 32px;
		padding: 0 12px;
		border: var(--hairline);
		border-radius: var(--radius-sm);
		background: var(--surface-card);
		font-size: var(--fs-xs);
		color: var(--ink-secondary);
		transition: all 200ms;
	}
	.chip.sel {
		background: var(--obsidian);
		border-color: var(--obsidian);
		color: var(--ink-inverse);
	}
	.sandbox {
		display: flex;
		gap: 10px;
		align-items: flex-start;
		font-size: var(--fs-xs);
		margin-top: 6px;
		accent-color: var(--brand-emerald);
	}
	.nav {
		display: flex;
		justify-content: space-between;
	}
	.err {
		font-size: var(--fs-xs);
		color: var(--status-crimson);
	}
	@media (max-width: 520px) {
		.two {
			grid-template-columns: 1fr;
		}
	}
</style>
