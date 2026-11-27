import { PMREMGenerator, type Texture, type WebGLRenderer } from 'three';
import { RoomEnvironment } from 'three/examples/jsm/environments/RoomEnvironment.js';

/** Soft-clay maquette palette tuned to the ZENRES blue & lilac glass theme. */
export const C = {
	clay: '#f7f9fd',
	clay2: '#eaf0f8',
	clay3: '#e4ebf5',
	plinth: '#e2eaf4',
	plinthEdge: '#c3d1e4',
	road: '#ccd8e8',
	ink: '#2c3137',
	blue: '#436c9b',
	blueSoft: '#a4d4f7',
	blueMid: '#6ab8ff',
	lilac: '#8265a8',
	lilacSoft: '#cbb5ed',
	lilacMid: '#cfa3f6',
	glass: '#dcecff',
	solar: '#56698a',
	mint: '#d6e8f4'
} as const;

/** Studio reflections without fetching an HDR: three's RoomEnvironment pre-filtered once. */
export function studioEnvironment(renderer: WebGLRenderer): Texture {
	const pmrem = new PMREMGenerator(renderer);
	const tex = pmrem.fromScene(new RoomEnvironment(), 0.04).texture;
	pmrem.dispose();
	return tex;
}

/** Deterministic 9×9 QR-like pattern with the three finder squares. */
export function qrCells(n = 9): [number, number][] {
	const out: [number, number][] = [];
	const finder = (x: number, y: number) => {
		for (const [fx, fy] of [
			[0, 0],
			[n - 3, 0],
			[0, n - 3]
		])
			if (x >= fx && x < fx + 3 && y >= fy && y < fy + 3) return !(x === fx + 1 && y === fy + 1);
		return null;
	};
	for (let y = 0; y < n; y++)
		for (let x = 0; x < n; x++) {
			const f = finder(x, y);
			if (f === true || (f === null && ((x * 7 + y * 13 + x * y) % 5) < 2)) out.push([x, y]);
		}
	return out;
}

export const easeOutBack = (t: number) => {
	const c1 = 1.4;
	const c3 = c1 + 1;
	return 1 + c3 * Math.pow(t - 1, 3) + c1 * Math.pow(t - 1, 2);
};
export const easeOutExpo = (t: number) => (t >= 1 ? 1 : 1 - Math.pow(2, -10 * t));
export const clamp01 = (t: number) => Math.min(1, Math.max(0, t));

export function webglAvailable(): boolean {
	try {
		const c = document.createElement('canvas');
		return !!(c.getContext('webgl2') || c.getContext('webgl'));
	} catch {
		return false;
	}
}
