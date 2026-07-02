"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
  ReactNode,
} from "react";
import type { CartSummary } from "./types";
import * as api from "./api";

interface CartCtx {
  cart: CartSummary | null;
  count: number;
  loading: boolean;
  refresh: () => Promise<void>;
  add: (productId: number, qty?: number, variantId?: number | null) => Promise<void>;
  remove: (productId: number) => Promise<void>;
}

const Ctx = createContext<CartCtx | null>(null);

export function CartProvider({ children }: { children: ReactNode }) {
  const [cart, setCart] = useState<CartSummary | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    try {
      setCart(await api.getCart());
    } catch {
      setCart({ items: [], total_cents: 0 });
    } finally {
      setLoading(false);
    }
  }, []);

  const add = useCallback(
    async (productId: number, qty = 1, variantId?: number | null) => {
      setCart(await api.addToCart(productId, qty, variantId));
    },
    []
  );

  const remove = useCallback(async (productId: number) => {
    setCart(await api.removeFromCart(productId));
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const count = cart?.items?.reduce((n, i) => n + i.quantity, 0) ?? 0;

  return (
    <Ctx.Provider value={{ cart, count, loading, refresh, add, remove }}>
      {children}
    </Ctx.Provider>
  );
}

export function useCart() {
  const c = useContext(Ctx);
  if (!c) throw new Error("useCart must be used within CartProvider");
  return c;
}
