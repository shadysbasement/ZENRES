<script lang="ts">
	import { onMount } from 'svelte';
	import { enhance } from '$app/forms';

	type FormState = {
		stage?: number;
		error?: string;
		challenge_id?: string;
		armored?: string;
		expires_in?: number;
		identifier?: string;
	} | null;
	let { form }: { form: FormState } = $props();

	let stage = $state(1);
	let challengeId = $state('');
	let armored = $state('');
	let identifier = $state('');
	let fingerprintInput = $state('');
	let signed = $state('');
	let error = $state('');
	let busy = $state(false);
	let remaining = $state(120);
	let total = $state(120);
	let copied = $state<'' | 'msg' | 'cmd'>('');
	let boot = $state<string[]>([]);
	let expiresAt = 0;

	$effect(() => {
		if (!form) return;
		error = form.error ?? '';
		if (form.stage === 2 && form.armored) {
			stage = 2;
			challengeId = form.challenge_id ?? '';
			armored = form.armored;
			identifier = form.identifier ?? identifier;
			total = form.expires_in ?? 120;
			expiresAt = Date.now() + total * 1000;
			remaining = total;
			signed = '';
		} else if (form.stage === 1) {
			stage = 1;
		}
	});

	onMount(() => {
		const lines = ['zv-vault 1.0 — rfc9580 challenge-response', 'transport ........ tls 1.3', 'auth mode ........ pgp proof-of-possession only', 'private key ...... never leaves your machine', 'ready.'];
		lines.forEach((l, i) => setTimeout(() => (boot = [...boot, l]), 140 * i + 120));
		const id = setInterval(() => {
			if (stage !== 2) return;
			remaining = Math.max(0, Math.ceil((expiresAt - Date.now()) / 1000));
			if (remaining === 0) {
				stage = 1;
				error = 'CHALLENGE EXPIRED — REQUEST A NEW ONE';
			}
		}, 250);
		return () => clearInterval(id);
	});

	const mmss = $derived(`${String(Math.floor(remaining / 60)).padStart(2, '0')}:${String(remaining % 60).padStart(2, '0')}`);
	const fpr = $derived(/^[0-9A-F]{40,64}$/.test(identifier) ? identifier : '<FINGERPRINT>');
	const helper = $derived(`gpg --decrypt challenge.asc | gpg --clearsign --local-user ${fpr} > response.asc`);

	async function copy(text: string, which: 'msg' | 'cmd') {
		await navigator.clipboard.writeText(text);
		copied = which;
		setTimeout(() => (copied = ''), 1600);
	}

	function downloadChallenge() {
		const a = document.createElement('a');
		a.href = URL.createObjectURL(new Blob([armored], { type: 'text/plain' }));
		a.download = 'challenge.asc';
		a.click();
		setTimeout(() => URL.revokeObjectURL(a.href), 1000);
	}

	const submitting = () => {
		busy = true;
		error = '';
		return async ({ update }: { update: (o?: { reset?: boolean }) => Promise<void> }) => {
			await update({ reset: false });
			busy = false;
		};
	};

	const R = 54;
	const C = 2 * Math.PI * R;
</script>

