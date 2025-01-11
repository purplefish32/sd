/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./views/**/*.{templ,html}"],
  theme: {
    extend: {
      colors: {
        'sd-dark': '#1a1a1a',
        'sd-darker': '#0f0f0f',
        'sd-light': '#2d2d2d',
        'sd-accent': '#00ff00',
      },
    },
  },
  plugins: [],
}
