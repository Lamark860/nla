export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: true },

  modules: [
    '@nuxtjs/color-mode',
  ],

  css: [
    'bootstrap/dist/css/bootstrap.min.css',
    'bootstrap-icons/font/bootstrap-icons.css',
    '~/assets/css/main.css',
  ],

  colorMode: {
    dataValue: 'theme',
    preference: 'system',
    fallback: 'light',
  },

  runtimeConfig: {
    apiBaseServer: 'http://localhost:8085/api/v1',
    public: {
      apiBase: '/api/v1',
    },
  },

  // Dev proxy → Go API
  nitro: {
    devProxy: {
      '/api': {
        target: 'http://localhost:8085',
        changeOrigin: true,
      },
    },
  },

  app: {
    head: {
      title: 'NLA — Анализатор облигаций',
      htmlAttrs: { lang: 'ru' },
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'Профессиональный анализатор облигаций MOEX с AI-оценкой' },
        { name: 'theme-color', content: '#111827' },
      ],
      link: [
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap' },
      ],
    },
  },
})
