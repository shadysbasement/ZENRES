<script lang="ts">
	import { onMount } from 'svelte';
	import { reveal } from '$design/motion';
	import { articles, findArticles, type SupportLanguage } from '$lib/support/articles';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	type Ticket = { id: string; number: number; kind: string; category: string; priority: string; status: string; subject: string; plan: string; assigned_to?: string; callback_at?: string; created_at: string; updated_at: string };
	type Message = { id: string; author: string; body: string; created_at: string };
	let view = $state<'assist' | 'tickets' | 'help' | 'create' | 'status'>('assist');
	let language = $state<SupportLanguage>('en');
	let search = $state('');
	let question = $state('');
	let answer = $state<{ answer: string; source: string | null; mode: string; escalate: boolean } | null>(null);
	let aiConsent = $state(false);
	let thinking = $state(false);
	let signedIn = $state(false);
	let checkingAccount = $state(true);
	let tickets = $state<Ticket[]>([]);
	let active = $state<Ticket | null>(null);
	let messages = $state<Message[]>([]);
	let rating = $state<number | null>(null);
	let draft = $state({ kind: 'question', category: 'platform', subject: '', message: '', callbackAt: '' });
	let reply = $state('');
	let feedbackComment = $state('');
	let busy = $state(false);
	let notice = $state('');
	let selectedArticle = $state<string | null>(null);
	const results = $derived(findArticles(search));
	const cards = [
		{ id: 'assist', no: '01', icon: '✧', title: 'ZENRES Assist', note: 'Reviewed answers, in English or Hindi', tone: 'blue' },
		{ id: 'chat', no: '02', icon: '◌', title: 'Human support', note: 'A conversation with our team', tone: 'lilac' },
		{ id: 'tickets', no: '03', icon: '▤', title: 'Support tickets', note: 'Track requests and replies', tone: 'mint' },
		{ id: 'help', no: '04', icon: '⌕', title: 'Help centre', note: 'Guides and common questions', tone: 'peach' },
		{ id: 'callback', no: '05', icon: '↗', title: 'Request a callback', note: 'Suggest a time for our team', tone: 'yellow' },
		{ id: 'status', no: '06', icon: '◉', title: 'System status', note: 'Reachability and notices', tone: 'slate' }
	] as const;
	function csrf() { return document.cookie.match(/(?:^|;\s*)zv_csrf=([^;]+)/)?.[1] ?? ''; }
	async function api(path: string, init: RequestInit = {}) {
		const headers = new Headers(init.headers);
		if (init.body) { headers.set('content-type', 'application/json'); headers.set('x-csrf-token', decodeURIComponent(csrf())); }
		return fetch(`/api/support/${path}`, { ...init, headers, credentials: 'same-origin' });
	}
	async function loadTickets() {
		const response = await api('tickets');
		if (response.ok) tickets = await response.json();
	}
	async function loadTicket(id: string) {
		const response = await api(`tickets/${id}`);
		if (!response.ok) return;
		const result = await response.json();
		active = result.ticket; messages = result.messages; rating = result.rating;
	}
	onMount(() => {
		let live = true;
		fetch('/api/auth/me', { credentials: 'same-origin' }).then(async (res) => {
			if (!live) return;
			signedIn = res.ok;
			checkingAccount = false;
			if (res.ok) await loadTickets();
		}).catch(() => { checkingAccount = false; });
		const timer = setInterval(() => { if (live && active) void loadTicket(active.id); }, 15000);
		return () => { live = false; clearInterval(timer); };
	});
	function openCard(id: string) {
		notice = '';
		if (id === 'chat' || id === 'callback') { draft.kind = id === 'chat' ? 'live_chat' : 'callback'; view = 'create'; }
		else view = id as typeof view;
	}
	async function ask() {
		if (question.trim().length < 3) return;
		thinking = true; answer = null;
		try {
			const response = await fetch('/api/support/ask', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ question, language, use_ai: aiConsent && data.aiEnabled }) });
			answer = response.ok ? await response.json() : { answer: 'The guide is unavailable. Please try the help centre or create a ticket.', source: null, mode: 'guided', escalate: true };
		} catch { answer = { answer: 'The guide is unavailable. Please try the help centre or create a ticket.', source: null, mode: 'guided', escalate: true }; }
		thinking = false;
	}
	async function createTicket() {
		if (busy || !signedIn) return;
		busy = true; notice = '';
		try {
			const body = { kind: draft.kind, category: draft.category, subject: draft.subject, message: draft.message, callback_at: draft.kind === 'callback' && draft.callbackAt ? new Date(draft.callbackAt).toISOString() : '', callback_timezone: Intl.DateTimeFormat().resolvedOptions().timeZone };
			const response = await api('tickets', { method: 'POST', body: JSON.stringify(body) });
			if (!response.ok) { const result = await response.json(); throw Error(result.error === 'invalid_callback_time' ? 'Choose a time at least two hours from now, within the next 60 days.' : 'Could not create the ticket. Check the fields and try again.'); }
			const created = await response.json();
			await loadTickets(); await loadTicket(created.id);
			draft = { kind: 'question', category: 'platform', subject: '', message: '', callbackAt: '' };
			view = 'tickets'; notice = `Ticket ZEN-${created.number} created.`;
		} catch (error) { notice = error instanceof Error ? error.message : 'Could not create the ticket.'; }
		busy = false;
	}
	async function sendReply() {
		if (!active || !reply.trim() || busy) return;
		busy = true; notice = '';
		const response = await api(`tickets/${active.id}/messages`, { method: 'POST', body: JSON.stringify({ body: reply }) }).catch(() => null);
		if (response?.ok) { reply = ''; await loadTicket(active.id); await loadTickets(); }
		else notice = 'Your reply was not sent. Please try again.';
		busy = false;
	}
	async function saveRating(value: number) {
		if (!active || busy) return;
		busy = true;
		const response = await api(`tickets/${active.id}/feedback`, { method: 'POST', body: JSON.stringify({ rating: value, comment: feedbackComment }) }).catch(() => null);
		if (response?.ok) { rating = value; notice = 'Thanks for your feedback.'; }
		else notice = 'Feedback could not be saved.';
		busy = false;
	}
