# Native macOS Theme Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restyle the Distilly Wails desktop app as a light-first macOS utility with sidebar navigation, wood-brown accent from the logo, and a Light / Dark / System preference.

**Architecture:** Semantic CSS tokens on `:root` / `[data-theme="dark"]`, resolved by a `useTheme` hook that persists `theme` in SQLite and sets `document.documentElement.dataset.theme`. App shell becomes a left sidebar; Settings uses grouped lists + Appearance control; Lint/Dashboard keep layouts and swap hardcoded dark colors for tokens.

**Tech Stack:** Wails v2, React 19, Vite 7, Tailwind CSS v4, TypeScript, SQLite settings via existing `GetSetting` / `SetSetting`, Vitest for pure theme helpers.

**Spec:** [`docs/specs/2026-08-13-native-macos-theme-design.md`](../specs/2026-08-13-native-macos-theme-design.md)

## Global Constraints

- Accent: wood brown `#8B5E3C` (light) / `#C4A484` (dark) — not system blue
- Surfaces stay neutral macOS gray/white (and dark equivalents); do not wood-wash the window
- Default preference: `light`; values: `light` | `dark` | `system`; SQLite key: `theme`
- Keep default Wails titlebar; no frameless / vibrancy / traffic-light inset
- No new component library; optional small helpers only (`SidebarNav`, `GroupedList`, `SegmentedControl`)
- Lint and Dashboard behavior unchanged; restyle only
- Branch: `feature/native-macos-theme` (already created — do not implement on `main`)

---

## File map

| File | Responsibility |
|------|----------------|
| `src/frontend/src/lib/theme.ts` | Pure preference parse/resolve + window RGB helpers |
| `src/frontend/src/lib/theme.test.ts` | Vitest coverage for theme helpers |
| `src/frontend/src/lib/settings.ts` | Add `SettingKey.Theme` + default `light` |
| `src/frontend/src/style.css` | Token definitions + Tailwind `@theme` color mapping |
| `src/frontend/src/hooks/useTheme.tsx` | Load/persist preference; set `data-theme`; sync Wails bg |
| `src/frontend/src/main.tsx` | Wrap app in `ThemeProvider` |
| `src/main.go` | Light default `BackgroundColour` |
| `src/frontend/src/assets/distilly-logo.png` | Bundled logo copy |
| `src/frontend/src/components/SidebarNav.tsx` | Sidebar brand + nav |
| `src/frontend/src/components/SegmentedControl.tsx` | Appearance control |
| `src/frontend/src/components/GroupedList.tsx` | Settings inset groups |
| `src/frontend/src/App.tsx` | Sidebar shell layout |
| `src/frontend/src/pages/Settings.tsx` | Appearance + grouped restyle |
| `src/frontend/src/pages/LintWorkspace.tsx` | Token restyle |
| `src/frontend/src/pages/Dashboard.tsx` | Token restyle |
| `src/frontend/src/components/{ScoreCard,SectionBreakdown,SuggestionList,DiffView}.tsx` | Token restyle |
| `src/frontend/package.json` | Add `vitest` + `test` script |

### Shared token class vocabulary (use everywhere)

| Role | Tailwind classes (after `@theme` mapping) |
|------|-------------------------------------------|
| Window / sidebar / surface / fill | `bg-window`, `bg-sidebar`, `bg-surface`, `bg-fill` |
| Borders | `border-hairline` |
| Text | `text-fg`, `text-muted` |
| Accent | `bg-accent`, `text-accent`, `text-accent-fg`, `bg-accent-soft` |
| Focus | `focus:border-accent focus:ring-1 focus:ring-accent` |
| Danger / success / warning | `text-danger`, `bg-danger-soft`, `border-danger`, `text-success`, `text-warning` |
| Primary button | `rounded-md bg-accent px-3 py-1.5 text-sm font-medium text-accent-fg hover:opacity-90 disabled:opacity-50` |
| Secondary button | `rounded-md border border-hairline bg-fill px-3 py-1.5 text-sm text-fg hover:bg-surface disabled:opacity-50` |
| Text field | `rounded-md border border-hairline bg-fill px-3 py-2 text-sm text-fg outline-none focus:border-accent focus:ring-1 focus:ring-accent` |

---

### Task 1: Theme helpers, SettingKey, CSS tokens, Vitest

