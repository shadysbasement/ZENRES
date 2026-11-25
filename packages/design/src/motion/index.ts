/**
 * ZENRES motion primitives — framework-light Svelte actions.
 * Every primitive is a no-op (content shown immediately) under prefers-reduced-motion.
 */

export const prefersReducedMotion = (): boolean =>
	typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches;

const finePointer = (): boolean =>
	typeof window !== 'undefined' && window.matchMedia('(hover: hover) and (pointer: fine)').matches;

/* ------------------------------------------------------------------ */
/* reveal — adds .is-in when the node enters the viewport              */
/* ------------------------------------------------------------------ */

type RevealOptions = { once?: boolean; threshold?: number; rootMargin?: string; delay?: number } | undefined;

let sharedObserver: IntersectionObserver | null = null;
const callbacks = new WeakMap<Element, (entry: IntersectionObserverEntry) => void>();

function observer(): IntersectionObserver {
	if (!sharedObserver) {
		sharedObserver = new IntersectionObserver(
			(entries) => {
				for (const entry of entries) callbacks.get(entry.target)?.(entry);
			},
			{ threshold: 0.12, rootMargin: '0px 0px -8% 0px' }
		);
	}
	return sharedObserver;
}

export function reveal(node: HTMLElement, opts: RevealOptions = {}) {
	const once = opts?.once ?? true;
	if (opts?.delay) node.style.setProperty('--delay', `${opts.delay}ms`);
	if (!node.hasAttribute('data-reveal')) node.setAttribute('data-reveal', '');
	if (prefersReducedMotion()) {
		node.classList.add('is-in');
		return {};
	}
	const io = observer();
	callbacks.set(node, (entry) => {
		if (entry.isIntersecting) {
			node.classList.add('is-in');
			if (once) {
				io.unobserve(node);
				callbacks.delete(node);
			}
		} else if (!once) {
			node.classList.remove('is-in');
		}
	});
	io.observe(node);
	return {
		destroy() {
			io.unobserve(node);
			callbacks.delete(node);
		}
	};
}

/** inview — toggles .is-in on a container so nested .split-inner children animate. */
export function inview(node: HTMLElement, opts: RevealOptions = {}) {
	if (opts?.delay) node.style.setProperty('--delay', `${opts.delay}ms`);
	if (prefersReducedMotion()) {
		node.classList.add('is-in');
		return {};
	}
	const io = observer();
	callbacks.set(node, (entry) => {
		if (entry.isIntersecting) {
			node.classList.add('is-in');
			io.unobserve(node);
		}
	});
	io.observe(node);
	return {
		destroy() {
			io.unobserve(node);
			callbacks.delete(node);
		}
	};
}

/* ------------------------------------------------------------------ */
/* magnetic — element drifts toward the pointer                        */
/* ------------------------------------------------------------------ */

export function magnetic(node: HTMLElement, strength = 0.28) {
	if (!finePointer() || prefersReducedMotion()) return {};
	let raf = 0;
	let tx = 0,
		ty = 0,
		cx = 0,
		cy = 0;
	node.style.willChange = 'transform';
	const loop = () => {
		cx += (tx - cx) * 0.18;
		cy += (ty - cy) * 0.18;
		node.style.transform = `translate3d(${cx.toFixed(2)}px, ${cy.toFixed(2)}px, 0)`;
		if (Math.abs(tx - cx) > 0.05 || Math.abs(ty - cy) > 0.05) raf = requestAnimationFrame(loop);
		else raf = 0;
	};
	const kick = () => {
		if (!raf) raf = requestAnimationFrame(loop);
	};
	const move = (e: PointerEvent) => {
		const r = node.getBoundingClientRect();
		tx = (e.clientX - (r.left + r.width / 2)) * strength;
		ty = (e.clientY - (r.top + r.height / 2)) * strength;
		kick();
	};
	const leave = () => {
		tx = 0;
		ty = 0;
		kick();
	};
	node.addEventListener('pointermove', move);
	node.addEventListener('pointerleave', leave);
	return {
		destroy() {
			cancelAnimationFrame(raf);
			node.removeEventListener('pointermove', move);
			node.removeEventListener('pointerleave', leave);
		}
	};
}

/* ------------------------------------------------------------------ */
/* scramble — technical text decode effect on hover / mount            */
/* ------------------------------------------------------------------ */

const GLYPHS = 'ABCDEFGHJKLMNPQRSTUVWXYZ0123456789/<>-_#';

export function scrambleText(node: HTMLElement, opts: { onMount?: boolean; duration?: number } = {}) {
	const original = node.textContent ?? '';
	const duration = opts.duration ?? 520;
	let raf = 0;
	const run = () => {
		if (prefersReducedMotion()) return;
		cancelAnimationFrame(raf);
		const start = performance.now();
		const text = node.dataset.text ?? original;
		const step = (now: number) => {
			const p = Math.min(1, (now - start) / duration);
			const settled = Math.floor(p * text.length);
			let out = '';
			for (let i = 0; i < text.length; i++) {
				const ch = text[i];
				if (i < settled || ch === ' ') out += ch;
				else out += GLYPHS[(Math.random() * GLYPHS.length) | 0];
			}
			node.textContent = out;
			if (p < 1) raf = requestAnimationFrame(step);
			else node.textContent = text;
		};
		raf = requestAnimationFrame(step);
	};
	const host = (node.closest('a,button') as HTMLElement | null) ?? node;
	host.addEventListener('pointerenter', run);
	if (opts.onMount) run();
	return {
		destroy() {
			cancelAnimationFrame(raf);
			host.removeEventListener('pointerenter', run);
		}
	};
}

