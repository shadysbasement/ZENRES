<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api, DEMO, APP_URL, isProtonDomain, passwordScore, loginWithPasskey } from '$lib/auth';
	import Segmented from '$design/components/Segmented.svelte';

	interface Props {
		mode: 'login' | 'signup';
	}
	let { mode }: Props = $props();

	const roles = [
		{ value: 'msme', label: 'Manufacturer' },
		{ value: 'eu_importer', label: 'EU Importer' },
		{ value: 'transporter', label: 'Transporter' },
		{ value: 'buyer', label: 'Buyer' }
	] as const;
	type Role = (typeof roles)[number]['value'];
	const roleParam = page.url.searchParams.get('role');
	let role = $state<Role>(roleParam === 'transporter' ? 'transporter' : roleParam === 'eu_importer' ? 'eu_importer' : 'msme');

	const legalPublished = page.data.legalPublished ?? false;
	const termsVersion = page.data.legalTermsVersion ?? '';
	const privacyVersion = page.data.legalPrivacyVersion ?? '';
	let method = $state<'email' | 'proton'>(page.url.searchParams.get('method') === 'proton' ? 'proton' : 'email');
	$effect(() => { if (mode === 'signup' && legalPublished) method = 'email'; });
	let fullName = $state('');
	let email = $state('');
	let password = $state('');
	let termsAccepted = $state(false);
	let marketingOptIn = $state(false);
	let showPw = $state(false);
	let busy = $state(false);
	let error = $state('');
	let sent = $state<null | { email: string; encrypted: boolean }>(null);

	// Proton flow
	let protonEmail = $state(page.url.searchParams.get('email') ?? '');
	let wkd = $state<'idle' | 'checking' | 'found' | 'none'>('idle');
	let wkdFingerprint = $state('');
	let encryptEmail = $state(true);
	let customDomain = $state<'unknown' | 'proton' | 'not_proton'>('unknown');
	let insist = $state(false);

	const strength = $derived(passwordScore(password));
	const next = page.url.searchParams.get('next');
	const plan = page.url.searchParams.get('plan');

	const errors: Record<string, string> = {
		invalid_credentials: 'That email and password combination was not recognised.',
		email_taken: 'An account already exists for this email. Sign in instead.',
		weak_password: 'Use at least 12 characters with a mix of words, numbers or symbols.',
		email_unverified: 'Please verify your email first — we have re-sent the link.',
		rate_limited: 'Too many attempts. Please wait a few minutes and try again.',
		link_required: 'This email already has a password account. Confirm your password to link it.',
		legal_terms_required: 'Please accept the current terms and acknowledge the privacy notice.',
		legal_signup_requires_email: 'Create your account with email so your legal choices can be recorded.',
		not_proton: 'This domain does not appear to be hosted by Proton.',
		network: 'Network error. Check your connection and retry.'
	};

	async function submitEmail(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		if (mode === 'signup' && password.length < 12) {
			error = errors.weak_password;
			return;
		}
		if (mode === 'signup' && legalPublished && !termsAccepted) {
			error = errors.legal_terms_required;
			return;
		}
		busy = true;
		const r =
			mode === 'signup'
				? await api('register', { full_name: fullName, email, password, role, plan, terms_accepted: termsAccepted, terms_version: termsVersion, privacy_version: privacyVersion, marketing_opt_in: marketingOptIn })
				: await api<{ redirect?: string; onboarded?: boolean }>('login', { email, password });
		busy = false;
		if (r.ok) {
			if (mode === 'signup') sent = { email, encrypted: false };
			else {
				const d = r.data as { onboarded?: boolean };
				if (next?.startsWith('/')) goto(next);
				else if (d.onboarded) window.location.assign(`${APP_URL}/dashboard`);
				else goto('/onboarding');
			}
			return;
		}
		if (r.offline && DEMO && !(mode === 'signup' && legalPublished)) {
			// Local demo without the Go auth service running.
			if (mode === 'signup') goto(`/onboarding?role=${role}${plan ? `&plan=${plan}` : ''}&name=${encodeURIComponent(fullName)}`);
			else window.location.assign(`${APP_URL}/dashboard`);
			return;
		}
		error = errors[r.error] ?? 'Something went wrong. Please try again.';
	}

	function protonDomainOk(addr: string) {
		return /^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(addr);
	}

	let wkdTimer: ReturnType<typeof setTimeout>;
	$effect(() => {
		const addr = protonEmail.trim().toLowerCase();
		clearTimeout(wkdTimer);
		wkd = 'idle';
		customDomain = 'unknown';
		if (!protonDomainOk(addr)) return;
		wkdTimer = setTimeout(async () => {
			wkd = 'checking';
			const r = await api<{ proton: boolean; wkd_key: boolean; fingerprint?: string }>('proton/wkd', { email: addr });
			if (r.ok) {
				customDomain = r.data.proton ? 'proton' : 'not_proton';
				wkd = r.data.wkd_key ? 'found' : 'none';
				wkdFingerprint = r.data.fingerprint ?? '';
			} else {
				customDomain = isProtonDomain(addr) ? 'proton' : 'unknown';
				wkd = 'none';
			}
		}, 450);
	});

	async function submitProton(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		const addr = protonEmail.trim().toLowerCase();
		if (!protonDomainOk(addr)) {
			error = 'Enter your Proton email address.';
			return;
		}
		if (customDomain === 'not_proton' && !insist) {
			error = errors.not_proton;
			return;
		}
		busy = true;
		const r = await api('proton/magic-link', { email: addr, encrypt: wkd === 'found' && encryptEmail, role: mode === 'signup' ? role : undefined, override: insist });
		busy = false;
		if (r.ok || (r.offline && DEMO)) {
			sent = { email: addr, encrypted: wkd === 'found' && encryptEmail };
			return;
		}
		error = errors[r.error] ?? 'Could not send the sign-in link.';
	}

	async function passkey() {
		error = '';
		if (!('credentials' in navigator) || !window.PublicKeyCredential) {
			error = 'This browser does not support passkeys.';
			return;
		}
		busy = true;
		const r = await loginWithPasskey(protonEmail || email || undefined);
		busy = false;
		if (r.ok) window.location.assign(`${APP_URL}/dashboard`);
		else if (r.error !== 'cancelled') error = r.offline ? 'Passkey sign-in needs the auth service running.' : 'Passkey not recognised.';
	}
