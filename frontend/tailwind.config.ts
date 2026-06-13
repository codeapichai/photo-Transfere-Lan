import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}", "./lib/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: "#17212b",
        field: "#f5f7f8",
        mint: "#36b37e",
        coral: "#e76f51",
        sky: "#3182ce"
      }
    }
  },
  plugins: []
};

export default config;

