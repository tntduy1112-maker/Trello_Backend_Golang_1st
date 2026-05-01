/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        trello: {
          blue: '#0079bf',
          'blue-hover': '#026aa7',
          green: '#61bd4f',
          yellow: '#f2d600',
          orange: '#ff9f1a',
          red: '#eb5a46',
          purple: '#c377e0',
          pink: '#ff78cb',
          sky: '#00c2e0',
          lime: '#51e898',
          gray: '#838c91',
        }
      },
    },
  },
  plugins: [],
}
