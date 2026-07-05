"use client";

import { useEffect, useState } from "react";
import { requestOTP, verifyOTP } from "@/lib/api";
import CountryCodePicker, { detectCountry, type Country } from "./CountryCodePicker";

interface Props {
  onSuccess: () => void;
  title?: string;
  subtitle?: string;
}

function isPhoneInput(val: string) {
  return val.length > 0 && /^[0-9+]/.test(val) && !val.includes("@");
}

function buildIdentifier(country: Country, raw: string): string {
  if (!isPhoneInput(raw)) return raw; // email — send as-is
  // Strip leading 0 (common in India: 09876… → 9876…)
  const digits = raw.replace(/^\+/, "").replace(/^0/, "");
  // If user already typed the full dial code, don't double-prepend
  if (raw.startsWith("+")) return `+${digits}`;
  return `${country.dial}${digits}`;
}

export default function OtpLogin({
  onSuccess,
  title = "Sign in to continue",
  subtitle = "Enter your mobile number or email — we'll send you a one-time code.",
}: Props) {
  const [country, setCountry] = useState<Country>({ name: "India", code: "IN", dial: "+91", flag: "🇮🇳" });
  const [raw, setRaw] = useState("");
  const [otp, setOtp] = useState("");
  const [step, setStep] = useState<"id" | "otp">("id");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [msg, setMsg] = useState("");
  const [rememberMe, setRememberMe] = useState(false);

  // Auto-detect country from browser timezone
  useEffect(() => {
    setCountry(detectCountry());
  }, []);

  const showPicker = isPhoneInput(raw) || raw === "";

  async function handleRequestOTP(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const identifier = buildIdentifier(country, raw);
      await requestOTP(identifier);
      setStep("otp");
      setMsg("OTP sent! Check your phone or email.");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to send OTP");
    } finally {
      setLoading(false);
    }
  }

  async function handleVerify(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const identifier = buildIdentifier(country, raw);
      await verifyOTP(identifier, otp, rememberMe);
      onSuccess();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Invalid OTP");
    } finally {
      setLoading(false);
    }
  }

  const displayIdentifier = isPhoneInput(raw) ? buildIdentifier(country, raw) : raw;

  return (
    <div>
      <h2 className="font-display text-3xl mb-2">{title}</h2>
      <p className="text-sm text-muted mb-8">{subtitle}</p>

      {step === "id" ? (
        <form onSubmit={handleRequestOTP} className="space-y-4">
          {/* Phone / email input with country picker */}
          <div className="flex border border-line focus-within:border-clay transition-colors">
            {showPicker && (
              <CountryCodePicker selected={country} onChange={setCountry} />
            )}
            <input
              type="text"
              inputMode={showPicker ? "numeric" : "text"}
              className="flex-1 px-3 py-3 text-sm focus:outline-none bg-transparent"
              placeholder={showPicker ? "Mobile number" : "Mobile number or email"}
              value={raw}
              onChange={(e) => setRaw(e.target.value)}
              required
              autoFocus
            />
          </div>
          <p className="text-[11px] text-muted -mt-2">
            {showPicker
              ? `We'll send an OTP to ${displayIdentifier || country.dial + "…"}`
              : "We'll send an OTP to your email"}
          </p>

          <label className="flex items-center gap-2 text-sm text-muted cursor-pointer select-none">
            <input
              type="checkbox"
              checked={rememberMe}
              onChange={(e) => setRememberMe(e.target.checked)}
              className="accent-clay w-4 h-4"
            />
            Remember me on this device
          </label>
          {error && <p className="text-red-500 text-xs">{error}</p>}
          <button type="submit" disabled={loading} className="btn-primary w-full">
            {loading ? "Sending…" : "Send OTP"}
          </button>
        </form>
      ) : (
        <form onSubmit={handleVerify} className="space-y-4">
          {msg && <p className="text-clay text-xs">{msg}</p>}
          <p className="text-sm text-muted">
            Code sent to <strong>{displayIdentifier}</strong>
          </p>
          <input
            type="text"
            inputMode="numeric"
            className="input-field text-center tracking-[0.4em] text-2xl"
            placeholder="000000"
            maxLength={6}
            value={otp}
            onChange={(e) => setOtp(e.target.value)}
            required
            autoFocus
          />
          {error && <p className="text-red-500 text-xs">{error}</p>}
          <button type="submit" disabled={loading} className="btn-primary w-full">
            {loading ? "Verifying…" : "Verify & Continue"}
          </button>
          <button
            type="button"
            onClick={() => { setStep("id"); setError(""); setOtp(""); setMsg(""); }}
            className="w-full text-xs text-muted hover:text-clay"
          >
            Use a different number / email
          </button>
        </form>
      )}
    </div>
  );
}
