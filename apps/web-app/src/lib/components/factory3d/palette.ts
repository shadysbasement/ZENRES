export type TwinMode = 'studio' | 'cad';
export const PAL = {
	studio: { bg: '#0E1412', floor: '#141C19', slab: '#1C2622', edge: '#2F3F38', body: '#26332E', body2: '#33443D', glass: '#5d8f7c', accent: '#38B884', crate: '#4A5A53', grid: '#1F2A26' },
	cad: { bg: '#F7F8F7', floor: '#EEF2F0', slab: '#FFFFFF', edge: '#9AA7A1', body: '#E6EBE8', body2: '#D5DDD9', glass: '#9cc4b4', accent: '#14533C', crate: '#C5CFCA', grid: '#DCE3DF' }
} as const;
export const EMERALD = '#38B884';
export const TERRACOTTA = '#C05A3C';
export const COBALT = '#1E4D6B';
export type Pal = (typeof PAL)['studio'] | (typeof PAL)['cad'];
