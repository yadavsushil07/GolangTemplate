"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useCart } from "@/lib/cart-context";
import { getToken, isTokenExpired, logout } from "@/lib/api";

const NAV = [
  { label: "Shop All", href: "/shop" },
  { label: "Suit Sets", href: "/shop?category=suit-sets" },
  { label: "Anarkali", href: "/shop?category=anarkali" },
  { label: "Co-ord Sets", href: "/shop?category=co-ord-sets" },
];

function UserIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.5}
      strokeLinecap="round"
      strokeLinejoin="round"
      className="w-5 h-5"
      aria-hidden="true"
    >
      <circle cx="12" cy="8" r="4" />
      <path d="M4 20c0-4 3.6-7 8-7s8 3 8 7" />
    </svg>
  );
}

export default function Navbar() {
  const { count } = useCart();
  const router = useRouter();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [loggedIn, setLoggedIn] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setLoggedIn(!!(getToken() && !isTokenExpired()));
  }, []);

  // Close dropdown on outside click
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  function handleLogout() {
    logout();
    setLoggedIn(false);
    setDropdownOpen(false);
    setDrawerOpen(false);
    router.push("/account");
  }

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
              onClick={() => setDrawerOpen((o) => !o)}
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
            className="font-display text-xl md:text-2xl tracking-[0.25em] uppercase text-center"
          >
            SBY <span className="text-champ italic">TWILIGHT</span>
          </Link>

          {/* right — Bag + Profile */}
          <div className="flex items-center justify-end gap-4 text-[11px] tracking-[0.22em] uppercase">
            {/* Bag */}
            <Link href="/cart" className="hover:text-clay flex items-center gap-1.5">
              Bag
              <span className="bg-clay text-white rounded-full min-w-[18px] h-[18px] inline-flex items-center justify-center text-[10px] px-1">
                {count}
              </span>
            </Link>

            {/* Profile */}
            <div className="relative" ref={dropdownRef}>
              {loggedIn ? (
                <>
                  <button
                    onClick={() => setDropdownOpen((o) => !o)}
                    className="w-8 h-8 rounded-full bg-plum text-[#d8c6a4] flex items-center justify-center hover:bg-clay transition-colors focus:outline-none"
                    aria-label="Profile menu"
                    aria-expanded={dropdownOpen}
                  >
                    <UserIcon />
                  </button>

                  {dropdownOpen && (
                    <div className="absolute right-0 top-full mt-2 w-44 bg-white border border-line shadow-lg z-50 py-1">
                      <Link
                        href="/account"
                        onClick={() => setDropdownOpen(false)}
                        className="block px-4 py-2.5 text-[11px] tracking-[0.18em] uppercase hover:bg-cream hover:text-clay transition-colors"
                      >
                        Account
                      </Link>
                      <Link
                        href="/account/orders"
                        onClick={() => setDropdownOpen(false)}
                        className="block px-4 py-2.5 text-[11px] tracking-[0.18em] uppercase hover:bg-cream hover:text-clay transition-colors"
                      >
                        Orders
                      </Link>
                      <div className="border-t border-line my-1" />
                      <button
                        onClick={handleLogout}
                        className="w-full text-left px-4 py-2.5 text-[11px] tracking-[0.18em] uppercase hover:bg-cream hover:text-clay transition-colors"
                      >
                        Log out
                      </button>
                    </div>
                  )}
                </>
              ) : (
                <Link
                  href="/account"
                  className="w-8 h-8 rounded-full border border-current flex items-center justify-center hover:text-clay hover:border-clay transition-colors"
                  aria-label="Sign in"
                >
                  <UserIcon />
                </Link>
              )}
            </div>
          </div>
        </div>

        {/* mobile drawer */}
        {drawerOpen && (
          <div className="md:hidden border-t border-line px-6 py-4 flex flex-col gap-4">
            {NAV.map((n) => (
              <Link
                key={n.label}
                href={n.href}
                onClick={() => setDrawerOpen(false)}
                className="text-[12px] tracking-[0.18em] uppercase"
              >
                {n.label}
              </Link>
            ))}
            <div className="border-t border-line pt-3 flex flex-col gap-3">
              <Link
                href="/account"
                onClick={() => setDrawerOpen(false)}
                className="text-[12px] tracking-[0.18em] uppercase"
              >
                Account
              </Link>
              {loggedIn && (
                <>
                  <Link
                    href="/account/orders"
                    onClick={() => setDrawerOpen(false)}
                    className="text-[12px] tracking-[0.18em] uppercase"
                  >
                    Orders
                  </Link>
                  <button
                    onClick={handleLogout}
                    className="text-left text-[12px] tracking-[0.18em] uppercase text-red-600"
                  >
                    Log out
                  </button>
                </>
              )}
            </div>
          </div>
        )}
      </nav>
    </>
  );
}
