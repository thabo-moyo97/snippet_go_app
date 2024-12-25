/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./html/**/*.{html,js,tmpl,go.tmpl,gohtml,gotmpl,tmpl.html,go.html}",
    "./static/**/*.{html,js}"
  ],
  theme: {
    extend: {
      colors: {
        primary: '#007bff',
        secondary: '#6c757d',
        success: '#28a745',
        danger: '#dc3545',
        warning: '#ffc107',
        info: '#17a2b8',
        light: '#f8f9fa',
        dark: '#343a40',
      },
    },
  },
  plugins: [],
} 