/* eslint-disable react-refresh/only-export-components -- context file exports AuthProvider + useAuth; keep for future use */
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';
import type { User } from '../types';
import type { UserApiResponse } from '../api/types';
import { apiClient, clearApiBearerAuthHeader } from '../api/client';
import { authService } from '../services/authService';

function mapApiUserToUser(api: UserApiResponse): User {
  const name =
    api.display_name?.trim() ||
    [api.first_name, api.last_name].filter(Boolean).join(' ').trim() ||
    api.username;
  return {
    id: api.id,
    email: api.email ?? '',
    name,
    avatarUrl: api.avatar ?? null,
    coverImageUrl: api.cover_image ?? null,
    isInstanceAdmin: api.is_instance_admin ?? false,
  };
}

interface AuthContextValue {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<boolean>;
  logout: () => Promise<void>;
  setUserFromApi: (api: UserApiResponse) => void;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // After OAuth, the API may pass the session token in the URL fragment
    // (cross-origin dev mode). Read it, set as Bearer header, then clear.
    const hash = window.location.hash;
    if (hash.includes('session_token=')) {
      const params = new URLSearchParams(hash.slice(1));
      const token = params.get('session_token');
      if (token) {
        apiClient.defaults.headers.common['Authorization'] = `Bearer ${token}`;
        window.history.replaceState(null, '', window.location.pathname + window.location.search);
      }
    }

    let cancelled = false;
    authService.getMe().then((api) => {
      if (!cancelled && api) setUser(mapApiUserToUser(api));
      if (!cancelled) setIsLoading(false);
    });
    return () => {
      cancelled = true;
    };
  }, []);

  const setUserFromApi = useCallback((api: UserApiResponse) => {
    setUser(mapApiUserToUser(api));
  }, []);

  // Re-fetches the canonical user profile from /api/users/me/. Sign-in/
  // sign-up/magic-code responses don't carry every field (e.g.
  // is_instance_admin is only computed on the Me endpoint), so callers that
  // need the full, authoritative profile right after establishing a session
  // should use this instead of setUserFromApi with the raw auth response.
  const refreshUser = useCallback(async () => {
    const api = await authService.getMe();
    if (api) setUser(mapApiUserToUser(api));
  }, []);

  const login = useCallback(
    async (email: string, password: string): Promise<boolean> => {
      await authService.signIn({ email, password });
      await refreshUser();
      return true;
    },
    [refreshUser],
  );

  const logout = useCallback(async () => {
    try {
      await authService.signOut();
    } finally {
      clearApiBearerAuthHeader();
      setUser(null);
    }
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      isAuthenticated: user != null,
      isLoading,
      login,
      logout,
      setUserFromApi,
      refreshUser,
    }),
    [user, isLoading, login, logout, setUserFromApi, refreshUser],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return ctx;
}
