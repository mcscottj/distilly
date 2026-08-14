# High Contrast Theme Toggle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a High contrast Appearance toggle that intensifies the active Light/Dark theme via `data-contrast` and a persisted `theme_contrast` setting.

**Architecture:** Extend existing theme helpers and `ThemeProvider` with a separate contrast preference (`normal` | `high`). Set `document.documentElement.dataset.contrast` independently of `data-theme`. Override semantic CSS tokens and border/focus chrome under `[data-contrast="high"]`. Settings Appearance gains an immediate-persist checkbox row.

**Tech Stack:** React 19, TypeScript, Tailwind CSS v4 tokens in `style.css`, Vitest, Wails `GetSetting` / `SetSetting` (SQLite).

**Spec:** [`docs/superpowers/specs/2026-08-14-high-contrast-theme-design.md`](../specs/2026-08-14-high-contrast-theme-design.md)

## Global Constraints

- UX: separate High contrast toggle; Light / Dark / System unchanged
- Strength: color + chrome only (no larger type or padding)
- Mechanism: `data-contrast` overlay on existing `data-theme`
- Preference values: `normal` | `high`; default `normal`
- SQLite key: `theme_contrast` (immediate `SetSetting`, like theme)
- Do not auto-follow OS `prefers-contrast`
- Keep wood accent; nudge only if button contrast fails
- Branch: `feat/high-contrast-theme` (already created — do not implement on `main`)

---

## File map

| File | Responsibility |
|------|----------------|
| `src/frontend/src/lib/theme.ts` | Add `ThemeContrast` type + `parseThemeContrast` |
| `src/frontend/src/lib/theme.test.ts` | Vitest for contrast parse |
| `src/frontend/src/lib/settings.ts` | `SettingKey.ThemeContrast` + default `normal` |
| `src/frontend/src/style.css` | `[data-contrast="high"]` token + chrome overrides |
| `src/frontend/src/hooks/useTheme.tsx` | Load/persist contrast; set `data-contrast` |
| `src/frontend/src/pages/Settings.tsx` | Appearance High contrast checkbox row |

---

### Task 1: Contrast helpers + SettingKey

**Files:**
- Modify: `src/frontend/src/lib/theme.ts`
- Modify: `src/frontend/src/lib/theme.test.ts`
- Modify: `src/frontend/src/lib/settings.ts`

**Interfaces:**
- Consumes: existing `parseThemePreference` / `resolveTheme` / `windowBackgroundRGBA` (unchanged)
- Produces:
  - `export type ThemeContrast = 'normal' | 'high'`
  - `export function parseThemeContrast(value: string | null | undefined): ThemeContrast`
  - `SettingKey.ThemeContrast = 'theme_contrast'`
  - `SETTING_DEFAULTS[SettingKey.ThemeContrast] = 'normal'`

- [ ] **Step 1: Write the failing tests**

Append to `src/frontend/src/lib/theme.test.ts`:

```ts
import {
  parseThemePreference,
  parseThemeContrast,
  resolveTheme,
  windowBackgroundRGBA,
} from './theme'

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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd src/frontend && npm test -- src/lib/theme.test.ts`

Expected: FAIL — `parseThemeContrast` is not exported / not defined.

- [ ] **Step 3: Implement helpers and SettingKey**

In `src/frontend/src/lib/theme.ts`, add:

```ts
export type ThemeContrast = 'normal' | 'high'

export function parseThemeContrast(value: string | null | undefined): ThemeContrast {
  if (value === 'normal' || value === 'high') {
    return value
  }
  return 'normal'
}
```

In `src/frontend/src/lib/settings.ts`, add to `SettingKey`:

```ts
ThemeContrast: 'theme_contrast',
```

And to `SETTING_DEFAULTS`:

```ts
[SettingKey.ThemeContrast]: 'normal',
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd src/frontend && npm test -- src/lib/theme.test.ts`

Expected: PASS (all theme tests green).

- [ ] **Step 5: Commit**

```bash
git add src/frontend/src/lib/theme.ts src/frontend/src/lib/theme.test.ts src/frontend/src/lib/settings.ts
git commit -m "feat(ui): add theme contrast preference helpers"
```