**Files:**
- Create: `src/frontend/src/lib/theme.ts`
- Create: `src/frontend/src/lib/theme.test.ts`
- Modify: `src/frontend/src/lib/settings.ts`
- Modify: `src/frontend/src/style.css`
- Modify: `src/frontend/package.json`
- Modify: `src/frontend/vite.config.ts` (vitest config if needed)

**Interfaces:**
- Consumes: none
- Produces:
  - `export type ThemePreference = 'light' | 'dark' | 'system'`
  - `export type ResolvedTheme = 'light' | 'dark'`
  - `export function parseThemePreference(value: string | null | undefined): ThemePreference`
  - `export function resolveTheme(preference: ThemePreference, systemDark: boolean): ResolvedTheme`
  - `export function windowBackgroundRGBA(resolved: ResolvedTheme): { R: number; G: number; B: number; A: number }`
  - `SettingKey.Theme` = `'theme'`, default `'light'`
  - CSS vars + `@theme` colors listed below

- [ ] **Step 1: Add Vitest**

In `src/frontend/`:

```bash
npm install -D vitest
```

Update `package.json` scripts:

```json
"test": "vitest run",
"test:watch": "vitest"
```

Update `vite.config.ts` to export vitest types (merge with existing config):

```ts
/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  test: {
    environment: 'node',
  },
})
```

- [ ] **Step 2: Write failing theme tests**

Create `src/frontend/src/lib/theme.test.ts`:

```ts
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
```

- [ ] **Step 3: Run tests — expect FAIL**

```bash
cd src/frontend && npm test
```

Expected: FAIL — cannot find module `./theme` or exports missing.

- [ ] **Step 4: Implement `theme.ts`**

Create `src/frontend/src/lib/theme.ts`:

```ts
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
```

- [ ] **Step 5: Run tests — expect PASS**

```bash
cd src/frontend && npm test
```

Expected: PASS (all theme tests).

- [ ] **Step 6: Add `SettingKey.Theme`**

In `src/frontend/src/lib/settings.ts`, add:

```ts
Theme: 'theme',
```

to `SettingKey`, and in `SETTING_DEFAULTS`:

```ts
[SettingKey.Theme]: 'light',
```

Update `withSettingDefault` so empty `theme` falls back to `'light'` (default path already applies for non-ApiKey/DefaultModel keys — verify `Theme` is not in the empty-keep list).

- [ ] **Step 7: Replace `style.css` with tokens**

Overwrite `src/frontend/src/style.css`:

```css
@import "tailwindcss";

:root {
  color-scheme: light;
  font-family: -apple-system, BlinkMacSystemFont, "SF Pro Text", "Segoe UI", sans-serif;
  --bg-window: #f5f5f7;
  --bg-sidebar: #ececef;
  --bg-surface: #ffffff;
  --bg-fill: #f0f0f2;
  --border: rgba(0, 0, 0, 0.1);
  --fg: #1d1d1f;
  --muted: #6e6e73;
  --accent: #8b5e3c;
  --accent-fg: #ffffff;
  --accent-soft: rgba(139, 94, 60, 0.14);
  --danger: #c0392b;
  --danger-soft: rgba(192, 57, 43, 0.12);
  --success: #1f7a4c;
  --warning: #b57900;
  color: var(--fg);
  background: var(--bg-window);
}

[data-theme="dark"] {
  color-scheme: dark;
  --bg-window: #1c1c1e;
  --bg-sidebar: #2c2c2e;
  --bg-surface: #3a3a3c;
  --bg-fill: #2c2c2e;
  --border: rgba(255, 255, 255, 0.12);
  --fg: #f5f5f7;
  --muted: #98989d;
  --accent: #c4a484;
  --accent-fg: #1c1c1e;
  --accent-soft: rgba(196, 164, 132, 0.22);
  --danger: #ff6b6b;
  --danger-soft: rgba(255, 107, 107, 0.16);
  --success: #6bcf8e;
  --warning: #e6b84d;
}

@theme {
  --font-sans: -apple-system, BlinkMacSystemFont, "SF Pro Text", "Segoe UI", sans-serif;
  --font-mono: ui-monospace, "SF Mono", Menlo, monospace;
  --color-window: var(--bg-window);
  --color-sidebar: var(--bg-sidebar);
  --color-surface: var(--bg-surface);
  --color-fill: var(--bg-fill);
  --color-hairline: var(--border);
  --color-fg: var(--fg);
  --color-muted: var(--muted);
  --color-accent: var(--accent);
  --color-accent-fg: var(--accent-fg);
  --color-accent-soft: var(--accent-soft);
  --color-danger: var(--danger);
  --color-danger-soft: var(--danger-soft);
  --color-success: var(--success);
  --color-warning: var(--warning);
}

html,
body,
#root {
  height: 100%;
  margin: 0;
}

body {
  background: var(--bg-window);
  color: var(--fg);
}
```

