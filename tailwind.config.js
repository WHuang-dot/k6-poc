/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./front-end/**/*.html"],
  theme: {
    extend: {},
  },
  plugins: [
    require('@tailwindcss/typography')
  ],
}

