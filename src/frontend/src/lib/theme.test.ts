import { describe, expect, it } from 'vitest'
import {
  parseThemePreference,
  parseThemeContrast,
  parseUiFontScale,
  parseEditorFontSize,
  editorFontSizePx,
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

describe('parseThemeContrast', () => {
  it('defaults invalid or empty to normal', () => {
    expect(parseThemeContrast(undefined)).toBe('normal')
    expect(parseThemeContrast(null)).toBe('normal')
    expect(parseThemeContrast('')).toBe('normal')
    expect(parseThemeContrast('nope')).toBe('normal')
  })

  it('accepts normal and high', () => {
    expect(parseThemeContrast('normal')).toBe('normal')
    expect(parseThemeContrast('high')).toBe('high')
  })
})

describe('parseUiFontScale', () => {
  it('defaults invalid or empty to default', () => {
    expect(parseUiFontScale(undefined)).toBe('default')
    expect(parseUiFontScale(null)).toBe('default')
    expect(parseUiFontScale('')).toBe('default')
    expect(parseUiFontScale('nope')).toBe('default')
  })

  it('accepts default, large, xlarge', () => {
    expect(parseUiFontScale('default')).toBe('default')
    expect(parseUiFontScale('large')).toBe('large')
    expect(parseUiFontScale('xlarge')).toBe('xlarge')
  })
})

describe('parseEditorFontSize', () => {
  it('defaults invalid or empty to 14', () => {
    expect(parseEditorFontSize(undefined)).toBe('14')
    expect(parseEditorFontSize(null)).toBe('14')
    expect(parseEditorFontSize('')).toBe('14')
    expect(parseEditorFontSize('20')).toBe('14')
    expect(parseEditorFontSize('nope')).toBe('14')
  })

  it('accepts 12, 14, 16, 18', () => {
    expect(parseEditorFontSize('12')).toBe('12')
    expect(parseEditorFontSize('14')).toBe('14')
    expect(parseEditorFontSize('16')).toBe('16')
    expect(parseEditorFontSize('18')).toBe('18')
  })
})

describe('editorFontSizePx', () => {
  it('appends px', () => {
    expect(editorFontSizePx('12')).toBe('12px')
    expect(editorFontSizePx('14')).toBe('14px')
    expect(editorFontSizePx('18')).toBe('18px')
  })
})
