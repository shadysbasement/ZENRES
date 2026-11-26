import type { LayoutServerLoad } from './$types';
import { legalConfig } from '$lib/server/legal';

export const load: LayoutServerLoad = () => ({ legalAvailable: legalConfig().visible });
