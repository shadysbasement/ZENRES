import type { RequestHandler } from './$types';
import { env } from '$env/dynamic/private';
import { findArticles, type SupportLanguage } from '$lib/support/articles';

/** Public, stateless guide over published help text. Optional model calls need explicit user consent. */
export const POST: RequestHandler = async ({ request, fetch }) => {
	let payload: { question?: string; language?: SupportLanguage; use_ai?: boolean };
	try { payload = await request.json(); } catch { return Response.json({ error: 'invalid_input' }, { status: 400 }); }
	const question = payload.question?.trim() ?? '';
	if (question.length < 3 || question.length > 500) return Response.json({ error: 'invalid_input' }, { status: 400 });
	const language: SupportLanguage = payload.language === 'hi' ? 'hi' : 'en';
	const matches = findArticles(question).slice(0, 3);
	const leading = matches[0];
	const escalate = !leading || leading.escalate || /\b(legal|tax|regulation|verify|certif|audit|declare|dispute)\b/i.test(question);
	const fallback = leading?.answer[language] ?? (language === 'hi' ? 'इस प्रश्न का सत्यापित उत्तर अभी उपलब्ध नहीं है। कृपया सहायता टिकट बनाएँ।' : 'I do not have a reviewed answer for that yet. Please create a support ticket.');
	const base = { answer: fallback, source: leading?.id ?? null, mode: 'guided', escalate };
	if (!payload.use_ai || !env.SUPPORT_AI_API_KEY || !env.SUPPORT_AI_MODEL || !leading) return Response.json(base);
	// Model use is account-only. The provider receives only public guide text, never the user's question.
	const me = await fetch('/api/auth/me', { headers: { cookie: request.headers.get('cookie') ?? '' } }).catch(() => null);
	if (!me?.ok) return Response.json(base);
	const guide = matches.map((a) => `${a.title[language]}: ${a.answer[language]}`).join('\n\n');
	try {
		const response = await fetch('https://api.openai.com/v1/responses', {
			method: 'POST',
			headers: { authorization: `Bearer ${env.SUPPORT_AI_API_KEY}`, 'content-type': 'application/json' },
			body: JSON.stringify({ model: env.SUPPORT_AI_MODEL, store: false, max_output_tokens: 250,
				instructions: `You are ZENRES Assist. Answer only from the reviewed help text below, in ${language === 'hi' ? 'Hindi' : 'English'}. If it does not answer the question, say you cannot verify it and advise a support ticket. Never provide legal, tax, regulatory, emissions-verification or account-specific advice. Never invent service availability. Reviewed help:\n${guide}`,
				input: `Explain the reviewed help topic “${leading.title[language]}” in clear, concise language.` }),
			signal: AbortSignal.timeout(7000)
		});
		if (!response.ok) return Response.json(base);
		const result = await response.json();
		const answer = result.output?.flatMap((o: { content?: { type?: string; text?: string }[] }) => o.content ?? []).filter((c: { type?: string }) => c.type === 'output_text').map((c: { text?: string }) => c.text ?? '').join('\n').trim();
		if (!answer || answer.length > 1600) return Response.json(base);
		return Response.json({ answer, source: leading.id, mode: 'ai', escalate });
	} catch { return Response.json(base); }
};
