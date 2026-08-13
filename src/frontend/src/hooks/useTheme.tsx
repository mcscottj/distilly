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
  resolveTheme,
  windowBackgroundRGBA,
  type ResolvedTheme,
  type ThemePreference,
} from '../lib/theme'

type ThemeContextValue = {
  preference: ThemePreference
  resolved: ResolvedTheme
  setPreference: (preference: ThemePreference) => Promise<void>
  loading: boolean
}

const ThemeContext = createContext<ThemeContextValue | null>(null)

function systemPrefersDark(): boolean {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

function applyResolved(resolved: ResolvedTheme) {
  document.documentElement.dataset.theme = resolved
  const { R, G, B, A } = windowBackgroundRGBA(resolved)
  try {
    WindowSetBackgroundColour(R, G, B, A)
  } catch {
    // Browser / non-Wails preview: ignore
  }
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [preference, setPreferenceState] = useState<ThemePreference>('light')
  const [systemDark, setSystemDark] = useState(systemPrefersDark)
  const [loading, setLoading] = useState(true)

  const resolved = useMemo(
    () => resolveTheme(preference, systemDark),
    [preference, systemDark],
  )

  useEffect(() => {
    let cancelled = false
    GetSetting(SettingKey.Theme)
      .then((value) => {
        if (!cancelled) {
          setPreferenceState(parseThemePreference(value))
        }
      })
      .catch(() => {
        if (!cancelled) setPreferenceState('light')
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
    applyResolved(resolved)
  }, [resolved])

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

  const value = useMemo(
    () => ({ preference, resolved, setPreference, loading }),
    [preference, resolved, setPreference, loading],
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
