<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api, DEMO, APP_URL, enrollPasskey } from '$lib/auth';

	const qp = page.url.searchParams;
	const purpose = qp.get('purpose') ?? 'email_verify';
	const token = qp.get('token') ?? '';
	const demo = qp.get('demo') === '1' && DEMO;

	let phase = $state<'working' | 'ok' | 'error' | 'passkey' | 'reset'>(purpose === 'password_reset' && !token ? 'reset' : 'working');
	let message = $state('');
	let passkeyState = $state<'idle' | 'busy' | 'done' | 'skipped' | 'error'>('idle');
	let resetEmail = $state('');
	let resetSent = $state(false);
	let onboarded = false;

	onMount(async () => {
		if (phase === 'reset') return;
		if (demo) {
			setTimeout(() => (phase = purpose === 'proton_magic' ? 'passkey' : 'ok'), 900);
			return;
		}
		if (!token) {
			phase = 'error';
			message = 'This link is missing its token.';
			return;
		}
		const r = await api<{ offer_passkey?: boolean; onboarded?: boolean }>(purpose === 'proton_magic' ? 'proton/verify' : 'verify-email', { token });
		if (r.ok) {
			onboarded = !!r.data.onboarded;
			phase = r.data.offer_passkey ? 'passkey' : 'ok';
		} else {
			phase = 'error';
			message = r.error === 'token_expired' ? 'This link has expired. Request a new one.' : r.error === 'token_consumed' ? 'This link was already used.' : 'This link is invalid.';
		}
	});

	async function addPasskey() {
		passkeyState = 'busy';
		if (demo) {
			setTimeout(() => (passkeyState = 'done'), 800);
			return;
		}
		const r = await enrollPasskey();
		passkeyState = r.ok ? 'done' : r.error === 'cancelled' ? 'idle' : 'error';
	}

	function proceed() {
		if (onboarded) window.location.assign(`${APP_URL}/dashboard`);
		else goto(`/onboarding${qp.get('role') ? `?role=${qp.get('role')}` : ''}`);
	}

	async function requestReset(e: SubmitEvent) {
		e.preventDefault();
		await api('password/forgot', { email: resetEmail });
		resetSent = true; // identical response whether or not the account exists
	}
</script>

<svelte:head><title>Verify — Zevra</title></svelte:head>

<div class="v">
	{#if phase === 'working'}
		<div class="spin" aria-hidden="true"></div>
		<h1>Verifying your link…</h1>
		<p class="mono muted">single-use token · constant-time check</p>
	{:else if phase === 'ok'}
		<span class="pill pill-emerald"><span class="dot"></span>Verified</span>
		<h1>Your email is <span class="serif">confirmed.</span></h1>
		<button class="btn btn-primary btn-lg" onclick={proceed}>Continue to onboarding <span class="arrow">→</span></button>
	{:else if phase === 'passkey'}
		<span class="pill pill-emerald"><span class="dot"></span>Signed in with Proton</span>
		<h1>Make the next sign-in <span class="serif">one tap.</span></h1>
		<p>Save a passkey to Proton Pass (or your device). Next time, sign in with a biometric or hardware key — no email round-trip.</p>
		<div class="pk card">
			<svg viewBox="0 0 48 48" aria-hidden="true"><circle cx="18" cy="18" r="9" /><path d="M24 24l14 14m-5-5 4-4m-8 0 4-4" /></svg>
			<div>
				<span class="label">Proton Pass / device passkey</span>
				<span class="hint">WebAuthn · phishing-resistant · bound to zevra.io</span>
			</div>
		</div>
		{#if passkeyState === 'done'}
			<p class="text-emerald">Passkey saved.</p>
			<button class="btn btn-primary btn-lg" onclick={proceed}>Continue <span class="arrow">→</span></button>
		{:else}
			<div class="row">
				<button class="btn btn-primary btn-lg" onclick={addPasskey} disabled={passkeyState === 'busy'}>{passkeyState === 'busy' ? 'Waiting for authenticator…' : 'Add passkey'}</button>
				<button class="btn btn-ghost btn-lg" onclick={proceed}>Skip for now</button>
			</div>
			{#if passkeyState === 'error'}<p class="text-crimson">Could not save the passkey. You can add one later in settings.</p>{/if}
		{/if}
	{:else if phase === 'reset'}
		<h1>Reset your <span class="serif">password.</span></h1>
		{#if resetSent}
			<p>If an account exists for <strong>{resetEmail}</strong>, a reset link is on its way. It expires in 1 hour.</p>
		{:else}
			<form class="reset" onsubmit={requestReset}>
				<label class="field"><span class="label">Email</span><input class="input" type="email" bind:value={resetEmail} required /></label>
				<button class="btn btn-dark btn-lg">Send reset link</button>
			</form>
		{/if}
	{:else}
		<span class="pill pill-crimson">Link problem</span>
		<h1>{message}</h1>
		<a class="btn btn-lg" href="/login">Back to sign in</a>
	{/if}
</div>

<style>
	.v {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 18px;
		animation: in 700ms var(--ease-expo);
	}
	@keyframes in {
		from {
			opacity: 0;
			transform: translateY(12px);
		}
	}
	h1 {
		font-size: clamp(1.9rem, 3.2vw, 2.6rem);
		letter-spacing: -0.03em;
	}
	h1 .serif {
		color: var(--brand-emerald);
	}
	.spin {
		width: 28px;
		height: 28px;
		border-radius: 50%;
		border: 2px solid var(--border-hairline);
		border-top-color: var(--brand-emerald);
		animation: spin 0.8s linear infinite;
	}
	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
	.pk {
		display: flex;
		gap: 14px;
		align-items: center;
		padding: 16px;
		width: 100%;
	}
	.pk svg {
		width: 40px;
		fill: none;
		stroke: #6d4aff;
		stroke-width: 2.5;
		stroke-linecap: round;
	}
	.pk > div {
		display: flex;
		flex-direction: column;
	}
	.row {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.reset {
		display: flex;
		flex-direction: column;
		gap: 14px;
		width: 100%;
	}
</style>
