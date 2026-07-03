"use client";

import { useEffect, useRef, useState } from "react";

export interface Country {
  name: string;
  code: string;
  dial: string;
  flag: string;
}

const COUNTRIES: Country[] = [
  { name: "India", code: "IN", dial: "+91", flag: "🇮🇳" },
  { name: "UAE", code: "AE", dial: "+971", flag: "🇦🇪" },
  { name: "United States", code: "US", dial: "+1", flag: "🇺🇸" },
  { name: "United Kingdom", code: "GB", dial: "+44", flag: "🇬🇧" },
  { name: "Canada", code: "CA", dial: "+1", flag: "🇨🇦" },
  { name: "Australia", code: "AU", dial: "+61", flag: "🇦🇺" },
  { name: "Singapore", code: "SG", dial: "+65", flag: "🇸🇬" },
  { name: "Malaysia", code: "MY", dial: "+60", flag: "🇲🇾" },
  { name: "New Zealand", code: "NZ", dial: "+64", flag: "🇳🇿" },
  { name: "Saudi Arabia", code: "SA", dial: "+966", flag: "🇸🇦" },
  { name: "Qatar", code: "QA", dial: "+974", flag: "🇶🇦" },
  { name: "Kuwait", code: "KW", dial: "+965", flag: "🇰🇼" },
  { name: "Oman", code: "OM", dial: "+968", flag: "🇴🇲" },
  { name: "Bahrain", code: "BH", dial: "+973", flag: "🇧🇭" },
  { name: "Nepal", code: "NP", dial: "+977", flag: "🇳🇵" },
  { name: "Sri Lanka", code: "LK", dial: "+94", flag: "🇱🇰" },
  { name: "Bangladesh", code: "BD", dial: "+880", flag: "🇧🇩" },
  { name: "Pakistan", code: "PK", dial: "+92", flag: "🇵🇰" },
  { name: "South Africa", code: "ZA", dial: "+27", flag: "🇿🇦" },
  { name: "Germany", code: "DE", dial: "+49", flag: "🇩🇪" },
];

// IANA timezone → country code
const TZ_MAP: Record<string, string> = {
  "Asia/Kolkata": "IN",
  "Asia/Calcutta": "IN",
  "Asia/Dubai": "AE",
  "Asia/Muscat": "OM",
  "Asia/Qatar": "QA",
  "Asia/Kuwait": "KW",
  "Asia/Bahrain": "BH",
  "Asia/Riyadh": "SA",
  "Asia/Singapore": "SG",
  "Asia/Kuala_Lumpur": "MY",
  "Asia/Colombo": "LK",
  "Asia/Kathmandu": "NP",
  "Asia/Dhaka": "BD",
  "Asia/Karachi": "PK",
  "America/New_York": "US",
  "America/Chicago": "US",
  "America/Denver": "US",
  "America/Los_Angeles": "US",
  "America/Toronto": "CA",
  "America/Vancouver": "CA",
  "Europe/London": "GB",
  "Australia/Sydney": "AU",
  "Australia/Melbourne": "AU",
  "Pacific/Auckland": "NZ",
  "Africa/Johannesburg": "ZA",
  "Europe/Berlin": "DE",
};

function detectCountry(): Country {
  try {
    const tz = Intl.DateTimeFormat().resolvedOptions().timeZone;
    const code = TZ_MAP[tz];
    if (code) {
      const match = COUNTRIES.find((c) => c.code === code);
      if (match) return match;
    }
  } catch {
    // fall through to default
  }
  return COUNTRIES[0]; // India
}

interface Props {
  selected: Country;
  onChange: (c: Country) => void;
}

export default function CountryCodePicker({ selected, onChange }: Props) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const modalRef = useRef<HTMLDivElement>(null);
  const searchRef = useRef<HTMLInputElement>(null);

  const filtered = COUNTRIES.filter(
    (c) =>
      c.name.toLowerCase().includes(search.toLowerCase()) ||
      c.dial.includes(search)
  );

  // Focus search input when modal opens
  useEffect(() => {
    if (open) setTimeout(() => searchRef.current?.focus(), 50);
  }, [open]);

  // Close on outside click
  useEffect(() => {
    function handler(e: MouseEvent) {
      if (modalRef.current && !modalRef.current.contains(e.target as Node)) {
        setOpen(false);
        setSearch("");
      }
    }
    if (open) document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  return (
    <div className="relative" ref={modalRef}>
      <button
        type="button"
        onClick={() => { setOpen((o) => !o); setSearch(""); }}
        className="flex items-center gap-1.5 h-full px-3 border-r border-line bg-cream hover:bg-oat transition-colors text-sm focus:outline-none"
        aria-label="Select country code"
        aria-expanded={open}
      >
        <span className="text-lg leading-none">{selected.flag}</span>
        <span className="text-[12px] tracking-wide text-ink">{selected.dial}</span>
        <svg className="w-3 h-3 text-muted" viewBox="0 0 10 6" fill="none">
          <path d="M1 1l4 4 4-4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
        </svg>
      </button>

      {open && (
        <div className="absolute left-0 top-full mt-1 w-64 bg-white border border-line shadow-xl z-50 rounded-sm overflow-hidden">
          {/* Search */}
          <div className="p-2 border-b border-line">
            <input
              ref={searchRef}
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search country…"
              className="w-full text-xs px-3 py-2 border border-line rounded-sm focus:outline-none focus:border-clay"
            />
          </div>
          {/* List */}
          <ul className="max-h-52 overflow-y-auto">
            {filtered.length === 0 ? (
              <li className="px-4 py-3 text-xs text-muted">No results</li>
            ) : (
              filtered.map((c) => (
                <li key={c.code + c.dial}>
                  <button
                    type="button"
                    onClick={() => { onChange(c); setOpen(false); setSearch(""); }}
                    className={`w-full text-left flex items-center gap-3 px-4 py-2.5 text-xs hover:bg-cream transition-colors ${
                      selected.code === c.code && selected.dial === c.dial
                        ? "bg-oat font-medium text-clay"
                        : "text-ink"
                    }`}
                  >
                    <span className="text-base leading-none">{c.flag}</span>
                    <span className="flex-1">{c.name}</span>
                    <span className="text-muted">{c.dial}</span>
                  </button>
                </li>
              ))
            )}
          </ul>
        </div>
      )}
    </div>
  );
}

export { detectCountry, COUNTRIES };
