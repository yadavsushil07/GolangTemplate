interface StatCardProps {
  label: string;
  value: string;
  sub?: string;
  accent?: "default" | "green" | "amber" | "red";
}

const ACCENT_STYLES = {
  default: "border-l-[#43293A]",
  green: "border-l-emerald-500",
  amber: "border-l-amber-500",
  red: "border-l-red-400",
};

export default function StatCard({
  label,
  value,
  sub,
  accent = "default",
}: StatCardProps) {
  return (
    <div
      className={`bg-white rounded-sm border border-[#E4DAC9] border-l-4 ${ACCENT_STYLES[accent]} p-6 shadow-sm`}
    >
      <p className="text-[10px] tracking-[0.22em] uppercase text-[#8B8175] mb-2">
        {label}
      </p>
      <p className="font-display text-3xl text-[#262019] leading-none mb-1">{value}</p>
      {sub && <p className="text-xs text-[#8B8175] mt-1">{sub}</p>}
    </div>
  );
}
