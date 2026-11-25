#!/usr/bin/env node
/**
 * Design-system guardrail (Phase 1 acceptance): fails if any yellow / gold / amber colour or a banned
 * generic font (Inter, Roboto, Poppins, Space Grotesk) appears in source.
 */
import { readFileSync, readdirSync, statSync } from 'node:fs';
import { join, extname } from 'node:path';

const ROOTS = ['apps/web-public/src', 'apps/web-app/src', 'packages/design/src', 'templates'];
const EXT = new Set(['.svelte', '.css', '.ts', '.js', '.typ', '.html']);
const BANNED_FONTS = /\b(Inter|Roboto|Poppins|Space Grotesk)\b/;
const NAMED = /\b(yellow|gold|goldenrod|amber)\b/i;
const HEX = /#([0-9a-f]{6}|[0-9a-f]{3})\b/gi;

function hexToHsl(hex) {
	const h = hex.length === 3 ? hex.split('').map((c) => c + c).join('') : hex;
	const [r, g, b] = [0, 2, 4].map((i) => parseInt(h.slice(i, i + 2), 16) / 255);
	const max = Math.max(r, g, b), min = Math.min(r, g, b);
	const l = (max + min) / 2;
	if (max === min) return { h: 0, s: 0, l };
	const d = max - min;
	const s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
	let hue = max === r ? (g - b) / d + (g < b ? 6 : 0) : max === g ? (b - r) / d + 2 : (r - g) / d + 4;
	return { h: hue * 60, s, l };
}
// Yellow/amber/gold band: hue 38°–68° with meaningful saturation.
const isYellow = (hex) => {
	const { h, s, l } = hexToHsl(hex);
	return h >= 38 && h <= 68 && s > 0.35 && l > 0.2 && l < 0.92;
};

const problems = [];
function walk(dir) {
	for (const name of readdirSync(dir)) {
		if (name === 'node_modules' || name.startsWith('.')) continue;
		const p = join(dir, name);
		if (statSync(p).isDirectory()) walk(p);
		else if (EXT.has(extname(p))) check(p);
	}
}
function check(file) {
	readFileSync(file, 'utf8').split('\n').forEach((line, i) => {
		if (line.includes('palette-ok')) return;
		if (BANNED_FONTS.test(line) && !line.includes('check-palette')) problems.push(`${file}:${i + 1} banned font: ${line.trim()}`);
		const isComment = /^\s*(\*|\/\/|\/\*|<!--)/.test(line);
		if (!isComment && NAMED.test(line)) problems.push(`${file}:${i + 1} named yellow/gold/amber: ${line.trim()}`);
		for (const m of line.matchAll(HEX)) if (isYellow(m[1])) problems.push(`${file}:${i + 1} yellow-family hex #${m[1]}`);
	});
}
ROOTS.forEach((r) => {
	try {
		walk(r);
	} catch {
		/* root may not exist */
	}
});
if (problems.length) {
	console.error(`✗ Palette/font guardrail failed (${problems.length}):\n` + problems.join('\n'));
	process.exit(1);
}
console.log('✓ Zero yellow / gold / amber and zero banned fonts across', ROOTS.join(', '));
