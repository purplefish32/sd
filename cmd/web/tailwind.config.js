/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./views/**/*.{templ,html}"],
  theme: {
    extend: {
      // borderWidth: {
      //   DEFAULT: "1px",
      //   0: "0",
      //   2: "2px",
      //   4: "4px",
      //   8: "8px",
      //   100: "100px",
      // },
      colors: {
        "sd-dark": "#1a1aaa",
        "sd-darker": "#0f0f0f",
        "sd-light": "#2d2d2d",
        "sd-accent": "#00ffff",
      },
    },
  },
  plugins: [],
};
