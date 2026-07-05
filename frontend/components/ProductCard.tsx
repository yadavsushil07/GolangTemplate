"use client";

import Link from "next/link";
import { useState } from "react";
import type { Product } from "@/lib/types";
import { formatPrice } from "@/lib/format";
import { useCart } from "@/lib/cart-context";
import ProductArt from "./ProductArt";

export default function ProductCard({ product }: { product: Product }) {
  const { add } = useCart();
  const [busy, setBusy] = useState(false);
  const [added, setAdded] = useState(false);

  async function quickAdd(e: React.MouseEvent) {
    e.preventDefault();
    if (busy) return;
    setBusy(true);
    try {
      await add(product.id, 1);
      setAdded(true);
      setTimeout(() => setAdded(false), 1500);
    } catch {
      /* ignore */
    } finally {
      setBusy(false);
    }
  }

  const collection = product.categories?.[0]?.name;

  return (
    <Link href={`/product/${product.slug}`} className="group block">
      <div className="card-media mb-4 group-hover:shadow-luxe transition-shadow">
        <ProductArt
          seed={product.id}
          imageUrl={product.image_url}
          label={product.name}
          className="group-hover:scale-105 transition-transform duration-700"
        />
        {product.stock <= 0 && (
          <span className="absolute top-3.5 left-3.5 bg-plum text-white text-[9px] tracking-[0.14em] uppercase px-3 py-1.5">
            Sold Out
          </span>
        )}
        <button
          onClick={quickAdd}
          disabled={busy || product.stock <= 0}
          className="absolute inset-x-0 bottom-0 bg-ink/90 text-white text-[10px] tracking-[0.22em] uppercase py-3.5 opacity-0 translate-y-2 group-hover:opacity-100 group-hover:translate-y-0 transition-all disabled:opacity-60"
        >
          {added ? "Added ✓" : busy ? "Adding…" : "Quick Add +"}
        </button>
      </div>
      {collection && (
        <div className="text-[10px] tracking-[0.22em] uppercase text-champ">
          {collection}
        </div>
      )}
      <h3 className="font-display text-lg leading-snug my-1">{product.name}</h3>
      <div className="text-sm text-muted">{formatPrice(product.price_cents)}</div>
    </Link>
  );
}
