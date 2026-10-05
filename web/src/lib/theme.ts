import { ref } from 'vue'

export type Theme = 'light' | 'dark'

// index.html's boot script already stamped data-theme on <html>, so just
// read it back — no flash, no double source of truth.
export const theme = ref<Theme>(
  document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light',
)

export function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  document.documentElement.dataset.theme = theme.value
  try {
    localStorage.setItem('bp-theme', theme.value)
  } catch {
    // private mode etc. — theme still applies for this session
  }
}