- [ ] **Step 8: Commit**

```bash
git add src/frontend/package.json src/frontend/package-lock.json src/frontend/vite.config.ts \
  src/frontend/src/lib/theme.ts src/frontend/src/lib/theme.test.ts \
  src/frontend/src/lib/settings.ts src/frontend/src/style.css
git commit -m "feat(ui): add theme tokens and preference helpers"
```

---

### Task 2: ThemeProvider + Wails background sync

**Files:**
- Create: `src/frontend/src/hooks/useTheme.tsx`
- Modify: `src/frontend/src/main.tsx`
- Modify: `src/main.go`

**Interfaces:**
- Consumes: `parseThemePreference`, `resolveTheme`, `windowBackgroundRGBA`, `SettingKey.Theme`, `GetSetting`, `SetSetting`, `WindowSetBackgroundColour`
- Produces:
  - `ThemeProvider({ children })`
  - `useTheme(): { preference: ThemePreference; resolved: ResolvedTheme; setPreference: (p: ThemePreference) => Promise<void>; loading: boolean }`

- [ ] **Step 1: Implement `useTheme.tsx`**

```tsx
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
    setPreferenceState(next)
    await SetSetting(SettingKey.Theme, next)
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
```

Note: apply light tokens immediately on first paint by setting `data-theme="light"` in `index.html` on `<html>` as a default attribute to avoid a flash.

- [ ] **Step 2: Set default on `index.html`**

In `src/frontend/index.html`, set `<html lang="en" data-theme="light">`.

- [ ] **Step 3: Wrap app in `main.tsx`**

```tsx
import React from 'react'
import { createRoot } from 'react-dom/client'
import './style.css'
import App from './App'
import { ThemeProvider } from './hooks/useTheme'

const container = document.getElementById('root')
const root = createRoot(container!)

root.render(
  <React.StrictMode>
    <ThemeProvider>
      <App />
    </ThemeProvider>
  </React.StrictMode>,
)
```

- [ ] **Step 4: Light default in `main.go`**

Change `BackgroundColour` to light window fill:

```go
BackgroundColour: &options.RGBA{R: 245, G: 245, B: 247, A: 255},
```

- [ ] **Step 5: Verify build**

```bash
cd src/frontend && npm test && npm run build
```

Expected: tests PASS; `tsc && vite build` succeeds.

- [ ] **Step 6: Commit**

```bash
git add src/frontend/src/hooks/useTheme.tsx src/frontend/src/main.tsx \
  src/frontend/index.html src/main.go
git commit -m "feat(ui): wire ThemeProvider and window background sync"
```

---

### Task 3: Sidebar shell + logo

**Files:**
- Create: `src/frontend/src/assets/distilly-logo.png` (copy from `docs/distilly-logo.png`)
- Create: `src/frontend/src/components/SidebarNav.tsx`
- Modify: `src/frontend/src/App.tsx`

**Interfaces:**
- Consumes: page ids `'lint' | 'dashboard' | 'settings'`
- Produces: `SidebarNav({ page, onNavigate })`

- [ ] **Step 1: Copy logo asset**

```bash
mkdir -p src/frontend/src/assets
cp docs/distilly-logo.png src/frontend/src/assets/distilly-logo.png
```

- [ ] **Step 2: Implement `SidebarNav.tsx`**

