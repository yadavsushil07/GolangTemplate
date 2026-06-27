import { useState, useEffect, useCallback } from 'react';
import { clearToken, getToken, setToken, verifyOTP, requestOTP } from '@/services/api';

interface User {
  id: number;
  identifier: string;
  role: string;
}

export function useAuth() {
  const [user, setUser] = useState<User | null>(null);
  const [token, setTokenState] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    (async () => {
      const t = await getToken();
      if (t) {
        setTokenState(t);
        const payload = parseJWT(t);
        if (payload) setUser({ id: payload.id, identifier: payload.sub, role: payload.role });
      }
      setLoading(false);
    })();
  }, []);

  const sendOTP = useCallback(async (identifier: string) => {
    const res = await requestOTP(identifier);
    return res.data;
  }, []);

  const login = useCallback(async (identifier: string, code: string) => {
    const res = await verifyOTP(identifier, code);
    const { token: t, user: u } = res.data;
    await setToken(t);
    setTokenState(t);
    setUser(u);
    return u;
  }, []);

  const logout = useCallback(async () => {
    await clearToken();
    setTokenState(null);
    setUser(null);
  }, []);

  const isVendor = user?.role === 'vendor';

  return { user, token, loading, isVendor, sendOTP, login, logout };
}

function parseJWT(token: string): Record<string, any> | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    const payload = atob(parts[1].replace(/-/g, '+').replace(/_/g, '/'));
    return JSON.parse(payload);
  } catch {
    return null;
  }
}
