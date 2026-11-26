import type { LayoutServerLoad } from './$types';
import { legalConfig } from '$lib/server/legal';

export const load: LayoutServerLoad = () => {
	const legal = legalConfig();
	return { legalAvailable: legal.visible, legalPublished: legal.published, legalTermsVersion: legal.termsVersion, legalPrivacyVersion: legal.privacyVersion };
};