```tsx
import logo from '../assets/distilly-logo.png'

export type AppPage = 'lint' | 'dashboard' | 'settings'

const navItems: { id: AppPage; label: string }[] = [
  { id: 'lint', label: 'Lint' },
  { id: 'dashboard', label: 'Dashboard' },
  { id: 'settings', label: 'Settings' },
]

type SidebarNavProps = {
  page: AppPage
  onNavigate: (page: AppPage) => void
}

export function SidebarNav({ page, onNavigate }: SidebarNavProps) {
  return (
    <aside className="flex w-[220px] shrink-0 flex-col border-r border-hairline bg-sidebar">
      <div className="flex items-center gap-2 px-4 py-4">
        <img src={logo} alt="" className="size-9 object-contain" />
        <span className="text-[15px] font-semibold tracking-tight text-fg">distilly</span>
      </div>
      <nav className="flex flex-col gap-0.5 px-2 pb-4">
        {navItems.map((item) => {
          const active = page === item.id
          return (
            <button
              key={item.id}
              type="button"
              onClick={() => onNavigate(item.id)}
              className={
                active
                  ? 'rounded-md bg-accent-soft px-3 py-1.5 text-left text-sm font-medium text-accent'
                  : 'rounded-md px-3 py-1.5 text-left text-sm text-fg hover:bg-fill'
              }
            >
              {item.label}
            </button>
          )
        })}
      </nav>
    </aside>
  )
}
```

If Vite needs an image module declaration, add to `src/frontend/src/vite-env.d.ts`:

```ts
declare module '*.png' {
  const src: string
  export default src
}
```

- [ ] **Step 3: Rewrite `App.tsx` shell**

Replace top header with sidebar layout:

```tsx
import { useState } from 'react'
import { store } from '../wailsjs/go/models'
import { SidebarNav, type AppPage } from './components/SidebarNav'
import { Dashboard } from './pages/Dashboard'
import { LintWorkspace } from './pages/LintWorkspace'
import { Settings } from './pages/Settings'

function App() {
  const [page, setPage] = useState<AppPage>('lint')
  const [lintModel, setLintModel] = useState<string | undefined>(undefined)

  function openRequestInLint(request: store.Request) {
    if (request.model) {
      setLintModel(request.model)
    }
    setPage('lint')
  }

  return (
    <div className="flex h-full min-h-0 bg-window text-fg">
      <SidebarNav page={page} onNavigate={setPage} />
      <main className="min-w-0 flex-1 overflow-auto p-6">
        {page === 'lint' && (
          <LintWorkspace
            preferredModel={lintModel}
            onPreferredModelConsumed={() => setLintModel(undefined)}
          />
        )}
        {page === 'dashboard' && <Dashboard onOpenRequest={openRequestInLint} />}
        {page === 'settings' && <Settings />}
      </main>
    </div>
  )
}

export default App
```

- [ ] **Step 4: Verify build**

```bash
cd src/frontend && npm run build
```

Expected: SUCCESS. Manually (`wails dev` if available): sidebar visible with logo; nav switches pages; light gray chrome.

- [ ] **Step 5: Commit**

```bash
git add src/frontend/src/assets/distilly-logo.png src/frontend/src/components/SidebarNav.tsx \
  src/frontend/src/App.tsx src/frontend/src/vite-env.d.ts
git commit -m "feat(ui): add macOS-style sidebar shell with logo"
```

---

### Task 4: Settings — Appearance + grouped lists

**Files:**
- Create: `src/frontend/src/components/SegmentedControl.tsx`
- Create: `src/frontend/src/components/GroupedList.tsx`
- Modify: `src/frontend/src/pages/Settings.tsx`

**Interfaces:**
- Consumes: `useTheme().preference`, `useTheme().setPreference`
- Produces:
  - `SegmentedControl<T>({ options, value, onChange, disabled? })`
  - `GroupedList({ caption?, footer?, children })`
  - `GroupedRow({ label, description?, children })`

- [ ] **Step 1: Implement `SegmentedControl.tsx`**

```tsx
type Option<T extends string> = { value: T; label: string }

type SegmentedControlProps<T extends string> = {
  options: Option<T>[]
  value: T
  onChange: (value: T) => void
  disabled?: boolean
}

export function SegmentedControl<T extends string>({
  options,
  value,
  onChange,
  disabled,
}: SegmentedControlProps<T>) {
  return (
    <div className="inline-flex rounded-md border border-hairline bg-fill p-0.5">
      {options.map((opt) => {
        const active = opt.value === value
        return (
          <button
            key={opt.value}
            type="button"
            disabled={disabled}
            onClick={() => onChange(opt.value)}
            className={
              active
                ? 'rounded-[5px] bg-surface px-3 py-1 text-sm font-medium text-fg shadow-sm'
                : 'rounded-[5px] px-3 py-1 text-sm text-muted hover:text-fg'
            }
          >
            {opt.label}
          </button>
        )
      })}
    </div>
  )
}
```

