export const BACKGROUNDS = [
	{ id: 'silk', name: 'Silk', description: 'Sculpted blue & lilac ribbons' },
	{ id: 'aurora', name: 'Aurora', description: 'A soft wash of light' },
	{ id: 'contour', name: 'Contour', description: 'Quiet architectural lines' },
	{ id: 'rose', name: 'Rosewater', description: 'A warmer point of view' },
	{ id: 'mist', name: 'Mist', description: 'Pure, understated focus' }
] as const;
export type Background = typeof BACKGROUNDS[number]['id'];
const STORAGE_KEY = 'zenres-appearance-v1';

class Appearance {
	background = $state<Background>('silk');
	ambient = $state(true);
	reducedMotion = $state(false);
	get motion() { return !this.reducedMotion; }
	init() {
		const preference = window.matchMedia('(prefers-reduced-motion: reduce)');
		const update = () => { this.reducedMotion = preference.matches; };
		update();
		preference.addEventListener('change', update);
		try {
			const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) ?? 'null');
			if (BACKGROUNDS.some((b) => b.id === saved?.background)) this.background = saved.background;
			if (typeof saved?.ambient === 'boolean') this.ambient = saved.ambient;
		} catch { /* Storage can be unavailable in private browsing. */ }
		return () => preference.removeEventListener('change', update);
	}
	setBackground(background: Background) { this.background = background; this.save(); }
	setAmbient(ambient: boolean) { this.ambient = ambient; this.save(); }
	private save() {
		try { localStorage.setItem(STORAGE_KEY, JSON.stringify({ background: this.background, ambient: this.ambient })); } catch { /* The current session still works without storage. */ }
	}
}
export const appearance = new Appearance();