<div class="vault">
	<div class="frame">
		<header>
			<span class="dot" class:on={stage === 2}></span>
			<span>ZV // VAULT</span>
			<span class="sp"></span>
			<span class="dim">{stage === 1 ? 'PHASE 1 · IDENTITY CLAIM' : 'PHASE 2 · SIGNED RESPONSE'}</span>
		</header>

		<pre class="boot" aria-live="polite">{#each boot as l (l)}<span>&gt; {l}</span>{'\n'}{/each}</pre>

		{#if stage === 1}
			<form method="POST" action="?/challenge" use:enhance={submitting} class="phase" autocomplete="off">
				<label for="identifier">PGP FINGERPRINT / KEY HANDLE</label>
				<input
					id="identifier"
					name="identifier"
					bind:value={fingerprintInput}
					oninput={() => (identifier = fingerprintInput.replace(/\s+/g, '').toUpperCase())}
					spellcheck="false"
					placeholder="4A9F 0C21 …  (40 hex)"
					maxlength="80"
					required
				/>
				<button disabled={busy || fingerprintInput.trim().length < 3}>{busy ? 'ENCRYPTING CHALLENGE…' : 'REQUEST ENCRYPTED CHALLENGE'}</button>
			</form>
		{:else}
			<div class="phase two">
				<div class="countdown">
					<svg viewBox="0 0 120 120" aria-hidden="true">
						<circle cx="60" cy="60" r={R} class="ring-bg" />
						<circle cx="60" cy="60" r={R} class="ring" class:low={remaining <= 20} style="stroke-dasharray:{C};stroke-dashoffset:{C * (1 - remaining / total)}" />
					</svg>
					<div class="cd-txt"><span class:low={remaining <= 20}>{mmss}</span><small>TTL</small></div>
					<div class="bar"><span style="transform:scaleX({remaining / total})" class:low={remaining <= 20}></span></div>
				</div>

				<div class="block">
					<div class="block-head">
						<span>ENCRYPTED CHALLENGE · {challengeId.slice(0, 13)}…</span>
						<button type="button" onclick={() => copy(armored, 'msg')}>{copied === 'msg' ? 'COPIED ✓' : 'COPY'}</button>
						<button type="button" onclick={downloadChallenge}>DOWNLOAD</button>
					</div>
					<pre class="armored" aria-label="Encrypted challenge">{armored}</pre>
				</div>

				<div class="block">
					<div class="block-head">
						<span>DECRYPT + SIGN LOCALLY</span>
						<button type="button" onclick={() => copy(helper, 'cmd')}>{copied === 'cmd' ? 'COPIED ✓' : 'COPY'}</button>
					</div>
					<code class="cmd">$ {helper}</code>
					<p class="dim">Run this command on the computer holding your private key, then open response.asc and paste its contents below.</p>
				</div>

				<form method="POST" action="?/verify" use:enhance={submitting} autocomplete="off">
					<input type="hidden" name="challenge_id" value={challengeId} />
					<label for="signed">PASTE CLEARSIGNED RESPONSE</label>
					<textarea id="signed" name="signed" bind:value={signed} spellcheck="false" rows="9" placeholder="-----BEGIN PGP SIGNED MESSAGE-----"></textarea>
					<button disabled={busy || !signed.includes('-----BEGIN PGP')}>{busy ? 'VERIFYING SIGNATURE…' : 'VERIFY SIGNATURE & ENTER VAULT'}</button>
				</form>
			</div>
		{/if}

		{#if error}<p class="error" role="alert">✗ {error}</p>{/if}

		<footer class="dim">max 5 challenges / 15 min · 3 verify attempts / challenge · every event is audit-logged</footer>
	</div>
</div>

<style>
	:global(html:has(.vault)),
	:global(body:has(.vault)) {
		background: var(--vault-bg);
	}
	.vault {
		min-height: 100dvh;
		background: var(--vault-bg);
		color: #c4d6ce;
		font-family: var(--font-mono);
		font-size: 13px;
		display: grid;
		place-items: center;
		padding: 24px;
		background-image: linear-gradient(var(--vault-panel) 1px, transparent 1px), linear-gradient(90deg, var(--vault-panel) 1px, transparent 1px);
		background-size: 32px 32px;
	}
	.frame {
		width: 100%;
		max-width: 760px;
		border: 1px solid var(--vault-border);
		background: rgba(10, 14, 13, 0.92);
		animation: in 600ms var(--ease-expo);
	}
	@keyframes in {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
	}
	header {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 16px;
		border-bottom: 1px solid var(--vault-border);
		font-size: 11px;
		letter-spacing: 0.12em;
		color: var(--vault-phosphor);
	}
	.dot {
		width: 7px;
		height: 7px;
		background: var(--vault-dim);
	}
	.dot.on {
		background: var(--vault-phosphor);
		box-shadow: 0 0 10px var(--vault-phosphor);
	}
	.sp {
		flex: 1;
	}
	.dim {
		color: var(--vault-dim);
	}
	.boot {
		margin: 0;
		padding: 14px 16px 0;
		font-size: 11.5px;
		color: var(--vault-dim);
		line-height: 1.7;
		min-height: 110px;
		white-space: pre-wrap;
	}
	.boot span:last-of-type {
		color: var(--vault-phosphor);
	}
	.phase {
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding: 18px 16px 8px;
	}
	label {
		font-size: 11px;
		letter-spacing: 0.12em;
		color: var(--vault-dim);
	}
	input,
	textarea {
		width: 100%;
		background: var(--vault-panel);
		border: 1px solid var(--vault-border);
		border-radius: 0;
		color: var(--vault-phosphor);
		font-family: var(--font-mono);
		font-size: 14px;
		padding: 12px;
		letter-spacing: 0.06em;
		caret-color: var(--vault-phosphor);
		outline: none;
		transition: border-color 200ms;
	}
	textarea {
		font-size: 11.5px;
		letter-spacing: 0;
		resize: vertical;
		line-height: 1.5;
	}
	input:focus,
	textarea:focus {
		border-color: var(--vault-phosphor);
	}
	input::placeholder,
	textarea::placeholder {
		color: #33463f;
	}
	button {
		height: 42px;
		border: 1px solid var(--vault-phosphor);
		background: transparent;
		color: var(--vault-phosphor);
		font-family: var(--font-mono);
		font-size: 12px;
		letter-spacing: 0.12em;
		transition: background 200ms, color 200ms;
	}
	button:hover:not(:disabled) {
		background: var(--vault-phosphor);
		color: var(--vault-bg);
	}
	button:disabled {
		opacity: 0.35;
	}
	.two form {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.countdown {
		display: grid;
		grid-template-columns: 64px 1fr;
		grid-template-rows: auto auto;
		align-items: center;
		gap: 4px 16px;
		position: relative;
	}
	.countdown svg {
		width: 64px;
		height: 64px;
		grid-row: span 2;
		transform: rotate(-90deg);
	}
	.ring-bg {
		fill: none;
		stroke: var(--vault-border);
		stroke-width: 4;
	}
	.ring {
		fill: none;
		stroke: var(--vault-phosphor);
		stroke-width: 4;
		transition: stroke-dashoffset 250ms linear, stroke 300ms;
	}
	.ring.low,
	.bar span.low {
		stroke: var(--status-crimson);
		background: var(--status-crimson);
	}
	.cd-txt {
		display: flex;
		align-items: baseline;
		gap: 8px;
		font-size: 26px;
		color: var(--vault-phosphor);
		font-variant-numeric: tabular-nums;
	}
	.cd-txt .low {
		color: var(--status-crimson);
	}
	.cd-txt small {
		font-size: 10px;
		color: var(--vault-dim);
		letter-spacing: 0.12em;
	}
	.bar {
		height: 2px;
		background: var(--vault-border);
	}
	.bar span {
		display: block;
		height: 100%;
		background: var(--vault-phosphor);
		transform-origin: left;
		transition: transform 250ms linear;
	}
	.block {
		border: 1px solid var(--vault-border);
	}
	.block-head {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0 0 0 12px;
		border-bottom: 1px solid var(--vault-border);
		font-size: 10.5px;
		letter-spacing: 0.1em;
		color: var(--vault-dim);
	}
	.block-head button {
		height: 30px;
		border: 0;
		border-left: 1px solid var(--vault-border);
		padding: 0 14px;
		font-size: 10.5px;
	}
	.armored {
		margin: 0;
		padding: 12px;
		max-height: 180px;
		overflow: auto;
		font-size: 11px;
		line-height: 1.45;
		color: #9fc9b5;
		white-space: pre-wrap;
		word-break: break-all;
	}
	.cmd {
		display: block;
		padding: 12px;
		font-size: 11.5px;
		color: #9fc9b5;
		word-break: break-all;
	}
	.error {
		margin: 8px 16px 0;
		padding: 10px 12px;
		border: 1px solid var(--status-crimson);
		color: #e48a8a;
		font-size: 12px;
		letter-spacing: 0.06em;
	}
	footer {
		padding: 14px 16px;
		margin-top: 12px;
		border-top: 1px solid var(--vault-border);
		font-size: 10.5px;
		letter-spacing: 0.06em;
	}
</style>
