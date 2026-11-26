<script lang="ts">
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Icon from '$lib/components/ui/Icon.svelte';
	import { ws } from '$lib/stores/workspace.svelte';
	import { ROLE_LABEL } from '$lib/demo/seed';
	import { timeAgo } from '$lib/format';

	let email = $state('');
	let role = $state('production');
	const scope: Record<string, string> = {
		owner: 'Everything, billing & team',
		finance: 'Finance, Tally, billing',
		production: 'Plants, zones, batches, inventory',
		compliance: 'Carbon, CBAM, audit, exports',
		logistics_partner: 'Assigned shipments only',
		buyer: 'Approved catalogues & RFQs only',
		verifier: 'Audit locker — read-only'
	};
	const audit = [
		['Meera Joshi', 'exported Registry XML 2026-Q3', 380],
		['Prakash Iyer', 'imported Tally Day Book (42 vouchers)', 64],
		['R. Kulkarni (CA)', 'commented on Q3 calculation v4', 1440],
		['Ananya Deshpande', 'rotated API token zvt_…3f9a', 4320]
	] as const;
	let tokens = $state([{ name: 'Tally bridge (Chakan)', prefix: 'zvt_…3f9a', scope: 'tally:write', last: '14 min ago' }]);
	function invite() {
		ws.team.push({ name: email.split('@')[0], email, role, last: new Date().toISOString() });
		ws.toast(`Invitation sent to ${email} as ${ROLE_LABEL[role]}`);
		email = '';
	}
</script>

<svelte:head><title>Team & Access — ZENRES</title></svelte:head>

<PageHeader eyebrow="Settings" title="Your team &" serif="workspace access" lede="Invite your team, manage roles, and decide who can see and work with each part of your business." />

<div class="g2">
	<section class="panel">
		<div class="panel-head"><h2>Members</h2><span class="mono small muted">{ws.team.length} people</span></div>
		<div class="table-wrap">
			<table class="table">
				<thead><tr><th>Member</th><th>Role</th><th>Scope</th><th>Last active</th></tr></thead>
				<tbody>
					{#each ws.team as m (m.email)}
						<tr><td><span class="strong">{m.name}</span><div class="small muted">{m.email}</div></td><td><span class="pill" class:pill-emerald={m.role === 'owner'} class:pill-cobalt={['logistics_partner', 'buyer', 'verifier'].includes(m.role)}>{ROLE_LABEL[m.role]}</span></td><td class="small">{scope[m.role]}</td><td class="small muted">{timeAgo(m.last)}</td></tr>
					{/each}
				</tbody>
			</table>
		</div>
		<div class="inv">
			<input class="input" bind:value={email} placeholder="name@company.com" />
			<select class="select" bind:value={role}>{#each Object.entries(ROLE_LABEL) as [k, v] (k)}<option value={k}>{v}</option>{/each}</select>
			<button class="btn btn-sm btn-dark" disabled={!/@/.test(email)} onclick={invite}>Invite</button>
		</div>
	</section>
	<div class="side">
		<section class="panel">
			<div class="panel-head"><h2>Activity audit log</h2></div>
			{#each audit as [who, what, mins] (what)}<div class="log"><Icon name="shield" size={13} /><span><strong>{who}</strong> {what}</span><span class="mono small muted">{timeAgo(new Date(Date.now() - mins * 60000).toISOString())}</span></div>{/each}
		</section>
		<section class="panel">
			<div class="panel-head"><h2>API & connector tokens</h2><button class="btn btn-sm" onclick={() => (tokens = [...tokens, { name: 'New token', prefix: `zvt_…${Math.random().toString(16).slice(2, 6)}`, scope: 'read', last: 'never' }])}>New token</button></div>
			{#each tokens as t (t.prefix)}<div class="log"><Icon name="key" size={13} /><span><strong>{t.name}</strong> <span class="mono small">{t.prefix} · {t.scope}</span></span><span class="small muted">{t.last}</span></div>{/each}
			<p class="fine">Tokens are shown once, stored as SHA-256 hashes and can be revoked at any time.</p>
		</section>
	</div>
</div>

<style>
	.g2 {
		display: grid;
		grid-template-columns: 1.6fr 1fr;
		gap: 16px;
		align-items: start;
	}
	.side {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.strong {
		color: var(--ink-primary);
		font-weight: 500;
	}
	.small {
		font-size: 11.5px;
	}
	.inv {
		display: grid;
		grid-template-columns: 1fr 220px auto;
		gap: 8px;
		margin-top: 14px;
	}
	.log {
		display: grid;
		grid-template-columns: 16px 1fr auto;
		gap: 10px;
		align-items: center;
		padding: 9px 0;
		border-bottom: var(--hairline);
		font-size: 12.5px;
	}
	@media (max-width: 1100px) {
		.g2,
		.inv {
			grid-template-columns: 1fr;
		}
	}
</style>
