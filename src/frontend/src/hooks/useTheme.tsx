import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { GetSetting, SetSetting } from '../../wailsjs/go/main/App'
import { WindowSetBackgroundColour } from '../../wailsjs/runtime/runtime'
import { SettingKey } from '../lib/settings'
import {
  parseThemePreference,
  parseThemeContrast,
  parseUiFontScale,
  parseEditorFontSize,
  editorFontSizePx,
  resolveTheme,
  windowBackgroundRGBA,
  type ResolvedTheme,
  type ThemeContrast,
  type ThemePreference,
  type UiFontScale,
  type EditorFontSize,
} from '../lib/theme'

type ThemeContextValue = {
  preference: ThemePreference
  resolved: ResolvedTheme
  contrast: ThemeContrast
  uiFontScale: UiFontScale
  editorFontSize: EditorFontSize
  setPreference: (preference: ThemePreference) => Promise<void>
  setContrast: (contrast: ThemeContrast) => Promise<void>
  setUiFontScale: (scale: UiFontScale) => Promise<void>
  setEditorFontSize: (size: EditorFontSize) => Promise<void>
  loading: boolean
}

const ThemeContext = createContext<ThemeContextValue | null>(null)

function systemPrefersDark(): boolean {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

function applyResolved(
  resolved: ResolvedTheme,
  contrast: ThemeContrast,
  uiFontScale: UiFontScale,
  editorFontSize: EditorFontSize,
) {
  const root = document.documentElement
  root.dataset.theme = resolved
  if (contrast === 'high') {
    root.dataset.contrast = 'high'
  } else {
    delete root.dataset.contrast
  }
  if (uiFontScale === 'default') {
    delete root.dataset.uiScale
  } else {
    root.dataset.uiScale = uiFontScale
  }
  root.style.setProperty('--editor-font-size', editorFontSizePx(editorFontSize))
  const { R, G, B, A } = windowBackgroundRGBA(resolved)
  try {
    WindowSetBackgroundColour(R, G, B, A)
  } catch {
    // Browser / non-Wails preview: ignore
  }
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [preference, setPreferenceState] = useState<ThemePreference>('light')
  const [contrast, setContrastState] = useState<ThemeContrast>('normal')
  const [uiFontScale, setUiFontScaleState] = useState<UiFontScale>('default')
  const [editorFontSize, setEditorFontSizeState] = useState<EditorFontSize>('14')
  const [systemDark, setSystemDark] = useState(systemPrefersDark)
  const [loading, setLoading] = useState(true)

  const resolved = useMemo(
    () => resolveTheme(preference, systemDark),
    [preference, systemDark],
  )

  useEffect(() => {
    let cancelled = false
    Promise.all([
      GetSetting(SettingKey.Theme),
      GetSetting(SettingKey.ThemeContrast),
      GetSetting(SettingKey.UiFontScale),
      GetSetting(SettingKey.EditorFontSize),
    ])
      .then(([themeValue, contrastValue, uiScaleValue, editorSizeValue]) => {
        if (!cancelled) {
          setPreferenceState(parseThemePreference(themeValue))
          setContrastState(parseThemeContrast(contrastValue))
          setUiFontScaleState(parseUiFontScale(uiScaleValue))
          setEditorFontSizeState(parseEditorFontSize(editorSizeValue))
        }
      })
      .catch(() => {
        if (!cancelled) {
          setPreferenceState('light')
          setContrastState('normal')
          setUiFontScaleState('default')
          setEditorFontSizeState('14')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    const onChange = () => setSystemDark(mq.matches)
    onChange()
    mq.addEventListener('change', onChange)
    return () => mq.removeEventListener('change', onChange)
  }, [])

  useEffect(() => {
    applyResolved(resolved, contrast, uiFontScale, editorFontSize)
  }, [resolved, contrast, uiFontScale, editorFontSize])

  const setPreference = useCallback(async (next: ThemePreference) => {
    let previous: ThemePreference = 'light'
    setPreferenceState((prev) => {
      previous = prev
      return next
    })
    try {
      await SetSetting(SettingKey.Theme, next)
    } catch (err) {
      setPreferenceState(previous)
      throw err
    }
  }, [])

  const setContrast = useCallback(async (next: ThemeContrast) => {
    let previous: ThemeContrast = 'normal'
    setContrastState((prev) => {
      previous = prev
      return next
    })
    try {
      await SetSetting(SettingKey.ThemeContrast, next)
    } catch (err) {
      setContrastState(previous)
      throw err
    }
  }, [])

  const setUiFontScale = useCallback(async (next: UiFontScale) => {
    let previous: UiFontScale = 'default'
    setUiFontScaleState((prev) => {
      previous = prev
      return next
    })
    try {
      await SetSetting(SettingKey.UiFontScale, next)
    } catch (err) {
      setUiFontScaleState(previous)
      throw err
    }
  }, [])

  const setEditorFontSize = useCallback(async (next: EditorFontSize) => {
    let previous: EditorFontSize = '14'
    setEditorFontSizeState((prev) => {
      previous = prev
      return next
    })
    try {
      await SetSetting(SettingKey.EditorFontSize, next)
    } catch (err) {
      setEditorFontSizeState(previous)
      throw err
    }
  }, [])

  const value = useMemo(
    () => ({
      preference,
      resolved,
      contrast,
      uiFontScale,
      editorFontSize,
      setPreference,
      setContrast,
      setUiFontScale,
      setEditorFontSize,
      loading,
    }),
    [
      preference,
      resolved,
      contrast,
      uiFontScale,
      editorFontSize,
      setPreference,
      setContrast,
      setUiFontScale,
      setEditorFontSize,
      loading,
    ],
  )

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export function useTheme(): ThemeContextValue {
  const ctx = useContext(ThemeContext)
  if (!ctx) {
    throw new Error('useTheme must be used within ThemeProvider')
  }
  return ctx
}
