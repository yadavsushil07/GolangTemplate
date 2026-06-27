import { useState, useCallback } from 'react';
import { getCart, addToCart, removeFromCart, getSessionID } from '@/services/api';

interface CartItem {
  id: number;
  product_id: number;
  quantity: number;
  product: {
    id: number;
    name: string;
    price_cents: number;
    image_url: string;
  };
}

interface CartSummary {
  items: CartItem[];
  total_cents: number;
}

export function useCart() {
  const [cart, setCart] = useState<CartSummary>({ items: [], total_cents: 0 });
  const [loading, setLoading] = useState(false);

  const fetchCart = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getCart();
      setCart(res.data || { items: [], total_cents: 0 });
    } finally {
      setLoading(false);
    }
  }, []);

  const add = useCallback(async (productId: number, qty = 1) => {
    const res = await addToCart(productId, qty);
    setCart(res.data);
  }, []);

  const remove = useCallback(async (productId: number) => {
    const res = await removeFromCart(productId);
    setCart(res.data);
  }, []);

  const sessionID = useCallback(() => getSessionID(), []);

  return { cart, loading, fetchCart, add, remove, sessionID };
}
