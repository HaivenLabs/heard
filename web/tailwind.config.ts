import type { Config } from "tailwindcss";

const tokenColor = (name: string) => `rgb(var(${name}) / <alpha-value>)`;

const config: Config = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx}",
    "./components/**/*.{js,ts,jsx,tsx}",
    "./lib/**/*.{js,ts,jsx,tsx}"
  ],
  theme: {
    extend: {
      colors: {
        canvas: tokenColor("--hv-surface-canvas-rgb"),
        surface: tokenColor("--hv-surface-raised-rgb"),
        parchment: tokenColor("--hv-surface-subtle-rgb"),
        ink: tokenColor("--hv-text-primary-rgb"),
        primary: tokenColor("--hv-action-primary-rgb"),
        spruce: tokenColor("--hv-action-primary-hover-rgb"),
        teal: tokenColor("--hv-action-primary-rgb"),
        sage: tokenColor("--hv-color-moss-rgb"),
        mist: tokenColor("--hv-action-primary-soft-rgb"),
        sand: tokenColor("--hv-color-line-rgb"),
        // Compatibility aliases. Use primary, teal, and sage for new UI work.
        clay: tokenColor("--hv-action-primary-hover-rgb"),
        olive: tokenColor("--hv-action-primary-rgb")
      },
      boxShadow: {
        soft: "var(--hv-shadow-md)"
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
