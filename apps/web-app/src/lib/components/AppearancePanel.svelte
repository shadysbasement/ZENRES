<script lang="ts">
	import Drawer from './ui/Drawer.svelte';
	import Icon from './ui/Icon.svelte';
	import { appearance, BACKGROUNDS } from '$lib/stores/appearance.svelte';
	let { open, onclose }: { open: boolean; onclose: () => void } = $props();
</script>

<Drawer {open} {onclose} title="Make space for your work" eyebrow="Workspace appearance" width={440}>
	<p class="intro">A little atmosphere. A clearer head. Choose a backdrop that feels like you.</p>
	<div class="options" role="group" aria-label="Workspace background">
		{#each BACKGROUNDS as bg (bg.id)}
			<button class="option" class:chosen={appearance.background === bg.id} aria-pressed={appearance.background === bg.id} onclick={() => appearance.setBackground(bg.id)}>
				<span class="preview {bg.id}"><span class="sample-card"></span><span class="sample-card two"></span>{#if appearance.background === bg.id}<span class="selected"><Icon name="check" size={12}/></span>{/if}</span>
				<strong>{bg.name}</strong><span class="description">{bg.description}</span>
			</button>
		{/each}
	</div>
	<div class="motion-row"><div><strong>Ambient motion</strong><p>Slow, gentle movement in the background.</p></div><button class="switch" role="switch" aria-checked={appearance.ambient} aria-label="Ambient motion" onclick={() => appearance.setAmbient(!appearance.ambient)}><span></span></button></div>
	{#if appearance.reducedMotion}<p class="fine">Your device’s reduced motion preference is on. Animations are paused.</p>{/if}
	<p class="saved"><Icon name="check" size={14}/> Saved automatically on this device</p>
</Drawer>

<style>
	.intro { font-size: 15px; line-height: 1.6; margin-bottom: 24px; color: var(--ink-muted); }
	.options { display: grid; grid-template-columns: 1fr 1fr; gap: 18px 14px; }
	.option { text-align: left; min-width: 0; }
	.preview { height: 106px; display: flex; gap: 6px; align-items: flex-end; padding: 12px; position: relative; overflow: hidden; border: 3px solid transparent; border-radius: 16px; box-shadow: 0 0 0 1px var(--border-hairline); transition: box-shadow 300ms, transform 400ms var(--ease-out); }
	.option:hover .preview { transform: translateY(-3px); }
	.chosen .preview { box-shadow: 0 0 0 2px #99a7e9; }
	.silk { background: linear-gradient(155deg, #dfe7f2 30%, #a6cdec 31%, #e5eefa 48%, #c0a9e6 55%, #e4e8f3 56%); }
	.aurora { background: radial-gradient(at 80% 20%, #c9afe9, transparent 70%), radial-gradient(at 20% 90%, #a9d8f5, #e2eaf3); }
	.contour { background: repeating-radial-gradient(ellipse at 100% 0%, #dfe7f0 0 9px, #bfcde0 10px, #dfe7f0 11px 18px); }
	.rose { background: radial-gradient(at 80% 10%, #eec8dc, transparent 80%), linear-gradient(40deg, #c3d9f3, #eee5ed); }
	.mist { background: #dfe6ef; }
	.sample-card { width: 50%; height: 40px; background: #ffffffc9; border: 1px solid #fff; border-radius: 8px; }
	.sample-card.two { height: 50px; }
	.selected { position: absolute; top: 8px; right: 8px; width: 22px; height: 22px; border-radius: 50%; background: #fcfdff; display: grid; place-items: center; color: #6675b8; }
	.option strong { display: block; margin: 9px 0 2px; font-size: 15px; font-weight: 600; }
	.description { display: block; color: var(--ink-muted); font-size: 12px; line-height: 1.4; }
	.motion-row { display: flex; align-items: center; gap: 14px; justify-content: space-between; border-top: var(--hairline); padding-top: 24px; margin-top: 28px; }
	.motion-row strong { font-size: 15px; font-weight: 600; }
	.motion-row p { font-size: 12px; margin-top: 4px; color: var(--ink-muted); }
	.switch { flex: none; width: 44px; height: 26px; border-radius: 20px; padding: 3px; background: #d9dfe8; transition: background 250ms; }
	.switch[aria-checked='true'] { background: linear-gradient(100deg, #79b9ee, #b7a1e5); }
	.switch span { display: block; height: 20px; width: 20px; border-radius: 50%; background: white; box-shadow: 0 2px 4px #40517120; transition: transform 350ms var(--ease-expo); }
	.switch[aria-checked='true'] span { transform: translateX(18px); }
	.saved { display: flex; align-items: center; gap: 6px; margin-top: 32px; font-size: 12px; color: var(--ink-muted); }
</style>
