import { dev } from '$app/environment';
import { env } from '$env/dynamic/private';

const required = [
	'ZENRES_LEGAL_OPERATOR_NAME', 'ZENRES_LEGAL_CONTACT_ADDRESS', 'ZENRES_LEGAL_SUPPORT_CONTACT',
	'ZENRES_LEGAL_PRIVACY_CONTACT', 'ZENRES_LEGAL_JURISDICTION', 'ZENRES_LEGAL_RETENTION_SUMMARY',
	'ZENRES_LEGAL_TERMS_VERSION', 'ZENRES_LEGAL_PRIVACY_VERSION', 'ZENRES_LEGAL_EFFECTIVE_DATE'
] as const;

// Legal copy remains a review draft until a reviewed revision changes this gate.
const LEGAL_COPY_APPROVED = false;

export function legalConfig() {
	const missing = required.filter((name) => !env[name]?.trim());
	const published = LEGAL_COPY_APPROVED && env.ZENRES_LEGAL_APPROVED === 'true' && missing.length === 0;
	return {
		published,
		visible: published || dev,
		missing,
		operator: env.ZENRES_LEGAL_OPERATOR_NAME || 'Operator to be confirmed',
		address: env.ZENRES_LEGAL_CONTACT_ADDRESS || 'Contact address to be confirmed',
		support: env.ZENRES_LEGAL_SUPPORT_CONTACT || 'Support contact to be confirmed',
		privacy: env.ZENRES_LEGAL_PRIVACY_CONTACT || 'Privacy contact to be confirmed',
		jurisdiction: env.ZENRES_LEGAL_JURISDICTION || 'Jurisdiction to be confirmed',
		retention: env.ZENRES_LEGAL_RETENTION_SUMMARY || 'Retention periods to be confirmed',
		termsVersion: env.ZENRES_LEGAL_TERMS_VERSION || 'draft-1',
		privacyVersion: env.ZENRES_LEGAL_PRIVACY_VERSION || 'draft-1',
		effectiveDate: env.ZENRES_LEGAL_EFFECTIVE_DATE || 'Not effective'
	};
}