</script>

<svelte:head>
	<title>Support centre · ZENRES</title>
	<meta name="description" content="ZENRES support centre: guides, assistance, tickets, callbacks and service status." />
</svelte:head>

<div class="support container-wide">
	<section class="hero" use:reveal>
		<div class="hero-copy"><span class="eyebrow">ZENRES / SUPPORT CENTRE</span><h1>How can we<br /><span>help you?</span></h1><p>Clear answers for the work that matters. Search a guide, start a conversation, or keep track of a request.</p></div>
		<div class="hero-art" aria-hidden="true"><span class="orb orb-a"></span><span class="orb orb-b"></span><span class="hero-symbol">✳</span><span class="hero-caption">GOOD QUESTIONS<br />DESERVE CLEAR ANSWERS.</span></div>
	</section>
	<div class="search-shell" use:reveal={{ delay: 80 }}><span class="search-icon">⌕</span><input aria-label="Search support articles" type="search" placeholder="Search for answers..." bind:value={search} oninput={() => { view = 'help'; selectedArticle = null; }} /><button onclick={() => (view = 'help')}>Search <span>↗</span></button></div>
	<div class="section-heading"><span class="eyebrow">CHOOSE A PATH</span><span class="micro">01—06 / SUPPORT TOOLS</span></div>
	<div class="cards" use:reveal={{ delay: 120 }}>
		{#each cards as card (card.id)}<button class="support-card {card.tone}" class:active={view === card.id || (view === 'create' && (card.id === 'chat' && draft.kind === 'live_chat' || card.id === 'callback' && draft.kind === 'callback'))} onclick={() => openCard(card.id)} aria-pressed={view === card.id}>
			<span class="card-top"><span class="card-no">{card.no}</span><span class="card-icon">{card.icon}</span></span><span class="card-bottom"><strong>{card.title}</strong><small>{card.note}</small></span><span class="card-arrow">↗</span>
		</button>{/each}
	</div>
	<section class="workbench" id="support-workbench" use:reveal={{ delay: 170 }}>
		<div class="workbench-head"><div><span class="eyebrow">YOUR SUPPORT SPACE</span><h2>{view === 'assist' ? 'Ask ZENRES Assist' : view === 'help' ? 'Explore the help centre' : view === 'tickets' ? 'Your support tickets' : view === 'create' ? (draft.kind === 'callback' ? 'Request a callback' : draft.kind === 'live_chat' ? 'Start a conversation' : draft.kind === 'bug' ? 'Report an issue' : 'Create a ticket') : 'Status & notices'}</h2></div><span class="workbench-mark">ZENRES / HELP</span></div>
		{#if view === 'assist'}
			<div class="assist-grid"><div><p class="intro">Answers use our reviewed help content. For account issues or questions requiring professional verification, a support ticket is the next step.</p><div class="language"><span>Language</span><button class:on={language === 'en'} onclick={() => (language = 'en')}>English</button><button class:on={language === 'hi'} onclick={() => (language = 'hi')}>हिन्दी</button></div><form onsubmit={(e) => { e.preventDefault(); void ask(); }}><label for="assist-question">Your question</label><textarea id="assist-question" bind:value={question} maxlength="500" rows="4" placeholder="How do I connect Tally or understand my CBAM estimate?"></textarea>{#if data.aiEnabled}<label class="checkbox"><input type="checkbox" bind:checked={aiConsent} /> Use AI to draft an answer from these guides. Do not include personal or financial data.</label>{/if}<button class="action" type="submit" disabled={thinking || question.trim().length < 3}>{thinking ? 'Finding an answer…' : 'Find an answer'} <span>↗</span></button></form></div><div class="assist-result"><span class="eyebrow">REVIEWED GUIDANCE</span>{#if answer}<p class="answer">{answer.answer}</p><span class="source">{answer.mode === 'ai' ? 'AI draft, check the source' : 'From the help centre'}{answer.source ? ` · ${articles.find((a) => a.id === answer?.source)?.title[language]}` : ''}</span>{#if answer.escalate}<p class="caution">This may need a qualified professional or a human support reply.</p>{/if}<button class="text-action" onclick={() => { draft.kind = 'question'; view = 'create'; }}>Create a support ticket ↗</button>{:else}<div class="empty-orb" aria-hidden="true">✳</div><p>Ask about Tally, carbon reporting, traceability, freight or billing. The guide will show where its answer came from.</p>{/if}</div></div>
		{:else if view === 'help'}
			<p class="intro">{results.length} {results.length === 1 ? 'guide' : 'guides'}{search ? ` for “${search}”` : ' to get you started'}.</p><div class="article-grid">{#each results as article (article.id)}<article class="article"><span class="eyebrow">{article.category}</span><button class="article-title" aria-expanded={selectedArticle === article.id} onclick={() => (selectedArticle = selectedArticle === article.id ? null : article.id)}>{article.title[language]} <span>{selectedArticle === article.id ? '−' : '+'}</span></button>{#if selectedArticle === article.id}<p>{article.answer[language]}</p>{#if article.escalate}<small>Professional verification may be required.</small>{/if}{/if}</article>{/each}</div>{#if results.length === 0}<div class="empty-state">No reviewed guide matches that search. <button class="text-action" onclick={() => { draft.kind = 'question'; view = 'create'; }}>Ask the team ↗</button></div>{/if}<div class="language"><span>Read in</span><button class:on={language === 'en'} onclick={() => (language = 'en')}>English</button><button class:on={language === 'hi'} onclick={() => (language = 'hi')}>हिन्दी</button></div>
		{:else if view === 'create'}
			{#if checkingAccount}<p class="intro">Checking your account…</p>{:else if !signedIn}<div class="empty-state"><h3>Sign in to contact support</h3><p>Your ticket and conversation history are private to your account.</p><a class="action" href="/login?next=%2Fsupport">Sign in <span>↗</span></a></div>{:else}<div class="form-grid"><form onsubmit={(e) => { e.preventDefault(); void createTicket(); }}><label for="ticket-kind">Request type</label><select id="ticket-kind" bind:value={draft.kind}><option value="question">Question or complaint</option><option value="live_chat">Human conversation</option><option value="callback">Callback request</option><option value="bug">Report an issue</option><option value="feedback">Product suggestion</option></select><label for="ticket-category">Topic</label><select id="ticket-category" bind:value={draft.category}><option value="platform">Platform</option><option value="tally">Tally integration</option><option value="cbam">CBAM and carbon</option><option value="trace">QR and traceability</option><option value="freight">Freight and shipments</option><option value="marketplace">Marketplace</option><option value="billing">Billing</option><option value="security">Security and privacy</option><option value="other">Other</option></select><label for="ticket-subject">Subject</label><input id="ticket-subject" bind:value={draft.subject} minlength="5" maxlength="160" required placeholder="A short description" /><label for="ticket-message">Details</label><textarea id="ticket-message" bind:value={draft.message} minlength="10" maxlength="4000" rows="6" required placeholder="Tell us what happened and what you expected. Do not include passwords or financial records."></textarea>{#if draft.kind === 'callback'}<label for="callback-time">Preferred time</label><input id="callback-time" type="datetime-local" bind:value={draft.callbackAt} required /><p class="field-note">A request, not a confirmed appointment. We will reply in the ticket. No phone number has been published yet.</p>{/if}<button class="action" type="submit" disabled={busy}>{busy ? 'Sending…' : 'Create ticket'} <span>↗</span></button>{#if notice}<p class="notice" role="status">{notice}</p>{/if}</form><aside class="form-aside"><span class="eyebrow">WHAT HAPPENS NEXT</span><h3>Keep the full story in one place.</h3><p>We create a private ticket with a reference number. You can return to read replies and add context. Human response times and staffed chat hours have not been published.</p><p>Support agents cannot access your financial records through this workflow. Such access would need separate explicit authorisation.</p></aside></div>{/if}
		{:else if view === 'tickets'}
			{#if checkingAccount}<p class="intro">Checking your account…</p>{:else if !signedIn}<div class="empty-state"><h3>Sign in to see your tickets</h3><a class="action" href="/login?next=%2Fsupport">Sign in <span>↗</span></a></div>{:else}<div class="ticket-tools"><p class="intro">Your conversations and requests, in one place.</p><button class="action" onclick={() => { active = null; draft.kind = 'question'; view = 'create'; }}>New ticket <span>↗</span></button></div>{#if notice}<p class="notice" role="status">{notice}</p>{/if}<div class="ticket-layout"><div class="ticket-list">{#each tickets as ticket (ticket.id)}<button class:chosen={active?.id === ticket.id} onclick={() => void loadTicket(ticket.id)}><span class="ticket-meta">ZEN-{ticket.number} · {ticket.category} <em>{ticket.status}</em></span><strong>{ticket.subject}</strong><small>Updated {new Date(ticket.updated_at).toLocaleDateString('en-IN')}</small></button>{/each}{#if tickets.length === 0}<p class="intro">No tickets yet. Start with a question, an issue report or a callback request.</p>{/if}</div><div class="conversation">{#if active}<div class="conversation-head"><span class="eyebrow">ZEN-{active.number} · {active.priority} priority</span><h3>{active.subject}</h3><p>{active.status} · {active.plan} plan{active.assigned_to ? ` · ${active.assigned_to}` : ''}</p>{#if active.callback_at}<p>Preferred callback: {new Date(active.callback_at).toLocaleString()}</p>{/if}</div><div class="messages" aria-live="polite">{#each messages as message (message.id)}<div class="message" class:team={message.author === 'support'}><span>{message.author === 'support' ? 'ZENRES support' : 'You'} · {new Date(message.created_at).toLocaleString()}</span><p>{message.body}</p></div>{/each}</div>{#if active.status !== 'resolved'}<form class="reply-form" onsubmit={(e) => { e.preventDefault(); void sendReply(); }}><label for="reply">Add a reply</label><textarea id="reply" bind:value={reply} maxlength="4000" rows="3" required></textarea><button class="action" type="submit" disabled={busy || !reply.trim()}>Send reply <span>↗</span></button></form>{:else}<div class="rating"><strong>How was this support experience?</strong><div>{#each [1, 2, 3, 4, 5] as score}<button class:chosen={rating === score} onclick={() => void saveRating(score)} aria-label="Rate {score} out of 5">{score}</button>{/each}</div><input aria-label="Optional feedback comment" placeholder="Optional comment" bind:value={feedbackComment} maxlength="1000" /></div>{/if}{:else}<div class="conversation-empty"><span>◌</span><p>Select a ticket to see the conversation.</p></div>{/if}</div></div>{/if}
		{:else}<p class="intro">This view combines the current service reachability check with notices published by the support team. It does not certify every workflow.</p><a class="action" href="/status">See service reachability <span>↗</span></a><div class="announcements"><h3>Maintenance & incident notices</h3>{#if !data.announcementsAvailable}<p>Notices are temporarily unavailable. Check the service status page or try again later.</p>{:else if data.announcements.length}{#each data.announcements as item (item.id)}<article><span class="eyebrow">{item.state} · {new Date(item.published_at).toLocaleString()}</span><h4>{item.title}</h4><p>{item.body}</p></article>{/each}{:else}<p>No published maintenance or incident notices. This does not guarantee all services are operating.</p>{/if}</div>{/if}
	</section>
	<section class="bottom-grid"><div><span class="eyebrow">POPULAR TOPICS</span><h2>Start with the <span>common questions.</span></h2><div class="topics">{#each articles.slice(0, 7) as article (article.id)}<button onclick={() => { search = article.keywords[0]; selectedArticle = article.id; view = 'help'; document.getElementById('support-workbench')?.scrollIntoView({ behavior: 'smooth' }); }}>{article.title.en}<span>↗</span></button>{/each}</div></div><div class="channels"><span class="eyebrow">OTHER WAYS TO REACH US</span><h3>Choose what is available.</h3><div class="channel"><span>WhatsApp</span>{#if data.whatsapp}<a href="https://wa.me/{data.whatsapp}" target="_blank" rel="noopener noreferrer">Open WhatsApp ↗</a>{:else}<span>Not configured yet</span>{/if}</div><div class="channel"><span>Email</span>{#if data.email}<a href="mailto:{data.email}">{data.email} ↗</a>{:else}<span>Address not published yet</span>{/if}</div><p>Tickets are the available account-based way to keep a request and its replies together. Human coverage and response targets will be published after staffing is confirmed.</p><button class="text-action" onclick={() => { draft.kind = 'bug'; view = 'create'; document.getElementById('support-workbench')?.scrollIntoView({ behavior: 'smooth' }); }}>Report an issue ↗</button><button class="text-action" onclick={() => { draft.kind = 'feedback'; view = 'create'; document.getElementById('support-workbench')?.scrollIntoView({ behavior: 'smooth' }); }}>Share product feedback ↗</button></div></section>
</div>

<style>
	.support { padding-top: 112px; padding-bottom: 40px; }
	.hero { display: grid; grid-template-columns: 1.25fr .75fr; min-height: 430px; border: 1px solid #ffffffd8; border-radius: 30px; overflow: hidden; background: linear-gradient(115deg,#f3f8fc,#f2f1fa 55%,#eaf7f2); box-shadow: 0 18px 65px #697c9c12; }
	.hero-copy { padding: clamp(38px,6vw,84px); display: flex; flex-direction: column; justify-content: center; }
	.hero h1 { margin: 24px 0 20px; color: var(--ink-primary); font-size: clamp(3.8rem,8vw,8.2rem); line-height: .9; letter-spacing: -.07em; }
	.hero h1 span,.bottom-grid h2 span { color: #7d97b3; }
	.hero p { max-width: 49ch; font-size: clamp(1rem,1.5vw,1.2rem); line-height: 1.55; color: var(--ink-secondary); }
	.hero-art { position: relative; overflow: hidden; display: grid; place-items: center; background: radial-gradient(circle at 65% 38%,#e4d8f5,#d7e7f2 54%,#d4efe8); }
	.orb { position: absolute; border-radius: 50%; filter: blur(1px); }
	.orb-a { width: 370px; height: 370px; border: 1px solid #ffffff90; box-shadow: 0 0 0 45px #ffffff25,0 0 0 100px #ffffff22; animation: orbit 18s ease-in-out infinite alternate; }
	.orb-b { width: 220px; height: 220px; background: linear-gradient(140deg,#ffffffa3,#cfe4ef86); box-shadow: inset -24px -28px 45px #b7cee366, 22px 35px 70px #829cbc42; animation: float 8s ease-in-out infinite; }
	.hero-symbol { position: relative; font-size: 7rem; color: #7392b0; text-shadow: 0 12px 18px #718ba656; animation: spin 40s linear infinite; }
	.hero-caption { position: absolute; right: 26px; bottom: 25px; font-size: 10px; letter-spacing: .16em; line-height: 1.55; color: #5f748b; }
	.search-shell { display: flex; align-items: center; gap: 16px; margin: -31px auto 70px; position: relative; z-index: 2; width: min(820px,calc(100% - 36px)); padding: 7px 7px 7px 22px; background: #fff; border: 1px solid #e5ebf1; border-radius: 100px; box-shadow: 0 18px 50px #6175951b; }
	.search-icon { font-size: 28px; color: #7993a9; }
	.search-shell input { flex: 1; min-width: 0; border: none; outline: none; padding: 12px 4px; background: transparent; color: var(--ink-primary); }
	.search-shell button,.action { display: inline-flex; align-items: center; justify-content: space-between; gap: 30px; border-radius: 100px; padding: 13px 20px; background: linear-gradient(110deg,#a9d8ef,#c8b8ea); color: #33465c; font-size: 13px; font-weight: 650; transition: transform .3s var(--ease-out),box-shadow .3s var(--ease-out); }
	.search-shell button:hover,.action:hover { transform: translateY(-2px); box-shadow: 0 8px 25px #8c9bba50; }
	.action:disabled { opacity: .55; cursor: wait; }
	.section-heading { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
	.micro,.workbench-mark { font-size: 10px; letter-spacing: .16em; color: #8391a3; }
	.cards { display: grid; grid-template-columns: repeat(3,1fr); gap: 14px; }
	.support-card { min-height: 182px; text-align: left; padding: 21px; border: 1px solid #ffffffdf; border-radius: 22px; display: flex; flex-direction: column; justify-content: space-between; position: relative; overflow: hidden; transition: transform .45s var(--ease-out),box-shadow .45s var(--ease-out),border-color .3s; }
	.support-card:hover,.support-card.active { transform: translateY(-5px); box-shadow: 0 18px 42px #667b9b1c; border-color: #fff; }
	.support-card.blue { background: #e7f1fa; }.support-card.lilac { background: #eee8f7; }.support-card.mint { background: #e7f3ed; }.support-card.peach { background: #f8eee9; }.support-card.yellow { background: #f6f1df; }.support-card.slate { background: #e9edf3; }
	.card-top,.card-bottom { display: flex; justify-content: space-between; align-items: flex-start; }
	.card-no { font-size: 11px; letter-spacing: .12em; color: #8492a1; }.card-icon { font-size: 39px; line-height: .8; color: #718ba9; transition: transform .5s var(--ease-out); }.support-card:hover .card-icon { transform: rotate(18deg) scale(1.1); }
	.card-bottom { flex-direction: column; gap: 4px; }.card-bottom strong { font-size: 22px; color: #394b61; letter-spacing: -.04em; }.card-bottom small { color: #6b7c89; font-size: 12px; }.card-arrow { position: absolute; right: 22px; bottom: 22px; color: #6b8295; transition: transform .3s; }.support-card:hover .card-arrow { transform: translate(3px,-3px); }
	.workbench { margin-top: 60px; padding: clamp(26px,4.5vw,58px); border: 1px solid #e5eaf0; background: #ffffffd6; border-radius: 30px; box-shadow: 0 10px 50px #6a7f9910; scroll-margin-top: 90px; min-height: 350px; }
	.workbench-head { display: flex; justify-content: space-between; gap: 20px; align-items: start; padding-bottom: 30px; margin-bottom: 30px; border-bottom: 1px solid #e6ebf0; }.workbench h2,.bottom-grid h2 { font-size: clamp(2rem,3.8vw,3.7rem); letter-spacing: -.06em; line-height: 1; color: var(--ink-primary); margin-top: 12px; }
	.intro { line-height: 1.6; color: var(--ink-secondary); max-width: 67ch; margin-bottom: 20px; }
	.assist-grid,.form-grid { display: grid; grid-template-columns: 1fr .9fr; gap: clamp(25px,6vw,90px); }
	.language { display: flex; gap: 8px; align-items: center; margin: 22px 0; }.language span { font-size: 12px; color: var(--ink-muted); margin-right: 8px; }.language button { padding: 7px 13px; border-radius: 100px; border: 1px solid #dde5ed; font-size: 12px; }.language button.on { background: #deecf8; border-color: #b9d1e6; color: #3d5b7b; }
	form { display: flex; flex-direction: column; align-items: start; gap: 10px; }form label { font-size: 12px; color: #536477; font-weight: 650; margin-top: 8px; }form input:not([type=checkbox]),form select,form textarea,.rating input { width: 100%; border: 1px solid #dae3ec; border-radius: 13px; padding: 13px 15px; background: #f9fbfd; color: #344a5e; outline: none; }form textarea { resize: vertical; line-height: 1.5; }form input:focus,form textarea:focus,form select:focus { border-color: #8eb6d7; box-shadow: 0 0 0 3px #c9e6f659; }.checkbox { display: flex; gap: 10px; align-items: start; font-weight: 400; line-height: 1.4; }.checkbox input { margin-top: 3px; }.field-note { font-size: 12px; color: var(--ink-muted); line-height: 1.5; }.form-grid form .action,.assist-grid form .action { margin-top: 10px; }
	.assist-result,.form-aside { border-radius: 24px; background: linear-gradient(160deg,#eff5fa,#f5f0fa); padding: clamp(25px,3.5vw,42px); min-height: 330px; }.assist-result { display: flex; flex-direction: column; align-items: start; }.assist-result p { line-height: 1.6; color: #52677a; }.assist-result .answer { font-size: 17px; color: #344c63; margin-top: 24px; }.source,.caution { font-size: 12px; color: #75869a; margin: 15px 0; }.empty-orb { font-size: 80px; color: #9ab5cf; margin: auto; animation: spin 40s linear infinite; }.text-action { display: block; color: #48749d; font-size: 13px; font-weight: 650; margin-top: 12px; text-align: left; }.text-action:hover { text-decoration: underline; }
	.article-grid { display: grid; grid-template-columns: repeat(2,1fr); gap: 12px; }.article { border: 1px solid #e3eaf0; border-radius: 17px; padding: 21px; background: #fcfdff; }.article-title { display: flex; align-items: center; justify-content: space-between; width: 100%; gap: 10px; font-size: 17px; font-weight: 650; color: #344b60; text-align: left; margin-top: 12px; }.article-title span { color: #8197ad; font-size: 22px; }.article p { line-height: 1.6; color: #5c6e80; margin-top: 13px; }.article small { color: #ac6a4c; display: block; margin-top: 10px; }
	.empty-state { padding: 45px; background: #f0f5f9; border-radius: 20px; }.empty-state h3,.form-aside h3 { font-size: 24px; color: #3d526b; letter-spacing: -.04em; }.empty-state p,.form-aside p { line-height: 1.6; max-width: 55ch; margin: 12px 0 22px; color: #65788b; }.form-aside h3 { margin-top: 24px; }.notice { padding: 11px 15px; background: #ecf5fa; border-radius: 11px; color: #4c6880; font-size: 13px; }
	.ticket-tools { display: flex; justify-content: space-between; align-items: center; gap: 20px; }.ticket-layout { display: grid; grid-template-columns: 320px 1fr; gap: 14px; min-height: 400px; }.ticket-list { display: flex; flex-direction: column; gap: 7px; }.ticket-list button { text-align: left; padding: 15px; border: 1px solid #e3eaf0; border-radius: 14px; background: #f9fbfd; display: flex; flex-direction: column; gap: 6px; }.ticket-list button.chosen { border-color: #9bbbd5; background: #edf5fa; }.ticket-list strong { color: #43586c; }.ticket-list small,.ticket-meta { color: #7b8da0; font-size: 11px; }.ticket-meta { display: flex; justify-content: space-between; gap: 10px; width: 100%; }.ticket-meta em { font-style: normal; text-transform: uppercase; font-size: 10px; }.conversation { border: 1px solid #e1e8f0; border-radius: 18px; padding: 23px; }.conversation-head h3 { color: #3f536b; font-size: 23px; margin: 8px 0; }.conversation-head p { color: #718496; font-size: 12px; }.messages { display: flex; flex-direction: column; gap: 10px; margin: 22px 0; max-height: 400px; overflow: auto; }.message { max-width: 85%; padding: 12px 15px; border-radius: 14px; background: #eaf2f8; align-self: end; }.message.team { align-self: start; background: #f0ebf7; }.message span { color: #7b8da0; font-size: 10px; }.message p { color: #3e5369; line-height: 1.55; white-space: pre-wrap; }.reply-form { border-top: 1px solid #e3eaf0; padding-top: 15px; }.rating { border-top: 1px solid #e3eaf0; padding-top: 17px; display: flex; flex-direction: column; gap: 12px; color: #4c6276; }.rating div { display: flex; gap: 7px; }.rating button { width: 36px; height: 36px; border: 1px solid #cddde8; border-radius: 50%; }.rating button.chosen { background: #b9dbee; }.conversation-empty { min-height: 300px; display: grid; place-content: center; text-align: center; color: #8597a9; }.conversation-empty span { font-size: 65px; }
	.announcements { margin-top: 40px; padding-top: 25px; border-top: 1px solid #e5ebf0; }.announcements h3 { font-size: 23px; color: #435970; margin-bottom: 20px; }.announcements article { padding: 17px 0; border-top: 1px solid #e6ebf0; }.announcements h4 { font-size: 18px; color: #3d5267; margin: 8px 0; }.announcements p { line-height: 1.55; color: #6f8091; }
	.bottom-grid { display: grid; grid-template-columns: 1.2fr .8fr; gap: 12%; margin: 110px 0 40px; }.bottom-grid h2 { margin: 17px 0 28px; max-width: 14ch; }.topics { border-top: 1px solid #dbe4ec; }.topics button { display: flex; justify-content: space-between; gap: 20px; width: 100%; padding: 16px 0; border-bottom: 1px solid #dbe4ec; text-align: left; color: #4c6074; transition: padding .3s,color .3s; }.topics button:hover { padding-left: 8px; color: #4379a2; }.channels { padding: 36px; border-radius: 24px; background: #eef3f8; align-self: start; }.channels h3 { font-size: 28px; color: #41566c; letter-spacing: -.04em; margin: 17px 0 30px; }.channel { display: flex; justify-content: space-between; gap: 15px; padding: 16px 0; border-top: 1px solid #dbe5ee; font-size: 13px; }.channel span:last-child { color: #8695a5; }.channel a { color: #447a9f; }.channels p { font-size: 13px; line-height: 1.6; color: #718296; margin: 24px 0; }
	@keyframes float { 50% { transform: translateY(-18px) rotate(6deg); } }@keyframes orbit { 50% { transform: scale(1.12) translate(10px,-10px); } }@keyframes spin { to { transform: rotate(360deg); } }
	@media(max-width:900px) { .cards { grid-template-columns: repeat(2,1fr); }.hero { grid-template-columns: 1fr .45fr; }.bottom-grid { gap: 5%; }.ticket-layout { grid-template-columns: 1fr; }.ticket-list { max-height: 320px; overflow: auto; } }
	@media(max-width:650px) { .support { padding-top: 85px; }.hero { min-height: 460px; grid-template-columns: 1fr; }.hero-copy { padding: 40px 28px; }.hero-art { min-height: 120px; }.hero-symbol { font-size: 4rem; }.orb-a { width: 150px;height:150px; }.orb-b { width: 100px;height:100px; }.hero-caption { display: none; }.cards,.assist-grid,.form-grid,.article-grid,.bottom-grid { grid-template-columns: 1fr; }.search-shell { margin-bottom: 45px; }.search-shell button { padding: 12px 15px; }.search-shell button span { display: none; }.support-card { min-height: 155px; }.workbench { margin-top: 40px; }.workbench-mark { display: none; }.ticket-tools { align-items: start; flex-direction: column; }.bottom-grid { margin-top: 75px; }.channels { padding: 26px; } }
	@media(prefers-reduced-motion:reduce) { .orb,.hero-symbol,.empty-orb { animation: none; }.support-card,.search-shell button,.action { transition: none; } }
</style>
