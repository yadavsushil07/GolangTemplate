"use client";

import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";

const STATUS_COLORS: Record<string, string> = {
  placed: "#C2A165",
  confirmed: "#43293A",
  shipped: "#B06A50",
  delivered: "#4CAF84",
  cancelled: "#E57373",
};

const STATUS_LABELS: Record<string, string> = {
  placed: "Placed",
  confirmed: "Confirmed",
  shipped: "Shipped",
  delivered: "Delivered",
  cancelled: "Cancelled",
};

interface Props {
  ordersByStatus: Record<string, number>;
}

export default function OrderStatusChart({ ordersByStatus }: Props) {
  const data = Object.entries(ordersByStatus)
    .filter(([, v]) => v > 0)
    .map(([status, count]) => ({
      name: STATUS_LABELS[status] ?? status,
      value: count,
      color: STATUS_COLORS[status] ?? "#8B8175",
    }));

  const total = data.reduce((s, d) => s + d.value, 0);

  if (data.length === 0) {
    return (
      <div className="bg-white border border-[#E4DAC9] rounded-sm p-6 shadow-sm flex items-center justify-center min-h-[280px]">
        <p className="text-[11px] tracking-[0.18em] uppercase text-[#8B8175]">No orders yet</p>
      </div>
    );
  }

  return (
    <div className="bg-white border border-[#E4DAC9] rounded-sm p-6 shadow-sm">
      <p className="text-[10px] tracking-[0.22em] uppercase text-[#8B8175] mb-1">
        Orders by Status
      </p>
      <p className="font-display text-2xl text-[#262019] mb-5">
        {total} total
      </p>
      <ResponsiveContainer width="100%" height={220}>
        <PieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            innerRadius={60}
            outerRadius={90}
            paddingAngle={3}
            dataKey="value"
          >
            {data.map((entry, i) => (
              <Cell key={i} fill={entry.color} />
            ))}
          </Pie>
          <Tooltip
            formatter={(val: number) => [val, "Orders"]}
            contentStyle={{
              border: "1px solid #E4DAC9",
              borderRadius: 2,
              fontSize: 12,
            }}
          />
          <Legend
            iconType="circle"
            iconSize={8}
            formatter={(value) => (
              <span style={{ fontSize: 11, color: "#8B8175", letterSpacing: "0.1em" }}>
                {value.toUpperCase()}
              </span>
            )}
          />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}
