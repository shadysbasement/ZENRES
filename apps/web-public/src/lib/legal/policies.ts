export type LegalSlug = 'terms' | 'privacy' | 'cookies' | 'disclaimer' | 'acceptable-use' | 'refunds';

export const LEGAL_POLICIES: Record<LegalSlug, { title: string; intro: string; sections: { heading: string; text: string }[] }> = {
	terms: {
		title: 'Terms and Conditions',
		intro: 'These draft terms describe use of the ZENRES business workspace. The legal entity, commercial terms, liability provisions and governing law require review before these terms take effect.',
		sections: [
			{ heading: 'Business accounts', text: 'The workspace is intended for people authorised to act for their organisations. Account holders should provide accurate details, protect their credentials and report suspected unauthorised access.' },
			{ heading: 'Customer information', text: 'Customers retain their rights in information they upload. ZENRES needs permission to process that information to provide the services the customer chooses. A user must have authority to upload or share business records, evidence and shipment details.' },
			{ heading: 'Available services', text: 'ZENRES may provide business data, emissions calculation, traceability and trade workflow tools. Availability depends on the plan, deployment and enabled features. A demonstration or preview does not promise that a transaction, integration or paid service is available.' },
			{ heading: 'Calculations and records', text: 'Outputs depend on source data and selected methods. An estimate is not independent verification or a guarantee of regulatory acceptance. QR links and any future integrity proof do not establish the accuracy of original inputs or the physical authenticity of goods.' },
			{ heading: 'Transactions', text: 'Where marketplace or logistics features are enabled, the parties to a transaction remain responsible for their own contracts, goods, delivery, customs and tax obligations. ZENRES must disclose its actual role and any applicable fee before a transaction is confirmed.' },
			{ heading: 'Further legal terms', text: 'Termination, suspension, liability limits, governing law, dispute resolution and mandatory statutory rights need approved wording before this draft can be published.' }
		]
	},
	privacy: {
		title: 'Privacy Notice',
		intro: 'This draft describes data categories visible in the current ZENRES product. Processing roles, vendors, transfer arrangements, retention periods and request handling must be confirmed before publication.',
		sections: [
			{ heading: 'Information handled', text: 'Account creation may use a name, work email, authentication credentials and organisation details. Depending on features used, the service may handle ERP records, uploaded evidence, shipment information, support messages and technical security logs.' },
			{ heading: 'Why information is used', text: 'Information is used to create and secure accounts, operate the requested workspace features, maintain records and respond to support or security requests. Any additional marketing use requires a separate choice.' },
			{ heading: 'Sharing and storage', text: 'The final notice must name or clearly describe actual hosting and service providers, subprocessors, storage locations, international transfers and safeguards. Customer data must only be shared for the agreed service or another valid purpose.' },
			{ heading: 'Retention and requests', text: 'Retention periods and any exceptions for legal obligations remain to be approved. The final notice must give a working privacy contact and explain how people can request access, correction, deletion or other rights available to them.' },
			{ heading: 'Security', text: 'The product uses account authentication and access controls. The final notice must accurately describe deployed protections and avoid promising controls that have not been verified in production.' }
		]
	},
	cookies: {
		title: 'Cookies and Device Storage',
		intro: 'The current code uses essential authentication cookies and local device preferences. No optional analytics or marketing tracker was found in this repository. This inventory must be checked against the deployed site before publication.',
		sections: [
			{ heading: 'Essential cookies', text: 'The authentication service uses a session cookie and a CSRF cookie to keep a signed-in session secure. The restricted administrator workspace uses its own session cookie. These are needed for the requested features.' },
			{ heading: 'Device preferences', text: 'The workspace saves display choices such as background, ambient motion, currency, demo tier and dismissed hints in local storage. These stay on this device until removed through browser storage controls.' },
			{ heading: 'Optional technologies', text: 'Optional analytics and marketing trackers must remain off until they are inventoried, disclosed and any required choice is obtained. If they are added, the site must offer equally accessible accept, reject and change-preference controls.' }
		]
	},
	disclaimer: {
		title: 'Product Disclaimer',
		intro: 'ZENRES supports business reporting and workflow decisions. This draft clarifies the limits of generated estimates and previews.',
		sections: [
			{ heading: 'Carbon and CBAM', text: 'Calculations reflect supplied data, selected emission factors and methods. Estimated values are not independent verification. ZENRES does not guarantee CBAM acceptance, tariff savings or compliance. Customers and relevant importers remain responsible for their legal obligations and any required qualified verification.' },
			{ heading: 'Traceability', text: 'A QR code points to a digital record and may be subject to access controls. A future blockchain anchor, if enabled, could support a record integrity check after anchoring; it cannot prove that the original data or physical goods were genuine.' },
			{ heading: 'Marketplace and freight', text: 'Listings, matches, costs and route estimates do not guarantee a buyer, order, carrier, delivery time or savings. Transaction parties should verify commercial and legal requirements before committing.' }
		]
	},
	'acceptable-use': {
		title: 'Acceptable Use Policy',
		intro: 'This draft sets expected conduct for a shared business platform. Enforcement and appeal procedures need final review.',
		sections: [
			{ heading: 'Accurate information', text: 'Do not create fraudulent listings, false certifications, misleading emissions evidence or claims about goods that cannot be supported.' },
			{ heading: 'Lawful access and content', text: 'Do not upload unlawful material, infringe another party’s rights, access another organisation’s data without permission, interfere with the service or attempt to bypass its security controls.' },
			{ heading: 'Reporting and response', text: 'A working reporting channel must be published before this policy takes effect. ZENRES may investigate reported misuse and use a proportionate restriction or suspension process under approved terms.' }
		]
	},
	refunds: {
		title: 'Subscriptions and Refunds',
		intro: 'This is a pre-launch draft. Payment terms, cancellation steps, refund eligibility and taxes require an approved commercial policy before checkout can be enabled.',
		sections: [
			{ heading: 'Before purchase', text: 'A checkout must show the full price, applicable taxes, billing frequency, renewal terms, usage limits and any service fees before a customer confirms payment.' },
			{ heading: 'Cancellation and refunds', text: 'The final policy must state how to cancel, when cancellation takes effect, whether any refund is available and how a customer receives confirmation. No refund window or fee is assumed in this draft.' }
		]
	}
};

export const LEGAL_LINKS: { slug: LegalSlug; label: string }[] = [
	{ slug: 'terms', label: 'Terms' }, { slug: 'privacy', label: 'Privacy' },
	{ slug: 'cookies', label: 'Cookies' }, { slug: 'disclaimer', label: 'Disclaimer' },
	{ slug: 'acceptable-use', label: 'Acceptable use' }, { slug: 'refunds', label: 'Refunds' }
];