- [ ] **Step 2: Implement `GroupedList.tsx`**

```tsx
import type { ReactNode } from 'react'

export function GroupedList({
  caption,
  footer,
  children,
}: {
  caption?: string
  footer?: ReactNode
  children: ReactNode
}) {
  return (
    <section className="space-y-2">
      {caption ? (
        <h3 className="px-1 text-xs font-medium uppercase tracking-wide text-muted">
          {caption}
        </h3>
      ) : null}
      <div className="overflow-hidden rounded-xl border border-hairline bg-surface">{children}</div>
      {footer ? <div className="px-1 text-xs text-muted">{footer}</div> : null}
    </section>
  )
}

export function GroupedRow({
  label,
  description,
  children,
}: {
  label: string
  description?: string
  children: ReactNode
}) {
  return (
    <div className="flex flex-col gap-2 border-t border-hairline px-4 py-3 first:border-t-0 sm:flex-row sm:items-center sm:justify-between">
      <div className="min-w-0 sm:max-w-[45%]">
        <p className="text-sm text-fg">{label}</p>
        {description ? <p className="mt-0.5 text-xs text-muted">{description}</p> : null}
      </div>
      <div className="min-w-0 sm:max-w-[55%] sm:text-right">{children}</div>
    </div>
  )
}
```

- [ ] **Step 3: Rewrite Settings page**

Keep existing hooks/`save`/`proxy` logic. Structure:

1. Page title + subtitle using `text-fg` / `text-muted`
2. Error banner: `border-danger bg-danger-soft text-danger`
3. **Appearance** `GroupedList` with `SegmentedControl` options Light/Dark/System — `onChange` calls `void setPreference(v)` (immediate; not part of Save)
4. **Upstream** / **Local proxy** / **Optimization** as `GroupedList` + `GroupedRow`; field classes use token vocabulary
5. Save button uses primary accent classes; “Saved” uses `text-success`

Do not put `theme` into `useSettings` / Save payload.

Appearance block sketch:

```tsx
const { preference, setPreference } = useTheme()

<GroupedList caption="Appearance">
  <GroupedRow label="Appearance" description="Choose light, dark, or match the system.">
    <SegmentedControl
      value={preference}
      onChange={(v) => void setPreference(v)}
      options={[
        { value: 'light', label: 'Light' },
        { value: 'dark', label: 'Dark' },
        { value: 'system', label: 'System' },
      ]}
    />
  </GroupedRow>
</GroupedList>
```

Replace all `sky-*`, `slate-*`, `white/*`, `black/*`, `text-white` classes with the shared vocabulary. Remove uppercase tracked “Status” / “Base URL” eyebrows — use sentence-case labels via `GroupedRow`.

- [ ] **Step 4: Manual verify**

```bash
cd src/frontend && npm run build
```

In the app: Settings → switch Light/Dark/System; UI flips immediately; restart app — preference persists; Save still writes upstream/proxy/opt-in fields only.

- [ ] **Step 5: Commit**

```bash
git add src/frontend/src/components/SegmentedControl.tsx \
  src/frontend/src/components/GroupedList.tsx \
  src/frontend/src/pages/Settings.tsx
git commit -m "feat(ui): Settings appearance control and grouped lists"
```

---

### Task 5: Restyle Lint workspace + shared components

**Files:**
- Modify: `src/frontend/src/pages/LintWorkspace.tsx`
- Modify: `src/frontend/src/components/ScoreCard.tsx`
- Modify: `src/frontend/src/components/SectionBreakdown.tsx`
- Modify: `src/frontend/src/components/SuggestionList.tsx`
- Modify: `src/frontend/src/components/DiffView.tsx`

**Interfaces:**
- Consumes: token class vocabulary from Task 1
- Produces: same component props/APIs as today (no behavior changes)

- [ ] **Step 1: Restyle shared components**

`ScoreCard.tsx` — `scoreTone`:

