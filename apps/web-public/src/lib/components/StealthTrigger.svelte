<script lang="ts">
	import { onMount, tick } from 'svelte';

	// Generic diagnostics prompt. Contains no route information: the server answers with a
	// redirect target only when the passphrase hash matches its configured value.
	let open = $state(false);
	let value = $state('');
	let busy = $state(false);
	let input: HTMLInputElement | undefined = $state();
	let timer: ReturnType<typeof setTimeout>;

	async function show() {
		open = true;
		value = '';
		await tick();
		input?.focus();
		clearTimeout(timer);
		timer = setTimeout(() => (open = false), 20000);
	}

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		if (!value || busy) return;
		busy = true;
		try {
			const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(value));
			const k = [...new Uint8Array(digest)].map((b) => b.toString(16).padStart(2, '0')).join('');
			const res = await fetch('/api/telemetry', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ k })
			});
			if (res.status === 200) {
				const { r } = await res.json();
				if (typeof r === 'string' && r.startsWith('/')) window.location.assign(r);
			}
		} catch {
			/* silent */
		} finally {
			busy = false;
			value = '';
			open = false;
		}
	}

	onMount(() => {
		const onKey = (e: KeyboardEvent) => {
			if (e.ctrlKey && e.shiftKey && e.altKey && (e.code === 'KeyZ' || e.key.toLowerCase() === 'z')) {
				e.preventDefault();
				show();
			} else if (e.key === 'Escape') open = false;
		};
		const onFooter = () => show();
		window.addEventListener('keydown', onKey);
		window.addEventListener('zv:diag', onFooter);
		return () => {
			window.removeEventListener('keydown', onKey);
			window.removeEventListener('zv:diag', onFooter);
			clearTimeout(timer);
		};
	});
</script>

{#if open}
	<form class="diag" onsubmit={submit} autocomplete="off">
		<span class="mono">diag&gt;</span>
		<input bind:this={input} bind:value type="password" aria-label="Diagnostics key" spellcheck="false" autocomplete="off" />
	</form>
{/if}

<style>
	.diag {
		position: fixed;
		right: 16px;
		bottom: 16px;
		z-index: 9000;
		display: flex;
		align-items: center;
		gap: 8px;
		height: 34px;
		padding: 0 10px;
		background: var(--vault-bg);
		border: 1px solid var(--vault-border);
		border-radius: 3px;
		color: var(--vault-phosphor);
		font-family: var(--font-mono);
		font-size: 12px;
	}
	input {
		width: 160px;
		background: transparent;
		border: 0;
		outline: none;
		color: var(--vault-phosphor);
		font-family: var(--font-mono);
		font-size: 12px;
		caret-color: var(--vault-phosphor);
	}
</style>
