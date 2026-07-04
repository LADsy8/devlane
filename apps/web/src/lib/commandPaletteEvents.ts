export const OPEN_COMMAND_PALETTE = 'devlane:open-command-palette';

export function dispatchOpenCommandPalette() {
  if (typeof window === 'undefined') return;
  window.dispatchEvent(new CustomEvent(OPEN_COMMAND_PALETTE));
}
