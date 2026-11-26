export type TwinMode = 'studio' | 'cad';
export const PAL = {
	studio: { bg: '#202B3D', floor: '#29374B', slab: '#33445B', edge: '#536986', body: '#405670', body2: '#4B6481', glass: '#91B4DB', accent: '#99C7F2', crate: '#687F9D', grid: '#38495F' },
	cad: { bg: '#E7EEF7', floor: '#DCE6F2', slab: '#FFFFFF', edge: '#8D9FBA', body: '#D0DDED', body2: '#BDCCE2', glass: '#A4C9EA', accent: '#7D97C8', crate: '#C2CFE3', grid: '#CEDBEA' }
} as const;
export const EMERALD = '#99C7F2';
export const TERRACOTTA = '#C3899D';
export const COBALT = '#9B82C9';
export type Pal = (typeof PAL)['studio'] | (typeof PAL)['cad'];
