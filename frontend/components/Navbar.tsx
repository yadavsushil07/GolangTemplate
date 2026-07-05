"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useCart } from "@/lib/cart-context";
import { getToken, isTokenExpired, logout, getRole, listCategories } from "@/lib/api";
import type { Category } from "@/lib/types";

// Decode the identifier (phone/email) from the JWT payload without a library
function getIdentifierFromToken(): string | null {
  if (typeof window === "undefined") return null;
  const token =
    localStorage.getItem("sby_token") ?? sessionStorage.getItem("sby_token");
  if (!token) return null;
  try {
    const payload = JSON.parse(atob(token.split(".")[1]));
    // JWT claims produced by the Go backend use "identifier" or fall back to "sub"
    return (payload.identifier as string) || (payload.sub as string) || null;
  } catch {
    return null;
  }
}

// Static fallback nav items (shown while categories are loading / if backend is down)
const STATIC_NAV = [
  { label: "Shop All", slug: null },
  { label: "Farshi Sets", slug: "farshi-sets" },
  { label: "Co-ord Sets", slug: "co-ord-sets" },
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
  const [identifier, setIdentifier] = useState<string | null>(null);
  const [navItems, setNavItems] = useState(STATIC_NAV);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const valid = !!(getToken() && !isTokenExpired());
    setLoggedIn(valid);
    if (valid) setIdentifier(getIdentifierFromToken());
  }, []);

  // Fetch real categories and resolve dynamic slugs for Farshi Sets / Co-ord Sets
  useEffect(() => {
    listCategories().then((cats: Category[]) => {
      function resolveSlug(label: string, fallback: string): string {
        const match = cats.find((c) =>
          c.name.toLowerCase().includes(label.toLowerCase())
        );
        return match ? match.slug : fallback;
      }
      setNavItems([
        { label: "Shop All", slug: null },
        { label: "Farshi Sets", slug: resolveSlug("farshi", "farshi-sets") },
        { label: "Co-ord Sets", slug: resolveSlug("co-ord", "co-ord-sets") },
      ]);
    }).catch(() => {}); // keep static fallback on network failure
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
    setIdentifier(null);
    setDropdownOpen(false);
    setDrawerOpen(false);
    router.push("/");
  }

  const navHref = (slug: string | null) =>
    slug ? `/shop?category=${slug}` : "/shop";

  return (
    <>
      <div className="bg-plum text-[#d8c6a4] text-center text-[11px] tracking-[0.22em] uppercase py-2.5 px-5">
        Free shipping worldwide on orders above ₹5,000 · Made to order
      </div>

      <nav className="sticky top-0 z-50 bg-cream/90 backdrop-blur border-b border-line">
        <div className="grid grid-cols-[auto_1fr_auto] md:grid-cols-3 items-center h-20 px-6 md:px-12">
          {/* Left / hamburger */}
          <div className="flex items-center gap-7">
            <button
              className="md:hidden text-2xl"
              onClick={() => setDrawerOpen((o) => !o)}
              aria-label="Menu"
            >
              ☰
            </button>
            <div className="hidden md:flex items-center gap-7">
              {navItems.map((n) => (
                <Link
                  key={n.label}
                  href={navHref(n.slug)}
                  className="text-[11px] tracking-[0.22em] uppercase hover:text-clay transition-colors"
                >
                  {n.label}
                </Link>
              ))}
            </div>
          </div>

          {/* Logo */}
          <Link
            href="/"
            className="font-display text-xl md:text-2xl tracking-[0.25em] uppercase text-center"
          >
            SBY <span className="text-champ italic">TWILIGHT</span>
          </Link>

          {/* Right — Bag + Profile */}
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
              <button
                onClick={() => setDropdownOpen((o) => !o)}
                className={`w-8 h-8 rounded-full flex items-center justify-center transition-colors focus:outline-none border ${
                  loggedIn
                    ? "bg-plum text-[#d8c6a4] border-plum hover:bg-clay hover:border-clay"
                    : "border-current hover:text-clay hover:border-clay"
                }`}
                aria-label="Profile menu"
                aria-expanded={dropdownOpen}
              >
                <UserIcon />
              </button>

              {dropdownOpen && (
                <div className="absolute right-0 top-full mt-2 w-56 bg-white border border-line shadow-xl z-50">
                  {loggedIn ? (
                    // ── Logged-in dropdown ────────────────────────────────
                    <>
                      <div className="px-5 pt-5 pb-4 border-b border-line">
                        <p className="font-semibold text-sm leading-snug">
                          Hello {identifier ?? "there"}
                        </p>
                        {identifier && (
                          <p className="text-xs text-muted mt-0.5 truncate">{identifier}</p>
                        )}
                      </div>
                      <div className="py-1">
                        <Link
                          href="/account/orders"
                          onClick={() => setDropdownOpen(false)}
                          className="block px-5 py-2.5 text-sm hover:bg-cream hover:text-clay transition-colors"
                        >
                          Orders
                        </Link>
                        <Link
                          href="/account/details"
                          onClick={() => setDropdownOpen(false)}
                          className="block px-5 py-2.5 text-sm hover:bg-cream hover:text-clay transition-colors"
                        >
                          Account Details
                        </Link>
                      </div>
                      <div className="border-t border-line py-1">
                        <button
                          onClick={handleLogout}
                          className="w-full text-left px-5 py-2.5 text-sm hover:bg-cream hover:text-clay transition-colors"
                        >
                          Log out
                        </button>
                      </div>
                    </>
                  ) : (
                    // ── Logged-out dropdown ───────────────────────────────
                    <>
                      <div className="px-5 pt-5 pb-4 border-b border-line">
                        <p className="font-semibold text-base">Welcome</p>
                        <p className="text-xs text-muted mt-1 leading-snug">
                          To access account and manage orders
                        </p>
                        <Link
                          href="/account"
                          onClick={() => setDropdownOpen(false)}
                          className="mt-4 block border border-clay text-clay text-[11px] tracking-[0.18em] uppercase text-center py-2.5 hover:bg-clay hover:text-white transition-colors"
                        >
                          Login / Signup
                        </Link>
                      </div>
                      <div className="py-1">
                        <Link
                          href="/account/orders"
                          onClick={() => setDropdownOpen(false)}
                          className="block px-5 py-2.5 text-sm hover:bg-cream hover:text-clay transition-colors"
                        >
                          Orders
                        </Link>
                        <a
                          href="mailto:sby.twilight4@gmail.com"
                          onClick={() => setDropdownOpen(false)}
                          className="block px-5 py-2.5 text-sm hover:bg-cream hover:text-clay transition-colors"
                        >
                          Contact Us
                        </a>
                      </div>
                    </>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Mobile drawer */}
        {drawerOpen && (
          <div className="md:hidden border-t border-line px-6 py-4 flex flex-col gap-4">
            {navItems.map((n) => (
              <Link
                key={n.label}
                href={navHref(n.slug)}
                onClick={() => setDrawerOpen(false)}
                className="text-[12px] tracking-[0.18em] uppercase"
              >
                {n.label}
              </Link>
            ))}
            <div className="border-t border-line pt-3 flex flex-col gap-3">
              {loggedIn ? (
                <>
                  <Link
                    href="/account/orders"
                    onClick={() => setDrawerOpen(false)}
                    className="text-[12px] tracking-[0.18em] uppercase"
                  >
                    Orders
                  </Link>
                  <Link
                    href="/account/details"
                    onClick={() => setDrawerOpen(false)}
                    className="text-[12px] tracking-[0.18em] uppercase"
                  >
                    Account Details
                  </Link>
                  <button
                    onClick={handleLogout}
                    className="text-left text-[12px] tracking-[0.18em] uppercase text-red-600"
                  >
                    Log out
                  </button>
                </>
              ) : (
                <>
                  <Link
                    href="/account"
                    onClick={() => setDrawerOpen(false)}
                    className="text-[12px] tracking-[0.18em] uppercase"
                  >
                    Login / Signup
                  </Link>
                  <a
                    href="mailto:sby.twilight4@gmail.com"
                    className="text-[12px] tracking-[0.18em] uppercase"
                  >
                    Contact Us
                  </a>
                </>
              )}
            </div>
          </div>
        )}
      </nav>
    </>
  );
}