---

### Task 2: High-contrast CSS tokens and chrome

**Files:**
- Modify: `src/frontend/src/style.css`

**Interfaces:**
- Consumes: existing `:root` / `[data-theme="dark"]` tokens
- Produces: `[data-contrast="high"]` and `[data-theme="dark"][data-contrast="high"]` overrides; `--border-width` used for thicker hairlines under HC

- [ ] **Step 1: Add border-width base token**

In `:root` (after existing tokens), add:

```css
--border-width: 1px;
```

In `@theme`, add:

```css
--border-width-hairline: var(--border-width);
```

(If Tailwind v4 `@theme` does not map border-width cleanly for utilities already using `border`, prefer the global rule in Step 3 that sets `border-width` under high contrast for elements that already have a border class — do not invent new utility names across every component.)

- [ ] **Step 2: Add high-contrast token overrides**

Append to `src/frontend/src/style.css`:

```css
[data-contrast="high"] {
  --fg: #000000;
  --muted: #3a3a3c;
  --border: rgba(0, 0, 0, 0.45);
  --border-width: 2px;
  --accent-soft: rgba(139, 94, 60, 0.28);
  --danger-soft: rgba(192, 57, 43, 0.22);
  --success-soft: rgba(31, 122, 76, 0.22);
}

[data-theme="dark"][data-contrast="high"] {
  --fg: #ffffff;
  --muted: #d1d1d6;
  --border: rgba(255, 255, 255, 0.45);
  --border-width: 2px;
  --accent-soft: rgba(196, 164, 132, 0.36);
  --danger-soft: rgba(255, 107, 107, 0.28);
  --success-soft: rgba(107, 207, 142, 0.28);
}
```

Do not change `--bg-window` / `--bg-sidebar` / `--bg-surface` / `--bg-fill` unless a later manual check shows soft surfaces fail readability; default is leave canvases stable per spec.

- [ ] **Step 3: Add chrome rules for border thickness and focus**

Append:

```css
[data-contrast="high"] .border-hairline,
[data-contrast="high"] [class*="border-hairline"] {
  border-width: var(--border-width);
}

[data-contrast="high"] :focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

[data-contrast="high"] :focus:not(:focus-visible) {
  outline: none;
}
```

Notes:
- Tailwind’s `border` utility sets width 1px; the override above thickens hairline borders under HC.
- Focus outline supplements existing `focus:ring-1` fields without editing every page class string.
- If `[class*="border-hairline"]` is too broad in practice, narrow to `.border-hairline` only (Tailwind v4 emits that class name).

- [ ] **Step 4: Manual CSS smoke (optional in browser preview)**

Run: `cd src/frontend && npm run dev`

In DevTools on `<html>`: set `data-theme="light"` then `data-contrast="high"`; confirm darker muted text and thicker borders. Repeat with `data-theme="dark"`. Remove `data-contrast` and confirm restore.

- [ ] **Step 5: Commit**

```bash
git add src/frontend/src/style.css
git commit -m "feat(ui): add high-contrast CSS token and chrome overrides"
```

---

### Task 3: ThemeProvider contrast load / persist / apply

**Files:**
- Modify: `src/frontend/src/hooks/useTheme.tsx`

**Interfaces:**
- Consumes:
  - `parseThemeContrast` from `../lib/theme`
  - `SettingKey.ThemeContrast` from `../lib/settings`
  - `GetSetting` / `SetSetting`
- Produces (extend `ThemeContextValue`):
  - `contrast: ThemeContrast`
  - `setContrast: (contrast: ThemeContrast) => Promise<void>`
  - existing `preference`, `resolved`, `setPreference`, `loading` unchanged in meaning

- [ ] **Step 1: Extend context type and apply helper**

Update imports:

```ts
import {
  parseThemePreference,
  parseThemeContrast,
  resolveTheme,
  windowBackgroundRGBA,
  type ResolvedTheme,
  type ThemeContrast,
  type ThemePreference,
} from '../lib/theme'
```

Extend context:

