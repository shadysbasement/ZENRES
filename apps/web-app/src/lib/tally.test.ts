import { describe, it, expect } from 'vitest';
import { parseTallyCsv, autoMap, upsertVouchers, SAMPLE_TALLY_CSV, normaliseDate } from './tally';

describe('Tally CSV import', () => {
	const { headers, rows } = parseTallyCsv(SAMPLE_TALLY_CSV);
	const map = autoMap(headers);

	it('parses quoted Indian-format amounts', () => {
		expect(rows).toHaveLength(6);
		expect(rows[0][4]).toBe('21,84,000');
	});
	it('auto-maps Tally Day Book columns', () => {
		expect(map['Voucher No']).toBe('ref');
		expect(map['Particulars']).toBe('counterparty');
		expect(map['Amount']).toBe('amount');
	});
	it('is idempotent: importing the same file twice yields zero duplicates', () => {
		const first = upsertVouchers([], headers, rows, map, 'tally_csv');
		expect(first.inserted).toBe(6);
		const second = upsertVouchers(first.vouchers, headers, rows, map, 'tally_csv');
		expect(second.inserted).toBe(0);
		expect(second.duplicates).toBe(6);
		expect(second.vouchers).toHaveLength(6);
	});
	it('normalises Tally dates', () => {
		expect(normaliseDate('21-Nov-2026')).toBe('2026-11-21');
		expect(normaliseDate('05/03/26')).toBe('2026-03-05');
	});
});
