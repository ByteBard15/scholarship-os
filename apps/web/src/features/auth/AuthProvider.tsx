import {
  createContext,
  PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { loginRequest, logoutRequest, meRequest } from "./api";
import { authStorage, unauthorizedEvent } from "./storage";
import type { AuthUser } from "./types";
import { useQueryClient } from "@tanstack/react-query";

interface AuthContextValue {
  user: AuthUser | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login(email: string, password: string): Promise<void>;
  logout(): Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: PropsWithChildren) {
  const queryClient = useQueryClient();
  const [token, setToken] = useState<string | null>(() => authStorage.get());
  const [user, setUser] = useState<AuthUser | null>(null);
  const [isLoading, setLoading] = useState(Boolean(token));

  const clear = useCallback(() => {
    authStorage.clear();
    setToken(null);
    setUser(null);
    setLoading(false);
    queryClient.clear();
  }, [queryClient]);

  useEffect(() => {
    const unauthorized = () => clear();
    window.addEventListener(unauthorizedEvent, unauthorized);
    return () => window.removeEventListener(unauthorizedEvent, unauthorized);
  }, [clear]);

  useEffect(() => {
    if (!token) {
      setLoading(false);
      return;
    }
    setLoading(true);
    void meRequest()
      .then((response) => setUser(response.data))
      .catch(() => clear())
      .finally(() => setLoading(false));
  }, [token, clear]);

  const login = useCallback(async (email: string, password: string) => {
    const response = await loginRequest(email, password);
    authStorage.set(response.data.accessToken);
    setToken(response.data.accessToken);
    setUser(response.data.user);
  }, []);

  const logout = useCallback(async () => {
    try {
      await logoutRequest();
    } finally {
      clear();
    }
  }, [clear]);

  const value = useMemo(
    () => ({
      user,
      token,
      isAuthenticated: Boolean(token && user),
      isLoading,
      login,
      logout,
    }),
    [user, token, isLoading, login, logout],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) throw new Error("useAuth must be used inside AuthProvider");
  return value;
}
