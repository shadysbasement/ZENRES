/**
 * Sanitised public passport fixtures mirroring `cmd/seed-pilot` (Deccan Alloy Works).
 * Used only when the core API is unreachable and PUBLIC_DEMO_MODE is enabled.
 * The real resolver is GET /api/v1/public/verify/{type}/{token} on the core API.
 */
export type PublicBatch = {
	type: 'p';
	token: string;
	status: 'active' | 'reprinted' | 'revoked';
	manufacturer: { name: string; city: string; country: string; verified: boolean };
	product: { name: string; cn_code: string; sector: string };
	batch: { code: string; produced_on: string; stage: string };
	carbon: { intensity_tco2e_per_t: number | null; data_quality: string; badge: string | null };
	verification_tier: 'Self-Declared' | 'Document-Reviewed' | 'Independently Verified';
	anchor: { anchored: boolean; anchored_at?: string; merkle_root?: string };
	container?: { code: string; token: string };
};
export type PublicContainer = {
	type: 'c';
	token: string;
	status: 'active' | 'reprinted' | 'revoked';
	manufacturer: { name: string; city: string; country: string; verified: boolean };
	container: { code: string; dispatch_status: string; route: string[]; milestones: { label: string; at: string | null; done: boolean }[] };
	batch_count: number;
	/** Linked batches & documents require partner authentication. */
	partner_gated: true;
	anchor: { anchored: boolean; anchored_at?: string; merkle_root?: string };
};

const mfr = { name: 'Deccan Alloy Works Pvt. Ltd.', city: 'Pune, Maharashtra', country: 'IN', verified: true };
const anchor = { anchored: true, anchored_at: '2026-11-18T09:42:00Z', merkle_root: '9f2c4be1d07a3e55c1f8a2b64d9e0c7731aa58e2f6b4d1c09e3a7f25b8c6e41a' };

export const PASSPORTS: Record<string, PublicBatch | PublicContainer> = {
	'01JD8ZV001CNTR7Q4M2X9KPWBA': {
		type: 'c',
		token: '01JD8ZV001CNTR7Q4M2X9KPWBA',
		status: 'active',
		manufacturer: mfr,
		container: {
			code: 'ZV-001',
			dispatch_status: 'Ready to dispatch',
			route: ['Pune', 'JNPT Nhava Sheva', 'Rotterdam'],
			milestones: [
				{ label: 'Batches packed & sealed', at: '2026-11-17T14:10:00Z', done: true },
				{ label: 'Container QR anchored', at: '2026-11-18T09:42:00Z', done: true },
				{ label: 'Gate-in at JNPT', at: null, done: false },
				{ label: 'Vessel departure', at: null, done: false },
				{ label: 'Arrival Rotterdam', at: null, done: false }
			]
		},
		batch_count: 3,
		partner_gated: true,
		anchor
	},
	'01JD8ZA101BTCH3N6R1VQ8MZTE': batch('01JD8ZA101BTCH3N6R1VQ8MZTE', 'A-101', 'Aluminium bars (6063-T6)', '76042910', '2026-11-09', 5.38, 'Measured', 'Document-Reviewed'),
	'01JD8ZA102BTCH9H2K7WD4FXRC': batch('01JD8ZA102BTCH9H2K7WD4FXRC', 'A-102', 'Aluminium bars (6063-T6)', '76042910', '2026-11-11', 5.41, 'Measured', 'Document-Reviewed'),
	'01JD8ZA103BTCH5T8P3LJ6NYGS': batch('01JD8ZA103BTCH5T8P3LJ6NYGS', 'A-103', 'Extruded profiles (6061)', '76042100', '2026-11-13', 5.62, 'Supplier-Reported', 'Self-Declared'),
	'01JD8ZA099BTCHREVOKED0000X': { ...batch('01JD8ZA099BTCHREVOKED0000X', 'A-099', 'Aluminium bars (6063-T6)', '76042910', '2026-10-02', 5.5, 'Measured', 'Self-Declared'), status: 'revoked' }
};

function batch(token: string, code: string, name: string, cn: string, date: string, intensity: number, quality: string, tier: PublicBatch['verification_tier']): PublicBatch {
	return {
		type: 'p',
		token,
		status: 'active',
		manufacturer: mfr,
		product: { name, cn_code: cn, sector: 'Aluminium' },
		batch: { code, produced_on: date, stage: 'Ready to dispatch' },
		carbon: { intensity_tco2e_per_t: intensity, data_quality: quality, badge: quality === 'Measured' ? 'Measured embedded emissions' : null },
		verification_tier: tier,
		anchor,
		container: { code: 'ZV-001', token: '01JD8ZV001CNTR7Q4M2X9KPWBA' }
	};
}
