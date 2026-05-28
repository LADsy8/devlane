import axios, { type AxiosError } from 'axios';

/**
 * Prefer env-driven API base (VITE_API_BASE_URL).
 * In local dev, fallback remains http://localhost:8080.
 * In production, empty string keeps requests relative (same-origin).
 */
export const API_BASE =
  import.meta.env.VITE_API_BASE_URL ?? (import.meta.env.DEV ? 'http://localhost:8080' : '');

export const apiClient = axios.create({
  baseURL: API_BASE,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

/** Clears Bearer token set from OAuth URL fragment (dev / cross-origin); cookie sessions unaffected. */
export function clearApiBearerAuthHeader(): void {
  delete apiClient.defaults.headers.common['Authorization'];
}

// When sending FormData (e.g. file upload), omit Content-Type so the browser sets
// multipart/form-data with the correct boundary. Otherwise the server gets
// Content-Type: application/json and cannot parse the multipart form → 400.
apiClient.interceptors.request.use((config) => {
  if (config.data instanceof FormData && config.headers) {
    const h = config.headers as Record<string, unknown>;
    delete h['Content-Type'];
  }
  return config;
});

/**
 * Shape of API error response body.
 * The auth subsystem returns { error_code, error_message };
 * DRF views may return { error }, { detail }, or { message }.
 */
export interface ApiErrorResponse {
  error?: string;
  detail?: string;
  message?: string;
  /** Auth-subsystem error fields (AUTHENTICATION_ERROR_CODES) */
  error_message?: string;
  error_code?: number;
}

/** Human-readable labels for auth error codes from the backend */
const AUTH_ERROR_LABELS: Record<number, string> = {
  5000: 'Instance is not configured yet.',
  5005: 'Invalid email address.',
  5010: 'Email is required.',
  5015: 'Sign-up is disabled for this instance.',
  5016: 'Magic-link login is disabled.',
  5018: 'Password login is disabled.',
  5019: 'Your account has been deactivated.',
  5020: 'Invalid password.',
  5021: 'Password is too weak.',
  5025: 'Email (SMTP) is not configured — contact your admin.',
  5030: 'An account with this email already exists.',
  5035: 'Sign-up failed. Please try again.',
  5060: 'No account found with this email.',
  5065: 'Sign-in failed. Check your credentials.',
  5090: 'Invalid magic code.',
  5092: 'Invalid magic code.',
  5095: 'Magic code has expired. Please request a new one.',
  5097: 'Magic code has expired. Please request a new one.',
  5100: 'Too many attempts. Request a new code.',
  5102: 'Too many attempts. Request a new code.',
  5104: 'OAuth is not configured.',
  5125: 'This reset link is invalid.',
  5130: 'This reset link has expired. Please request a new one.',
  5135: 'Current password is incorrect.',
  5138: 'Password is required.',
  5140: 'New password is invalid.',
  5145: 'A password is already set on this account.',
  5900: 'Too many requests. Please wait a moment and try again.',
  5999: 'Authentication failed. Please try again.',
};

/**
 * Extract a user-facing error message from an Axios error.
 * Handles:
 *   - Auth subsystem: { error_code, error_message }
 *   - DRF: { error }, { detail }, { message }
 *   - Generic HTTP status fallback
 */
export function getApiErrorMessage(err: unknown): string {
  if (axios.isAxiosError(err)) {
    const ax = err as AxiosError<ApiErrorResponse>;
    const body = ax.response?.data;
    // Auth-subsystem errors: prefer human label, fall back to raw error_message
    if (body?.error_code && AUTH_ERROR_LABELS[body.error_code]) {
      return AUTH_ERROR_LABELS[body.error_code];
    }
    if (body?.error_message) return body.error_message;
    // Standard DRF / custom error fields
    if (body?.error) return body.error;
    if (body?.detail) return body.detail;
    if (body?.message) return body.message;
    if (ax.response?.status) return `Request failed (${ax.response.status}).`;
  }
  if (err instanceof Error) return err.message;
  return 'An unexpected error occurred.';
}

apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiErrorResponse>) => {
    const message = getApiErrorMessage(error);
    return Promise.reject(new Error(message));
  },
);
