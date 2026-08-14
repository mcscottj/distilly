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
  resolveTheme,
  windowBackgroundRGBA,
  type ResolvedTheme,
  type ThemeContrast,
  type ThemePreference,
} from '../lib/theme'

type ThemeContextValue = {
  preference: ThemePreference
  resolved: ResolvedTheme
  contrast: ThemeContrast
  setPreference: (preference: ThemePreference) => Promise<void>
  setContrast: (contrast: ThemeContrast) => Promise<void>
  loading: boolean
}

const ThemeContext = createContext<ThemeContextValue | null>(null)

function systemPrefersDark(): boolean {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

function applyResolved(resolved: ResolvedTheme, contrast: ThemeContrast) {
  document.documentElement.dataset.theme = resolved
  if (contrast === 'high') {
    document.documentElement.dataset.contrast = 'high'
  } else {
    delete document.documentElement.dataset.contrast
  }
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
  const [systemDark, setSystemDark] = useState(systemPrefersDark)
  const [loading, setLoading] = useState(true)

  const resolved = useMemo(
    () => resolveTheme(preference, systemDark),
    [preference, systemDark],
  )

  useEffect(() => {
    let cancelled = false
    Promise.all([GetSetting(SettingKey.Theme), GetSetting(SettingKey.ThemeContrast)])
      .then(([themeValue, contrastValue]) => {
        if (!cancelled) {
          setPreferenceState(parseThemePreference(themeValue))
          setContrastState(parseThemeContrast(contrastValue))
        }
      })
      .catch(() => {
        if (!cancelled) {
          setPreferenceState('light')
          setContrastState('normal')
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
    applyResolved(resolved, contrast)
  }, [resolved, contrast])

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

  const value = useMemo(
    () => ({ preference, resolved, contrast, setPreference, setContrast, loading }),
    [preference, resolved, contrast, setPreference, setContrast, loading],
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