```ts
function scoreTone(score: number): string {
  if (score >= 80) return 'text-success'
  if (score >= 50) return 'text-warning'
  return 'text-danger'
}
```

Section wrapper: `rounded-xl border border-hairline bg-surface p-4`. Labels `text-muted`; body `text-fg`. Drop uppercase tracking on “Prompt score”. Potential savings: `text-accent` (not sky). Issue bullets: `text-warning`.

`SectionBreakdown.tsx` — surface/hairline/fg/muted; bar track `bg-fill`; bar fill `bg-accent`.

`SuggestionList.tsx` — surface + `text-fg` / `text-muted`.

`DiffView.tsx` line tones — first add `--success-soft` to light/dark tokens in `style.css` and `--color-success-soft` in `@theme`, then:

```ts
if (marker === '-') return 'bg-danger-soft text-danger'
if (marker === '+') return 'bg-success-soft text-success'
return 'text-fg'
```

Empty state: dashed `border-hairline`, `text-muted`. Header label sentence case (“Diff”), not uppercase tracking.

- [ ] **Step 2: Restyle `LintWorkspace.tsx`**

Keep layout/hooks. Swap classes to token vocabulary (page title, model select, prompt textarea, Analyze/Apply buttons, error banner, optimize panel, checkboxes). Sentence-case section headers (“Optimize & diff”, “Before”, “After”).

- [ ] **Step 3: Build + smoke**

```bash
cd src/frontend && npm run build
```

Manual: analyze a prompt; score/suggestions/diff readable in light and dark.

- [ ] **Step 4: Commit**

```bash
git add src/frontend/src/pages/LintWorkspace.tsx \
  src/frontend/src/components/ScoreCard.tsx \
  src/frontend/src/components/SectionBreakdown.tsx \
  src/frontend/src/components/SuggestionList.tsx \
  src/frontend/src/components/DiffView.tsx \
  src/frontend/src/style.css
git commit -m "feat(ui): restyle Lint workspace for native theme"
```

---

### Task 6: Restyle Dashboard + final pass

**Files:**
- Modify: `src/frontend/src/pages/Dashboard.tsx`
- Verify: no remaining dark-only palette classes under `src/frontend/src`

**Interfaces:**
- Consumes: token vocabulary
- Produces: unchanged Dashboard behavior

- [ ] **Step 1: Restyle `Dashboard.tsx`**

- Titles/subtitles: `text-fg` / `text-muted`
- Refresh: secondary button classes
- Error: danger soft banner
- `StatCard`: `rounded-xl border border-hairline bg-surface p-4`; label sentence case `text-muted`; value `text-fg`
- Tables/lists: surface + hairline dividers; hover `hover:bg-fill`
- Savings figures: `text-success`; model chips / pct: `text-accent` or `text-muted` as appropriate
- Drop uppercase tracked table headers → `text-xs font-medium text-muted`

- [ ] **Step 2: Grep for leftover dark palette**

```bash
rg -n "sky-|slate-|text-white|white/|black/|rose-950|emerald-950|bg-black" src/frontend/src
```

Expected: no matches in UI files (tests/docs exempt). Fix any leftovers.

- [ ] **Step 3: Full verify**

```bash
cd src/frontend && npm test && npm run build
```

From `src/` if Wails available:

```bash
wails build
# or: wails dev
```

Manual checklist:
1. Fresh look is light; sidebar + logo + wood accent selection
2. Settings Appearance Light/Dark/System; survives restart
3. System follows OS appearance change
4. Lint analyze/apply still works
5. Dashboard refresh + open request still works
6. Save settings still persists non-theme fields

- [ ] **Step 4: Commit**

```bash
git add src/frontend/src/pages/Dashboard.tsx
git commit -m "feat(ui): restyle Dashboard for native theme"
```

---

## Self-review

1. **Spec coverage:** Tokens, ThemeProvider, SQLite `theme`, sidebar+logo, wood accent, Settings Appearance + grouped lists, Lint/Dashboard restyle, Wails bg sync, no frameless — all have tasks.
2. **Placeholders:** None intentional; DiffView soft-success may add `--success-soft` in Task 5 (explicit step).
3. **Type consistency:** `ThemePreference` / `ResolvedTheme` / `SettingKey.Theme` / `useTheme` / `AppPage` names align across tasks.
