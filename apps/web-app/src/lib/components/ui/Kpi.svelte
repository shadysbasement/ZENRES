<script lang="ts">
	import type { Snippet } from 'svelte';
	import Icon from './Icon.svelte';
	interface Props { label: string; value: string; delta?: string; tone?: 'up' | 'down' | 'flat'; href?: string; i?: number; children?: Snippet; icon?: string; spacious?: boolean; }
	let { label, value, delta, tone = 'flat', href, i = 0, children, icon, spacious = false }: Props = $props();
</script>

<svelte:element this={href ? 'a' : 'div'} {href} class="kpi card" class:spacious style="--i:{i}">
	<div class="top">{#if icon}<span class="metric-icon"><Icon name={icon} size={16}/></span>{/if}<span class="metric-label">{label}</span>{#if href}<span class="go" aria-hidden="true">↗</span>{/if}</div>
	<div class="metric-body"><div class="value-group"><div class="val num">{value}</div>{#if delta}<span class="delta {tone}">{delta}</span>{/if}</div>{#if children}<div class="extra">{@render children()}</div>{/if}</div>
</svelte:element>

<style>
	.kpi { display: flex; flex-direction: column; gap: 21px; padding: 20px; min-width: 0; border-radius: 20px; transition: transform 450ms var(--ease-expo), box-shadow 400ms, background 300ms; animation: k-in 650ms var(--ease-expo) backwards; animation-delay: calc(var(--i) * 65ms); }
	@keyframes k-in { from { opacity: 0; transform: translateY(15px) scale(.985); } }
	a.kpi:hover { transform: translateY(-4px); background: #fff; box-shadow: 0 12px 28px -16px #6a80a640; }
	a.kpi:active { transform: scale(.99); }
	.top { display: flex; align-items: center; justify-content: space-between; gap: 10px; min-height: 22px; }
	.metric-label { color: #4f5a67; font-size: 13px; font-weight: 500; flex: 1; }
	.metric-icon { width: 31px; height: 31px; border-radius: 50%; display: grid; place-items: center; background: linear-gradient(120deg, #e3f1fb, #ede6f7); color: #7b8dad; }
	.go { font-size: 17px; color: #9aa6b7; transition: transform 300ms var(--ease-out), color 300ms; }
	a.kpi:hover .go { transform: translate(2px, -2px); color: #7184b7; }
	.val { font-size: clamp(25px, 2.4vw, 37px); letter-spacing: -.055em; color: var(--ink-primary); line-height: 1.2; font-weight: 450; white-space: nowrap; }
	.delta { display: block; font-size: 11px; line-height: 1.5; color: var(--ink-muted); margin-top: 5px; }
	.delta.up { color: #528574; }
	.delta.down { color: #a86d70; }
	.extra { margin-top: 12px; min-width: 0; }
	.spacious { padding: 22px 26px 25px; min-height: 190px; justify-content: space-between; }
	.spacious .top { margin-bottom: 13px; }
	.spacious .metric-label { text-align: center; padding-right: 10px; font-size: 15px; }
	.spacious .metric-body { display: flex; align-items: flex-end; justify-content: space-between; gap: 22px; }
	.spacious .extra { flex: 1; margin: 0 0 5px; max-width: 160px; }
	.spacious .val { font-size: clamp(30px, 3.15vw, 46px); }
	@media (max-width: 850px) { .spacious { padding: 20px; min-height: 185px; } .spacious .metric-body { gap: 12px; } .spacious .extra { max-width: 130px; } }
	@media (max-width: 650px) { .kpi { padding: 17px; gap: 14px; } .metric-label { font-size: 12px; } .spacious { min-height: 173px; } .spacious .metric-label { font-size: 14px; } .spacious .val { font-size: 36px; } .spacious .extra { max-width: 160px; } }
</style>