</script>

{#if sent}
	<div class="sent">
		<div class="env" aria-hidden="true">
			<svg viewBox="0 0 64 48"><rect x="1" y="1" width="62" height="46" rx="4" /><path d="M2 3l30 22L62 3" /></svg>
			{#if sent.encrypted}<span class="lock mono">PGP</span>{/if}
		</div>
		<h2>Check your inbox</h2>
		<p>
			We sent a single-use {method === 'proton' ? 'sign-in link (valid 15 minutes)' : 'verification link (valid 24 hours)'} to
			<strong>{sent.email}</strong>.
			{#if sent.encrypted}It is end-to-end encrypted to your Proton public key found via Web Key Directory.{/if}
		</p>
		{#if DEMO}
			<a class="btn btn-dark" href="/auth/verify-email?purpose={method === 'proton' ? 'proton_magic' : 'email_verify'}&demo=1&role={role}">Open demo link <span class="arrow">→</span></a>
		{/if}
		<button class="btn btn-ghost btn-sm" onclick={() => (sent = null)}>Use a different address</button>
	</div>
{:else}
	{#if page.url.searchParams.get('error') === 'legal_signup_requires_email'}<p class="err" role="alert">{errors.legal_signup_requires_email}</p>{/if}
	{#if mode === 'signup' && legalPublished}<p class="signup-note">For now, create a new account with work email. Other sign-in methods remain available for existing accounts.</p>{/if}
	{#if mode !== 'signup' || !legalPublished}
	<div class="providers">
		<a class="prov" href="/api/auth/oauth/google/start{next ? `?next=${encodeURIComponent(next)}` : ''}">
			<svg viewBox="0 0 24 24" aria-hidden="true"><path fill="currentColor" d="M12.48 10.92v3.28h7.84c-.24 1.84-.853 3.187-1.787 4.133-1.147 1.147-2.933 2.4-6.053 2.4-4.827 0-8.6-3.893-8.6-8.72s3.773-8.72 8.6-8.72c2.6 0 4.507 1.027 5.907 2.347l2.307-2.307C18.747 1.44 16.133 0 12.48 0 5.867 0 .307 5.387.307 12s5.56 12 12.173 12c3.573 0 6.267-1.173 8.373-3.36 2.16-2.16 2.84-5.213 2.84-7.667 0-.76-.053-1.467-.173-2.053H12.48z" /></svg>
			Continue with Google
		</a>
		<a class="prov" href="/api/auth/oauth/apple/start{next ? `?next=${encodeURIComponent(next)}` : ''}">
			<svg viewBox="0 0 24 24" aria-hidden="true"><path fill="currentColor" d="M12.152 6.896c-.948 0-2.415-1.078-3.96-1.04-2.04.027-3.91 1.183-4.961 3.014-2.117 3.675-.546 9.103 1.519 12.09 1.013 1.454 2.208 3.09 3.792 3.039 1.52-.065 2.09-.987 3.935-.987 1.831 0 2.35.987 3.96.948 1.637-.026 2.676-1.48 3.676-2.948 1.156-1.688 1.636-3.325 1.662-3.415-.039-.013-3.182-1.221-3.22-4.857-.026-3.04 2.48-4.494 2.597-4.559-1.429-2.09-3.623-2.324-4.39-2.376-2-.156-3.675 1.09-4.61 1.09zM15.53 3.83c.843-1.012 1.4-2.427 1.245-3.83-1.207.052-2.662.805-3.532 1.818-.78.896-1.454 2.338-1.273 3.714 1.338.104 2.715-.688 3.559-1.701" /></svg>
			Continue with Apple
		</a>
		<button class="prov proton" class:active={method === 'proton'} onclick={() => (method = method === 'proton' ? 'email' : 'proton')} aria-expanded={method === 'proton'}>
			<svg viewBox="0 0 24 24" aria-hidden="true"><path fill="#6D4AFF" d="M4 3h9.2a6.3 6.3 0 0 1 0 12.6H9.6V21H4V3Z" /><path fill="#fff" fill-opacity=".9" d="M9.6 8h3.4a1.7 1.7 0 0 1 0 3.4H9.6z" /></svg>
			Continue with Proton
		</button>
	</div>
	{/if}

	{#if mode === 'signup'}
		<div class="role">
			<span class="label">I am signing up as</span>
			<Segmented label="Role" bind:value={role} options={[...roles]} size="sm" />
		</div>
	{/if}

	{#if method === 'proton'}
		<form class="proton-panel" onsubmit={submitProton}>
			<div class="pp-head">
				<span class="eyebrow">Privacy-first Proton identity</span>
				<span class="pill pill-outline">Magic link · WKD · Passkey</span>
			</div>
			<label class="field">
				<span class="label">Proton email address</span>
				<input class="input" type="email" bind:value={protonEmail} placeholder="you@proton.me" autocomplete="email" required />
				<span class="hint">
					{#if wkd === 'checking'}Checking domain & Web Key Directory…
					{:else if customDomain === 'not_proton'}<span class="text-terracotta">MX records don't point to Proton.</span>
						<button type="button" class="linkbtn" onclick={() => (insist = !insist)}>{insist ? 'Undo override' : 'Continue anyway'}</button>
					{:else if customDomain === 'proton' || isProtonDomain(protonEmail)}<span class="text-emerald">Proton-hosted domain recognised.</span>
					{:else}Supports @proton.me, @protonmail.com, @pm.me, @protonmail.ch and custom Proton domains.{/if}
				</span>
			</label>
			<label class="toggle" class:disabled={wkd !== 'found'}>
				<input type="checkbox" bind:checked={encryptEmail} disabled={wkd !== 'found'} />
				<span class="sw" aria-hidden="true"></span>
				<span>
					<span class="label">PGP-encrypt login email</span>
					<span class="hint">
						{#if wkd === 'found'}Public key found via WKD · <span class="mono">{wkdFingerprint.slice(-16)}</span>{:else}Available when your domain publishes a WKD key.{/if}
					</span>
				</span>
			</label>
			<button class="btn btn-primary btn-lg full" disabled={busy}>{busy ? 'Sending…' : 'Email me a sign-in link'}</button>
			{#if mode === 'login'}
				<button type="button" class="btn btn-lg full" onclick={passkey} disabled={busy}>Sign in with passkey (Proton Pass)</button>
			{/if}
		</form>
	{:else}
		<div class="or"><span>or with email</span></div>
		<form class="email" onsubmit={submitEmail}>
			{#if mode === 'signup'}
				<label class="field">
					<span class="label">Full name</span>
					<input class="input" bind:value={fullName} autocomplete="name" required placeholder="Ananya Deshpande" />
				</label>
			{/if}
			<label class="field">
				<span class="label">Work email</span>
				<input class="input" type="email" bind:value={email} autocomplete="email" required placeholder="you@company.in" />
			</label>
			<label class="field">
				<span class="label-row"><span class="label">Password</span>{#if mode === 'login'}<a class="hint link-u" href="/auth/verify-email?purpose=password_reset">Forgot?</a>{/if}</span>
				<span class="pw">
					<input class="input" type={showPw ? 'text' : 'password'} bind:value={password} autocomplete={mode === 'signup' ? 'new-password' : 'current-password'} required minlength={mode === 'signup' ? 12 : 1} aria-invalid={mode === 'signup' && password.length > 0 && password.length < 12} />
					<button type="button" class="show mono" onclick={() => (showPw = !showPw)}>{showPw ? 'hide' : 'show'}</button>
				</span>
				{#if mode === 'signup'}
					<span class="meter" aria-hidden="true">
						{#each [1, 2, 3, 4] as n (n)}<span class:on={strength.score >= n} class="s{strength.score}"></span>{/each}
					</span>
					<span class="hint">Minimum 12 characters. {strength.label}</span>
				{/if}
			</label>
			{#if mode === 'signup' && legalPublished}
				<label class="legal-choice"><input type="checkbox" bind:checked={termsAccepted} required /><span>I agree to the <a href="/legal/terms" target="_blank" rel="noopener">Terms and Conditions</a> and acknowledge the <a href="/legal/privacy" target="_blank" rel="noopener">Privacy Notice</a>.</span></label>
				<label class="legal-choice"><input type="checkbox" bind:checked={marketingOptIn} /><span>Email me optional ZENRES product updates. I can withdraw this choice later.</span></label>
			{/if}
			<button class="btn btn-primary btn-lg full" disabled={busy}>
				{busy ? 'Please wait…' : mode === 'signup' ? 'Create account' : 'Sign in'} <span class="arrow">→</span>
			</button>
		</form>
	{/if}

	{#if error}<p class="err" role="alert">{error}</p>{/if}
{/if}

<style>
	.legal-choice { display: flex; align-items: flex-start; gap: 10px; margin: 14px 0; font-size: 13px; line-height: 1.5; }
	.legal-choice input { margin-top: 3px; accent-color: #6ab8ff; flex: none; }
	.legal-choice a { text-decoration: underline; text-underline-offset: 2px; }
	.signup-note { font-size: 13px; color: var(--ink-muted); margin-bottom: 16px; }
	.providers {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.prov {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 10px;
		height: 46px;
		border: var(--hairline);
		background: var(--surface-card);
		border-radius: var(--radius-sm);
		font-size: var(--fs-sm);
		font-weight: 500;
		color: var(--ink-primary);
		transition: border-color var(--dur-2), background var(--dur-2), transform var(--dur-2) var(--ease-out);
	}
	.prov:hover {
		border-color: var(--border-strong);
		background: var(--surface-canvas);
	}
	.prov:active {
		transform: translateY(1px);
	}
	.prov svg {
		width: 17px;
		height: 17px;
	}
	.prov.proton.active {
		border-color: #6d4aff;
		box-shadow: 0 0 0 3px rgba(109, 74, 255, 0.1);
	}
	.role {
		margin-top: 22px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.or {
		display: flex;
		align-items: center;
		gap: 12px;
		margin: 22px 0 18px;
		font-family: var(--font-mono);
		font-size: 10.5px;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--ink-faint);
	}
	.or::before,
	.or::after {
		content: '';
		flex: 1;
		border-top: var(--hairline);
	}
	.email,
	.proton-panel {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.proton-panel {
		margin-top: 18px;
		padding: 18px;
		border: 1px solid rgba(109, 74, 255, 0.25);
		border-radius: var(--radius-md);
		background: linear-gradient(180deg, rgba(109, 74, 255, 0.04), transparent 50%);
		animation: open 600ms var(--ease-expo);
	}
	@keyframes open {
		from {
			opacity: 0;
			transform: translateY(-8px);
		}
	}
	.pp-head {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}
	.label-row {
		display: flex;
		justify-content: space-between;
	}
	.pw {
		position: relative;
		display: block;
	}
	.pw .input {
		padding-right: 56px;
	}
	.show {
		position: absolute;
		right: 10px;
		top: 50%;
		transform: translateY(-50%);
		font-size: 11px;
		color: var(--ink-muted);
	}
	.meter {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 4px;
	}
	.meter span {
		height: 3px;
		border-radius: 2px;
		background: var(--surface-sunken);
		transition: background 300ms;
	}
	.meter span.on.s1 {
		background: var(--status-crimson);
	}
	.meter span.on.s2 {
		background: var(--status-terracotta);
	}
	.meter span.on.s3,
	.meter span.on.s4 {
		background: var(--brand-emerald);
	}
	.full {
		width: 100%;
	}
	.toggle {
		display: grid;
		grid-template-columns: auto 1fr;
		gap: 12px;
		align-items: start;
		cursor: pointer;
	}
	.toggle.disabled {
		opacity: 0.55;
		cursor: default;
	}
	.toggle input {
		position: absolute;
		opacity: 0;
		pointer-events: none;
	}
	.toggle > span:last-child {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.sw {
		position: relative;
		width: 34px;
		height: 20px;
		border-radius: 10px;
		background: var(--surface-sunken);
		transition: background 300ms;
		margin-top: 1px;
	}
	.sw::after {
		content: '';
		position: absolute;
		top: 2px;
		left: 2px;
		width: 16px;
		height: 16px;
		border-radius: 50%;
		background: #fff;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
		transition: transform 400ms var(--ease-expo);
	}
	.toggle input:checked + .sw {
		background: var(--brand-emerald);
	}
	.toggle input:checked + .sw::after {
		transform: translateX(14px);
	}
	.toggle input:focus-visible + .sw {
		outline: 2px solid var(--brand-emerald);
		outline-offset: 2px;
	}
	.linkbtn {
		margin-left: 6px;
		text-decoration: underline;
		font-size: inherit;
		color: var(--ink-primary);
	}
	.err {
		margin-top: 14px;
		padding: 10px 12px;
		font-size: var(--fs-xs);
		color: var(--status-crimson);
		background: var(--status-crimson-tint);
		border-radius: var(--radius-sm);
	}
	.sent {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 16px;
		animation: open 700ms var(--ease-expo);
	}
	.sent h2 {
		font-size: 1.8rem;
	}
	.sent p {
		font-size: var(--fs-sm);
	}
	.env {
		position: relative;
		width: 64px;
	}
	.env svg {
		fill: var(--surface-card);
		stroke: var(--ink-primary);
		stroke-width: 1.2;
	}
	.lock {
		position: absolute;
		right: -10px;
		bottom: -6px;
		font-size: 9px;
		background: var(--brand-emerald);
		color: #fff;
		padding: 2px 5px;
		border-radius: 3px;
	}
</style>
