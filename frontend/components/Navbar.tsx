"use client";

import Link from "next/link";
import { useState } from "react";
import { useCart } from "@/lib/cart-context";

const NAV = [
  { label: "Shop All", href: "/shop" },
  { label: "Suit Sets", href: "/shop?category=suit-sets" },
  { label: "Anarkali", href: "/shop?category=anarkali" },
  { label: "Co-ord Sets", href: "/shop?category=co-ord-sets" },
];

export default function Navbar() {
  const { count } = useCart();
  const [open, setOpen] = useState(false);

  return (
    <>
      <div className="bg-plum text-[#d8c6a4] text-center text-[11px] tracking-[0.22em] uppercase py-2.5 px-5">
        Free shipping worldwide on orders above ₹5,000 · Made to order
      </div>

      <nav className="sticky top-0 z-50 bg-cream/90 backdrop-blur border-b border-line">
        <div className="grid grid-cols-[auto_1fr_auto] md:grid-cols-3 items-center h-20 px-6 md:px-12">
          {/* left / hamburger */}
          <div className="flex items-center gap-7">
            <button
              className="md:hidden text-2xl"
              onClick={() => setOpen((o) => !o)}
              aria-label="Menu"
            >
              ☰
            </button>
            <div className="hidden md:flex items-center gap-7">
              {NAV.map((n) => (
                <Link
                  key={n.label}
                  href={n.href}
                  className="text-[11px] tracking-[0.22em] uppercase hover:text-clay transition-colors"
                >
                  {n.label}
                </Link>
              ))}
            </div>
          </div>

          {/* logo */}
          <Link
            href="/"
            className="font-display text-2xl md:text-[27px] tracking-[0.3em] uppercase text-center"
          >
            AARY<span className="text-champ italic">A</span>
          </Link>

          {/* right */}
          <div className="flex items-center justify-end gap-5 text-[11px] tracking-[0.22em] uppercase">
            <Link href="/account" className="hidden sm:inline hover:text-clay">
              Account
            </Link>
            <Link href="/cart" className="hover:text-clay flex items-center gap-1.5">
              Bag
              <span className="bg-clay text-white rounded-full min-w-[18px] h-[18px] inline-flex items-center justify-center text-[10px] px-1">
                {count}
              </span>
            </Link>
          </div>
        </div>

        {/* mobile drawer */}
        {open && (
          <div className="md:hidden border-t border-line px-6 py-4 flex flex-col gap-4">
            {NAV.map((n) => (
              <Link
                key={n.label}
                href={n.href}
                onClick={() => setOpen(false)}
                className="text-[12px] tracking-[0.18em] uppercase"
              >
                {n.label}
              </Link>
            ))}
            <Link
              href="/account"
              onClick={() => setOpen(false)}
              className="text-[12px] tracking-[0.18em] uppercase"
            >
              Account
            </Link>
          </div>
        )}
      </nav>
    </>
  );
}
