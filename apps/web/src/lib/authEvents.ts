export const AUTH_UNAUTHORIZED = 'devlane:auth-unauthorized';

/** Fired when an API request returns 401 so the app can drop the expired session. */
export function dispatchAuthUnauthorized() {
  if (typeof window === 'undefined') return;
  window.dispatchEvent(new CustomEvent(AUTH_UNAUTHORIZED));
}
