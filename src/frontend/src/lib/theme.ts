export type ThemePreference = 'light' | 'dark' | 'system'
export type ResolvedTheme = 'light' | 'dark'

export function parseThemePreference(value: string | null | undefined): ThemePreference {
  if (value === 'light' || value === 'dark' || value === 'system') {
    return value
  }
  return 'light'
}

export function resolveTheme(preference: ThemePreference, systemDark: boolean): ResolvedTheme {
  if (preference === 'system') {
    return systemDark ? 'dark' : 'light'
  }
  return preference
}

/** Matches --bg-window token RGB. */
export function windowBackgroundRGBA(resolved: ResolvedTheme): {
  R: number
  G: number
  B: number
  A: number
} {
  if (resolved === 'dark') {
    return { R: 28, G: 28, B: 30, A: 255 }
  }
  return { R: 245, G: 245, B: 247, A: 255 }
}
