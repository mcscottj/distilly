import { describe, expect, it } from 'vitest'
import {
  parseThemePreference,
  resolveTheme,
  windowBackgroundRGBA,
} from './theme'

describe('parseThemePreference', () => {
  it('defaults invalid or empty to light', () => {
    expect(parseThemePreference(undefined)).toBe('light')
    expect(parseThemePreference(null)).toBe('light')
    expect(parseThemePreference('')).toBe('light')
    expect(parseThemePreference('nope')).toBe('light')
  })

  it('accepts light, dark, system', () => {
    expect(parseThemePreference('light')).toBe('light')
    expect(parseThemePreference('dark')).toBe('dark')
    expect(parseThemePreference('system')).toBe('system')
  })
})

describe('resolveTheme', () => {
  it('resolves explicit preferences', () => {
    expect(resolveTheme('light', true)).toBe('light')
    expect(resolveTheme('dark', false)).toBe('dark')
  })

  it('follows system when preference is system', () => {
    expect(resolveTheme('system', true)).toBe('dark')
    expect(resolveTheme('system', false)).toBe('light')
  })
})

describe('windowBackgroundRGBA', () => {
  it('returns light and dark window fills', () => {
    expect(windowBackgroundRGBA('light')).toEqual({ R: 245, G: 245, B: 247, A: 255 })
    expect(windowBackgroundRGBA('dark')).toEqual({ R: 28, G: 28, B: 30, A: 255 })
  })
})
