import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx}",
    "./components/**/*.{js,ts,jsx,tsx}",
    "./lib/**/*.{js,ts,jsx,tsx}"
  ],
  theme: {
    extend: {
      colors: {
        parchment: "#f7f1e3",
        ink: "#111827",
        clay: "#cb6843",
        olive: "#6d7d46",
        mist: "#e8e1d4",
        sand: "#d8c9b4"
      },
      boxShadow: {
        soft: "0 24px 60px rgba(17, 24, 39, 0.12)"
      },
      fontFamily: {
        display: ["Avenir Next", "Segoe UI", "Helvetica Neue", "sans-serif"],
        body: ["'Trebuchet MS'", "Avenir Next", "Segoe UI", "sans-serif"]
      }
    }
  },
  plugins: []
};

export default config;
