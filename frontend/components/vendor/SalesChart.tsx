"use client";

import { useState, useMemo } from "react";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";

interface MonthlyPoint {
  year: number;
  month: number;
  revenue_cents: number;
  order_count: number;
}

const MONTH_NAMES = [
  "Jan", "Feb", "Mar", "Apr", "May", "Jun",
  "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
];

const YEAR_OPTIONS = () => {
  const now = new Date().getFullYear();
  return [now, now - 1, now - 2];
};

type FilterMode = "monthly" | "yearly";

interface Props {
  data: MonthlyPoint[];
}

function formatRupees(cents: number) {
  const rupees = Math.round(cents / 100);
  if (rupees >= 100000) return `₹${(rupees / 100000).toFixed(1)}L`;
  if (rupees >= 1000) return `₹${(rupees / 1000).toFixed(1)}K`;
  return `₹${rupees}`;
}

export default function SalesChart({ data }: Props) {
  const [mode, setMode] = useState<FilterMode>("monthly");
  const [selectedYear, setSelectedYear] = useState(new Date().getFullYear());

  const chartData = useMemo(() => {
    if (mode === "monthly") {
      const filtered = data.filter((d) => d.year === selectedYear);
      const map: Record<number, MonthlyPoint> = {};
      filtered.forEach((d) => { map[d.month] = d; });
      return Array.from({ length: 12 }, (_, i) => ({
        label: MONTH_NAMES[i],
        revenue: Math.round((map[i + 1]?.revenue_cents ?? 0) / 100),
        orders: map[i + 1]?.order_count ?? 0,
      }));
    }
    // yearly — group by year
    const yearMap: Record<number, { revenue: number; orders: number }> = {};
    data.forEach((d) => {
      if (!yearMap[d.year]) yearMap[d.year] = { revenue: 0, orders: 0 };
      yearMap[d.year].revenue += Math.round(d.revenue_cents / 100);
      yearMap[d.year].orders += d.order_count;
    });
    return Object.entries(yearMap)
      .sort(([a], [b]) => Number(a) - Number(b))
      .map(([year, v]) => ({ label: year, revenue: v.revenue, orders: v.orders }));
  }, [data, mode, selectedYear]);

  const totalRevenue = chartData.reduce((s, d) => s + d.revenue, 0);
  const totalOrders = chartData.reduce((s, d) => s + d.orders, 0);

  return (
    <div className="bg-white border border-[#E4DAC9] rounded-sm p-6 shadow-sm">
      <div className="flex flex-wrap items-start justify-between gap-4 mb-6">
        <div>
          <p className="text-[10px] tracking-[0.22em] uppercase text-[#8B8175] mb-1">
            Sales Overview
          </p>
          <p className="font-display text-2xl text-[#262019]">
            {formatRupees(totalRevenue * 100)}
            <span className="text-sm font-body text-[#8B8175] ml-2">
              {totalOrders} orders
            </span>
          </p>
        </div>
        <div className="flex items-center gap-2 flex-wrap">
          <div className="flex border border-[#E4DAC9] rounded overflow-hidden text-[11px]">
            <button
              onClick={() => setMode("monthly")}
              className={`px-4 py-1.5 tracking-[0.1em] uppercase transition-colors ${
                mode === "monthly" ? "bg-[#43293A] text-white" : "text-[#8B8175] hover:bg-[#f0ebe3]"
              }`}
            >
              Monthly
            </button>
            <button
              onClick={() => setMode("yearly")}
              className={`px-4 py-1.5 tracking-[0.1em] uppercase transition-colors ${
                mode === "yearly" ? "bg-[#43293A] text-white" : "text-[#8B8175] hover:bg-[#f0ebe3]"
              }`}
            >
              Yearly
            </button>
          </div>
          {mode === "monthly" && (
            <select
              value={selectedYear}
              onChange={(e) => setSelectedYear(Number(e.target.value))}
              className="border border-[#E4DAC9] bg-transparent text-sm px-3 py-1.5 focus:outline-none"
            >
              {YEAR_OPTIONS().map((y) => (
                <option key={y} value={y}>{y}</option>
              ))}
            </select>
          )}
        </div>
      </div>
      <ResponsiveContainer width="100%" height={260}>
        <AreaChart data={chartData} margin={{ top: 5, right: 10, left: 0, bottom: 0 }}>
          <defs>
            <linearGradient id="revenueGrad" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#43293A" stopOpacity={0.15} />
              <stop offset="95%" stopColor="#43293A" stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="#E4DAC9" vertical={false} />
          <XAxis
            dataKey="label"
            tick={{ fontSize: 10, fill: "#8B8175" }}
            axisLine={false}
            tickLine={false}
          />
          <YAxis
            tickFormatter={formatRupees}
            tick={{ fontSize: 10, fill: "#8B8175" }}
            axisLine={false}
            tickLine={false}
            width={52}
          />
          <Tooltip
            formatter={(val: number) => [`₹${val.toLocaleString("en-IN")}`, "Revenue"]}
            contentStyle={{
              border: "1px solid #E4DAC9",
              borderRadius: 2,
              fontSize: 12,
              background: "#fff",
            }}
          />
          <Area
            type="monotone"
            dataKey="revenue"
            stroke="#43293A"
            strokeWidth={2}
            fill="url(#revenueGrad)"
            dot={false}
            activeDot={{ r: 4, fill: "#43293A" }}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
