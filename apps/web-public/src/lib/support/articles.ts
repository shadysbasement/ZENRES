export type SupportLanguage = 'en' | 'hi';
export type SupportArticle = {
	id: string;
	category: string;
	title: Record<SupportLanguage, string>;
	answer: Record<SupportLanguage, string>;
	keywords: string[];
	escalate?: boolean;
};

/** Reviewed product guidance only. No customer records, tax advice or unverifiable service promises. */
export const articles: SupportArticle[] = [
	{
		id: 'tally', category: 'Integrations',
		title: { en: 'Connecting Tally.ERP 9', hi: 'Tally.ERP 9 से कनेक्ट करना' },
		answer: { en: 'The workspace supports a Tally CSV import workflow. Export your data from Tally, review the columns and upload the CSV in Integrations → Tally. The integration is currently a workspace demo; check the feature availability page before relying on a production sync.', hi: 'वर्कस्पेस में Tally CSV आयात का तरीका है। Tally से CSV निर्यात करें और Integrations → Tally में कॉलम जाँचें। अभी यह वर्कस्पेस डेमो है; उत्पादन उपयोग से पहले सुविधा की उपलब्धता जाँचें।' },
		keywords: ['tally', 'erp', 'csv', 'sync', 'integration']
	},
	{
		id: 'cbam', category: 'Carbon',
		title: { en: 'Understanding CBAM calculations', hi: 'CBAM गणना को समझना' },
		answer: { en: 'ZENRES presents illustrative default-versus-actual emissions comparisons and reporting support. Final values depend on verified activity data, product classification and applicable rules. Have a qualified professional verify any declaration before submission.', hi: 'ZENRES उदाहरण के तौर पर डिफ़ॉल्ट और वास्तविक उत्सर्जन की तुलना तथा रिपोर्टिंग सहायता दिखाता है। अंतिम घोषणा से पहले योग्य विशेषज्ञ से डेटा और नियमों का सत्यापन कराएँ।' },
		keywords: ['cbam', 'carbon', 'emissions', 'calculation', 'tax', 'declaration'], escalate: true
	},
	{
		id: 'qr', category: 'Traceability',
		title: { en: 'Generating product and container QR codes', hi: 'उत्पाद और कंटेनर QR कोड बनाना' },
		answer: { en: 'Open QR Studio in the workspace to explore product and container passports. The workspace currently uses sample records. A QR code identifies the linked record; it does not itself prove that source data is accurate.', hi: 'वर्कस्पेस में QR Studio खोलें और उत्पाद तथा कंटेनर पासपोर्ट देखें। अभी नमूना रिकॉर्ड उपयोग होते हैं। QR कोड स्रोत डेटा की शुद्धता का प्रमाण नहीं है।' },
		keywords: ['qr', 'passport', 'product', 'container', 'trace']
	},
	{
		id: 'freight', category: 'Logistics',
		title: { en: 'Managing shipments and logistics', hi: 'शिपमेंट और लॉजिस्टिक्स का प्रबंधन' },
		answer: { en: 'FreightLink demonstrates return-load matching, shipment milestones and proof-of-delivery flows. These are sample workspace records today. Transport is priced separately from a ZENRES subscription.', hi: 'FreightLink वापसी लोड मिलान, शिपमेंट चरण और डिलीवरी प्रमाण के उदाहरण दिखाता है। अभी वर्कस्पेस में नमूना रिकॉर्ड हैं। परिवहन शुल्क सदस्यता से अलग है।' },
		keywords: ['freight', 'shipment', 'logistics', 'transport', 'truck', 'delivery']
	},
	{
		id: 'marketplace', category: 'Marketplace',
		title: { en: 'Marketplace and buyer enquiries', hi: 'मार्केटप्लेस और खरीदार संबंधी प्रश्न' },
		answer: { en: 'ZENRES Connect shows an illustrative buyer and supplier discovery flow. Named buyer ecosystems on the site are targets, not confirmed partnerships or endorsements.', hi: 'ZENRES Connect खरीदार और आपूर्तिकर्ता खोज का उदाहरण दिखाता है। वेबसाइट पर दिखाए गए खरीदार संभावित लक्ष्य हैं, पुष्टि किए गए भागीदार नहीं।' },
		keywords: ['marketplace', 'buyer', 'supplier', 'connect']
	},
	{
		id: 'billing', category: 'Account',
		title: { en: 'Billing and subscriptions', hi: 'बिलिंग और सदस्यता' },
		answer: { en: 'The current public pricing is $99/month for MSME Starter, $249/month for MSME Growth and $499/month for EU Importer Hub, excluding applicable taxes. Billing is not yet active in the demo workspace. Contact the team through a support ticket for an account-specific question.', hi: 'वर्तमान सार्वजनिक मूल्य MSME Starter के लिए $99/माह, Growth के लिए $249/माह और EU Importer Hub के लिए $499/माह हैं; लागू कर अलग हैं। डेमो वर्कस्पेस में बिलिंग अभी सक्रिय नहीं है।' },
		keywords: ['billing', 'price', 'subscription', 'plan', 'payment']
	},
	{
		id: 'privacy', category: 'Security',
		title: { en: 'Data security and privacy', hi: 'डेटा सुरक्षा और गोपनीयता' },
		answer: { en: 'Support tickets are visible only to their requester and authorised support staff. Do not paste financial records, passwords or private keys into the assistant. Support access to financial records would require a separate, explicit authorisation flow.', hi: 'सहायता टिकट केवल अनुरोधकर्ता और अधिकृत सहायता कर्मचारियों को दिखते हैं। सहायक में वित्तीय रिकॉर्ड, पासवर्ड या निजी कुंजी साझा न करें। वित्तीय रिकॉर्ड तक पहुँच के लिए अलग स्पष्ट अनुमति आवश्यक है।' },
		keywords: ['privacy', 'security', 'data', 'financial', 'access']
	},
	{
		id: 'help', category: 'Getting started',
		title: { en: 'Getting help from a person', hi: 'किसी व्यक्ति से सहायता लेना' },
		answer: { en: 'Sign in to create a ticket or start a support conversation. A human can reply in the ticket thread when the support team is available. No response-time commitment or staffed live-chat hours have been published.', hi: 'टिकट बनाने या सहायता बातचीत शुरू करने के लिए साइन इन करें। टीम उपलब्ध होने पर कोई व्यक्ति टिकट में जवाब दे सकता है। अभी प्रतिक्रिया समय या लाइव चैट के घंटे प्रकाशित नहीं हुए हैं।' },
		keywords: ['human', 'chat', 'ticket', 'support', 'agent', 'help']
	}
];

export function findArticles(query: string): SupportArticle[] {
	const words = query.toLowerCase().match(/[\p{L}\p{N}]{2,}/gu) ?? [];
	if (!words.length) return articles;
	return articles.map((article) => {
		const haystack = [article.title.en, article.title.hi, article.category, ...article.keywords].join(' ').toLowerCase();
		const score = words.reduce((sum, word) => sum + (haystack.includes(word) ? 1 : 0), 0);
		return { article, score };
	}).filter((hit) => hit.score > 0).sort((a, b) => b.score - a.score).map((hit) => hit.article);
}