/* ------------------------------------------------------------------ */
/* parallax — translate a node relative to scroll position             */
/* ------------------------------------------------------------------ */

export function parallax(node: HTMLElement, speed = 0.12) {
	if (prefersReducedMotion()) return {};
	let raf = 0;
	const update = () => {
		const r = node.parentElement?.getBoundingClientRect() ?? node.getBoundingClientRect();
		const center = r.top + r.height / 2 - window.innerHeight / 2;
		node.style.transform = `translate3d(0, ${(-center * speed).toFixed(2)}px, 0)`;
		raf = 0;
	};
	const onScroll = () => {
		if (!raf) raf = requestAnimationFrame(update);
	};
	update();
	window.addEventListener('scroll', onScroll, { passive: true });
	return {
		destroy() {
			cancelAnimationFrame(raf);
			window.removeEventListener('scroll', onScroll);
		}
	};
}

/* ------------------------------------------------------------------ */
/* tilt — subtle 3D tilt for cards                                     */
/* ------------------------------------------------------------------ */

export function tilt(node: HTMLElement, max = 4) {
	if (!finePointer() || prefersReducedMotion()) return {};
	node.style.transition = 'transform 600ms cubic-bezier(0.16, 1, 0.3, 1)';
	node.style.transformStyle = 'preserve-3d';
	const move = (e: PointerEvent) => {
		const r = node.getBoundingClientRect();
		const px = (e.clientX - r.left) / r.width - 0.5;
		const py = (e.clientY - r.top) / r.height - 0.5;
		node.style.transform = `perspective(1200px) rotateX(${(-py * max).toFixed(2)}deg) rotateY(${(px * max).toFixed(2)}deg)`;
		node.style.setProperty('--mx', `${((px + 0.5) * 100).toFixed(1)}%`);
		node.style.setProperty('--my', `${((py + 0.5) * 100).toFixed(1)}%`);
	};
	const leave = () => {
		node.style.transform = 'perspective(1200px) rotateX(0) rotateY(0)';
	};
	node.addEventListener('pointermove', move);
	node.addEventListener('pointerleave', leave);
	return {
		destroy() {
			node.removeEventListener('pointermove', move);
			node.removeEventListener('pointerleave', leave);
		}
	};
}

/* ------------------------------------------------------------------ */
/* countUp — animates numeric text content when visible                */
/* ------------------------------------------------------------------ */

export function tweenNumber(
	from: number,
	to: number,
	duration: number,
	onUpdate: (v: number) => void,
	ease: (t: number) => number = (t) => 1 - Math.pow(1 - t, 4)
) {
	if (prefersReducedMotion() || duration <= 0) {
		onUpdate(to);
		return () => {};
	}
	let raf = 0;
	const start = performance.now();
	const step = (now: number) => {
		const p = Math.min(1, (now - start) / duration);
		onUpdate(from + (to - from) * ease(p));
		if (p < 1) raf = requestAnimationFrame(step);
	};
	raf = requestAnimationFrame(step);
	return () => cancelAnimationFrame(raf);
}

/* ------------------------------------------------------------------ */
/* smooth scrolling (Lenis) — lazily imported so SSR stays clean       */
/* ------------------------------------------------------------------ */

export async function initSmoothScroll(): Promise<(() => void) | undefined> {
	if (typeof window === 'undefined' || prefersReducedMotion()) return;
	const { default: Lenis } = await import('lenis');
	const lenis = new Lenis({ duration: 1.15, easing: (t: number) => Math.min(1, 1.001 - Math.pow(2, -10 * t)), smoothWheel: true });
	(window as unknown as { __lenis?: unknown }).__lenis = lenis;
	let raf = 0;
	const loop = (time: number) => {
		lenis.raf(time);
		raf = requestAnimationFrame(loop);
	};
	raf = requestAnimationFrame(loop);
	return () => {
		cancelAnimationFrame(raf);
		lenis.destroy();
	};
}

export function scrollToTarget(target: string | HTMLElement, offset = -72) {
	const lenis = (window as unknown as { __lenis?: { scrollTo: (t: unknown, o: unknown) => void } }).__lenis;
	if (lenis) lenis.scrollTo(target, { offset, duration: 1.4 });
	else {
		const el = typeof target === 'string' ? document.querySelector(target) : target;
		el?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}
}

export function lockScroll(lock: boolean) {
	const lenis = (window as unknown as { __lenis?: { stop: () => void; start: () => void } }).__lenis;
	if (lenis) lock ? lenis.stop() : lenis.start();
	document.documentElement.style.overflow = lock ? 'hidden' : '';
}
