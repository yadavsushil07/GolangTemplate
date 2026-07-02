/** Format integer paise/cents into an Indian-Rupee string, e.g. 2435000 -> "₹24,350". */
export function formatPrice(cents: number): string {
  const rupees = Math.round((cents ?? 0) / 100);
  return "₹" + rupees.toLocaleString("en-IN");
}

/** Slugify used only for display fallbacks; the backend owns canonical slugs. */
export function initials(name: string): string {
  return (name || "?").trim().charAt(0).toUpperCase();
}
