/** Scans the templates of every module and of resources/views. */
module.exports = {
  darkMode: "class",
  content: ["./app-modules/*/resources/views/**/*.tmpl", "./resources/views/**/*.tmpl"],
  theme: {
    extend: {
      colors: {
        brand: { navy: "#1A3C5E", amber: "#E8A020", emerald: "#27A06E", cloud: "#F5F7FA", coral: "#E84B4B" },
        "page-dark": "#0b1220",
      },
      fontFamily: {
        latin: ["Inter", "ui-sans-serif", "system-ui", "sans-serif"],
        bangla: ["Noto Sans Bengali", "Inter", "ui-sans-serif", "system-ui", "sans-serif"],
      },
    },
  },
  plugins: [],
};
