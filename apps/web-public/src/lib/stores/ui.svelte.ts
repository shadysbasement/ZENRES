/** Shared UI state for the public site (Svelte 5 runes). */
class UiState {
	/** True once the preloader has lifted — hero animations key off this. */
	ready = $state(false);
	menuOpen = $state(false);
}
export const ui = new UiState();
