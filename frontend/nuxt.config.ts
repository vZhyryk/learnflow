import * as v from 'valibot'

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },

  modules: [
    '@nuxt/eslint',
    '@nuxt/image',
    '@nuxt/scripts',
    '@nuxt/a11y',
    '@nuxtjs/i18n',
    '@nuxtjs/seo',
    '@pinia/nuxt',
    'dayjs-nuxt',
    '@vueuse/nuxt',
    '@vee-validate/nuxt',
    'nuxt-capo',
    'nuxt-delay-hydration',
    'nuxt-charts',
    'nuxt-gtag',
    'nuxt-payload-analyzer',
    'nuxt-security',
    'nuxt-typed-router',
    'nuxt-safe-runtime-config',
    'vue3-carousel-nuxt',
    'nuxt-toast',
    'nuxt-posthog',
    '@nuxt/ui',
  ],

  safeRuntimeConfig: {
    $schema: v.object({})
  }
})
