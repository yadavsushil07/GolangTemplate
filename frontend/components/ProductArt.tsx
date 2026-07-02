// Deterministic warm-toned gradient placeholder used until real photography
// is uploaded. Same product id always renders the same swatch.
const GRADIENTS = [
  "linear-gradient(135deg,#E7D3C4,#cfa98f)",
  "linear-gradient(135deg,#DED4BD,#bca97f)",
  "linear-gradient(135deg,#D7C4C9,#b0909d)",
  "linear-gradient(135deg,#D9BFB6,#bf9384)",
  "linear-gradient(135deg,#d8c9b0,#c4ac85)",
];

export default function ProductArt({
  seed,
  imageUrl,
  label,
  className = "",
}: {
  seed: number;
  imageUrl?: string;
  label?: string;
  className?: string;
}) {
  if (imageUrl) {
    // eslint-disable-next-line @next/next/no-img-element
    return (
      <img
        src={imageUrl}
        alt={label || "product"}
        className={`w-full h-full object-cover ${className}`}
      />
    );
  }
  const bg = GRADIENTS[Math.abs(seed) % GRADIENTS.length];
  return (
    <div
      className={`w-full h-full ${className}`}
      style={{ background: bg }}
      aria-hidden
    />
  );
}