```ts
type ThemeContextValue = {
  preference: ThemePreference
  resolved: ResolvedTheme
  contrast: ThemeContrast
  setPreference: (preference: ThemePreference) => Promise<void>
  setContrast: (contrast: ThemeContrast) => Promise<void>
  loading: boolean
}
```

Update apply function so theme + contrast are both applied:

```ts
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
```

- [ ] **Step 2: Load and persist contrast in ThemeProvider**

Add state:

```ts
const [contrast, setContrastState] = useState<ThemeContrast>('normal')
```

In the existing load `useEffect`, also load contrast (same cancelled flag). Prefer loading both before clearing `loading`:

```ts
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
```

Apply on change:

```ts
useEffect(() => {
  applyResolved(resolved, contrast)
}, [resolved, contrast])
```

Add `setContrast`:

```ts
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
```

Include `contrast` and `setContrast` in the context `value` `useMemo` deps.

- [ ] **Step 3: Typecheck**

Run: `cd src/frontend && npx tsc --noEmit`

Expected: PASS (or only pre-existing unrelated errors). `useTheme()` consumers that destructure only theme fields still type-check; Settings will use the new fields in Task 4.

- [ ] **Step 4: Commit**

```bash
git add src/frontend/src/hooks/useTheme.tsx
git commit -m "feat(ui): persist and apply theme contrast on document"
```

---

### Task 4: Settings Appearance High contrast control

**Files:**
- Modify: `src/frontend/src/pages/Settings.tsx`

**Interfaces:**
- Consumes: `useTheme()` → `contrast`, `setContrast`, `loading` (themeLoading)
- Produces: Appearance row that sets `high` / `normal` immediately

- [ ] **Step 1: Wire useTheme contrast**

Change:

```ts
const { preference, setPreference, loading: themeLoading } = useTheme()
```

to:

```ts
const {
  preference,
  setPreference,
  contrast,
  setContrast,
  loading: themeLoading,
} = useTheme()
```

- [ ] **Step 2: Add High contrast row under Theme**

Inside the Appearance `GroupedList`, immediately after the Theme `GroupedRow`, add:

```tsx
<GroupedRow
  label="High contrast"
  description="Stronger text, borders, and focus rings for low vision."
>
  <input
    type="checkbox"
    checked={contrast === 'high'}
    disabled={themeLoading}
    onChange={(e) => {
      void setContrast(e.target.checked ? 'high' : 'normal').catch((err) => {
        setError(err instanceof Error ? err.message : 'Failed to save contrast preference')
      })
    }}
    className={checkboxClass}
  />
</GroupedRow>
```

Do not put this behind the form Save button.

- [ ] **Step 3: Manual verification checklist**

Run the desktop app (or `npm run dev` in `src/frontend` if preview is enough for DOM attributes):

1. Settings → Appearance → enable High contrast (light theme): near-black text, darker muted, thicker borders, visible focus outline on tab.
2. Switch Theme to Dark with High contrast still on: near-white text / strong borders remain.
3. Disable High contrast: restore previous soft tokens.
4. Restart app: High contrast preference restored from SQLite.
5. Light / Dark / System still work with contrast off.

- [ ] **Step 4: Run unit tests**

Run: `cd src/frontend && npm test`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add src/frontend/src/pages/Settings.tsx
git commit -m "feat(ui): add High contrast toggle in Appearance settings"
```

---

## Self-review (plan vs spec)

| Spec requirement | Task |
|------------------|------|
| Separate toggle; keep Light/Dark/System | Task 4 |
| Color + chrome (not type/padding) | Task 2 |
| `data-contrast` overlay | Task 3 |
| `normal` \| `high`, default `normal` | Task 1 |
| SQLite `theme_contrast`, immediate save | Tasks 1, 3, 4 |
| No OS `prefers-contrast` auto | Not implemented (correct) |
| Token overrides + thicker borders + stronger focus | Task 2 |
| Unit tests for parse | Task 1 |
| Manual light/dark + persistence | Task 4 |

No placeholders remaining. Types consistent: `ThemeContrast`, `parseThemeContrast`, `setContrast`, `SettingKey.ThemeContrast`.
