import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        cream: "#F7F3EC",
        oat: "#ECE3D5",
        warmwhite: "#F3ECDF",
        clay: "#B06A50",
        claydark: "#8E4F38",
        plum: "#43293A",
        plumdark: "#2C1926",
        wine: "#6E2F44",
        champ: "#C2A165",
        ink: "#262019",
        muted: "#8B8175",
        line: "#E4DAC9",
      },
      fontFamily: {
        display: ["var(--font-display)", "Georgia", "serif"],
        body: ["var(--font-body)", "system-ui", "sans-serif"],
      },
      boxShadow: {
        luxe: "0 18px 50px rgba(38,32,25,0.10)",
      },
    },
  },
  plugins: [],
};

export default config;
