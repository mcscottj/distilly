export type ThemePreference = 'light' | 'dark' | 'system'
export type ResolvedTheme = 'light' | 'dark'
export type ThemeContrast = 'normal' | 'high'
export type UiFontScale = 'default' | 'large' | 'xlarge'
export type EditorFontSize = '12' | '14' | '16' | '18'

export function parseThemePreference(value: string | null | undefined): ThemePreference {
  if (value === 'light' || value === 'dark' || value === 'system') {
    return value
  }
  return 'light'
}

export function parseThemeContrast(value: string | null | undefined): ThemeContrast {
  if (value === 'normal' || value === 'high') {
    return value
  }
  return 'normal'
}

export function parseUiFontScale(value: string | null | undefined): UiFontScale {
  if (value === 'default' || value === 'large' || value === 'xlarge') {
    return value
  }
  return 'default'
}

export function parseEditorFontSize(value: string | null | undefined): EditorFontSize {
  if (value === '12' || value === '14' || value === '16' || value === '18') {
    return value
  }
  return '14'
}

export function editorFontSizePx(size: EditorFontSize): string {
  return `${size}px`
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
