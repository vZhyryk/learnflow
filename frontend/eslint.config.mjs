// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt(
  {
    ignores: ['**/dist/**', '**/dist-ssr/**', '**/.nuxt/**', '**/.output/**', '**/coverage/**'],
  },
  {
    files: ['**/*.vue'],
    rules: {
      'vue/multi-word-component-names': 'off',
      'vue/no-unused-components': 'error',
      'vue/no-unused-vars': 'error',
      'vue/component-name-in-template-casing': ['error', 'PascalCase'],
      'vue/no-v-html': 'warn',
      'vue/no-mutating-props': 'error',
      'vue/no-setup-props-reactivity-loss': 'error',
      'vue/no-ref-as-operand': 'error',
      'vue/no-template-shadow': 'error',
      'vue/block-order': ['error', { order: ['script', 'template', 'style'] }],
      'vue/define-macros-order': ['error', { order: ['defineProps', 'defineEmits'] }],
    },
  },
  {
    files: ['**/*.{ts,tsx,vue}'],
    rules: {
      'no-unused-vars': 'off',
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/no-unused-vars': 'error',
      '@typescript-eslint/consistent-type-imports': 'error',
      '@typescript-eslint/no-non-null-assertion': 'error',
    },
  },
  {
    files: ['**/*.{ts,tsx,vue}'],
    rules: {
      complexity: ['error', 10],
      'max-depth': ['error', 4],
      'max-lines': ['error', 300],
      'max-lines-per-function': ['error', 60],
      'max-params': ['error', 3],
      'no-console': 'warn',
      'no-debugger': 'error',
    },
  },
  {
    files: ['components/**/*.{ts,tsx,vue}', 'pages/**/*.{ts,tsx,vue}'],
    rules: {
      'no-restricted-syntax': [
        'error',
        {
          selector: 'CallExpression[callee.name="fetch"]',
          message: 'Use composables or shared data-access helpers instead of fetch directly in UI components.',
        },
      ],
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: ['@/services/*', '@/api/*'],
              message: 'Do not import services/api directly here. Use composables instead.',
            },
          ],
        },
      ],
    },
  },
)
