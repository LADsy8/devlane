/**
 * Extracts a human-readable message from an axios-style API error. The API
 * returns `{ "error": "..." }` on failure, so surface that when present and
 * fall back to a caller-supplied message otherwise (e.g. network errors).
 */
export function apiErrorMessage(err: unknown, fallback: string): string {
  const data = (err as { response?: { data?: { error?: unknown } } })?.response?.data;
  if (data && typeof data.error === 'string' && data.error.trim()) {
    return data.error;
  }
  return fallback;
}
