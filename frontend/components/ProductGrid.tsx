import type { Product } from "@/lib/types";
import ProductCard from "./ProductCard";

export default function ProductGrid({
  products,
  empty = "No pieces found.",
}: {
  products: Product[];
  empty?: string;
}) {
  if (!products.length) {
    return (
      <p className="text-sm text-muted py-16 text-center">{empty}</p>
    );
  }
  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 md:gap-7">
      {products.map((p) => (
        <ProductCard key={p.id} product={p} />
      ))}
    </div>
  );
}
