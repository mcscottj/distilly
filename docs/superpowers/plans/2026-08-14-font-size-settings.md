# Font Size Settings Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Appearance controls for stepped interface text scale (Default / Large / Extra large) and independent Lint editor font size (12 / 14 / 16 / 18 px).

**Architecture:** Extend theme helpers and `ThemeProvider` with `ui_font_scale` and `editor_font_size` preferences. Apply `data-ui-scale` and `--editor-font-size` on `<html>`. Remap Tailwind `--text-*` tokens under scale attributes (no rem-root zoom). Lint monospace surfaces use a shared `.font-editor` class.

**Tech Stack:** React 19, TypeScript, Tailwind CSS v4, Vitest, Wails `GetSetting` / `SetSetting` (SQLite).

**Spec:** [`docs/superpowers/specs/2026-08-14-font-size-settings-design.md`](../specs/2026-08-14-font-size-settings-design.md)

## Global Constraints

- Interface control: stepped Default / Large / Extra large (`default` | `large` | `xlarge`)
- Editor control: explicit px 12 / 14 / 16 / 18
- Mechanism: CSS variables + `data-ui-scale` on `<html>` (same Appearance persistence pattern as theme / high contrast)
- SQLite keys: `ui_font_scale`, `editor_font_size`
- Defaults: `ui_font_scale=default`; `editor_font_size=14`
- Immediate `SetSetting` (not behind Settings form Save)
- Editor scope: Lint only (prompt, before/after, DiffView, other Lint mono) — not Settings mono
- Do not change `html` rem root solely to scale UI
- No OS text-size auto-follow
- Branch: `feat/font-size-settings` (already created — do not implement on `main`)

---

## File map

| File | Responsibility |
|------|----------------|
| `src/frontend/src/lib/theme.ts` | `UiFontScale`, `EditorFontSize`, parsers, `editorFontSizePx` |
| `src/frontend/src/lib/theme.test.ts` | Vitest for new parsers |
| `src/frontend/src/lib/settings.ts` | Setting keys + defaults |
| `src/frontend/src/style.css` | `--editor-font-size`, `.font-editor`, `[data-ui-scale]` text tokens |
| `src/frontend/src/hooks/useTheme.tsx` | Load/persist/apply scale + editor size |
| `src/frontend/src/pages/Settings.tsx` | Appearance controls |
| `src/frontend/src/pages/LintWorkspace.tsx` | Apply `font-editor` to Lint mono |
| `src/frontend/src/components/DiffView.tsx` | Apply `font-editor` to diff body |

---

### Task 1: Font-size helpers + SettingKey

**Files:**
- Modify: `src/frontend/src/lib/theme.ts`
- Modify: `src/frontend/src/lib/theme.test.ts`
- Modify: `src/frontend/src/lib/settings.ts`

**Interfaces:**
- Consumes: existing theme/contrast helpers (unchanged)
- Produces:
  - `export type UiFontScale = 'default' | 'large' | 'xlarge'`
  - `export type EditorFontSize = '12' | '14' | '16' | '18'`
  - `export function parseUiFontScale(value: string | null | undefined): UiFontScale`
  - `export function parseEditorFontSize(value: string | null | undefined): EditorFontSize`
  - `export function editorFontSizePx(size: EditorFontSize): string` → e.g. `'14px'`
  - `SettingKey.UiFontScale = 'ui_font_scale'`
  - `SettingKey.EditorFontSize = 'editor_font_size'`
  - Defaults: `'default'` and `'14'`

- [ ] **Step 1: Write the failing tests**

Append to `src/frontend/src/lib/theme.test.ts` (keep existing imports; add the new symbols):

```ts
import {
  parseThemePreference,
  parseThemeContrast,
  parseUiFontScale,
  parseEditorFontSize,
  editorFontSizePx,
  resolveTheme,
  windowBackgroundRGBA,
} from './theme'

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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd src/frontend && npm test -- src/lib/theme.test.ts`

Expected: FAIL — new parsers / `editorFontSizePx` not defined.

- [ ] **Step 3: Implement helpers and SettingKey**

In `src/frontend/src/lib/theme.ts`, add:

```ts
export type UiFontScale = 'default' | 'large' | 'xlarge'
export type EditorFontSize = '12' | '14' | '16' | '18'

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
```

In `src/frontend/src/lib/settings.ts`, add to `SettingKey`:

```ts
UiFontScale: 'ui_font_scale',
EditorFontSize: 'editor_font_size',
```

And to `SETTING_DEFAULTS`:

```ts
[SettingKey.UiFontScale]: 'default',
[SettingKey.EditorFontSize]: '14',
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd src/frontend && npm test -- src/lib/theme.test.ts`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add src/frontend/src/lib/theme.ts src/frontend/src/lib/theme.test.ts src/frontend/src/lib/settings.ts
git commit -m "feat(ui): add ui font scale and editor font size helpers"
```

---

### Task 2: CSS for UI scale tokens and `.font-editor`

**Files:**
- Modify: `src/frontend/src/style.css`

**Interfaces:**
- Consumes: none from TS
- Produces: `--editor-font-size` default; `.font-editor`; `[data-ui-scale="large|xlarge"]` remaps of Tailwind text size theme vars

- [ ] **Step 1: Add editor font default and utility**

In `:root`, after existing tokens, add:

```css
--editor-font-size: 14px;
```

After the `body { ... }` block, add:

```css
.font-editor {
  font-family: var(--font-mono);
  font-size: var(--editor-font-size);
  line-height: 1.5;
}
```

- [ ] **Step 2: Remap Tailwind text tokens under UI scale**

Append (exact values — tune only if visual check requires a small nudge):

```css
/* Interface text scale — override Tailwind v4 --text-* theme vars (no rem-root zoom). */
[data-ui-scale="large"] {
  --text-xs: 0.875rem; /* 14px */
  --text-sm: 1rem; /* 16px */
  --text-base: 1.125rem;
  --text-lg: 1.25rem;
  --text-xl: 1.375rem;
  --text-2xl: 1.6875rem;
  --text-4xl: 2.5rem;
}

[data-ui-scale="xlarge"] {
  --text-xs: 0.9375rem; /* 15px */
  --text-sm: 1.125rem; /* 18px */
  --text-base: 1.25rem;
  --text-lg: 1.375rem;
  --text-xl: 1.5rem;
  --text-2xl: 1.875rem;
  --text-4xl: 2.75rem;
}
```

Do **not** set `html { font-size: ... }` for scaling.

Note: `.font-editor` uses `--editor-font-size` in px and is independent of these remaps. Elements that keep `text-xs` / `font-mono` without `font-editor` will still follow UI scale — Task 4 replaces Lint mono classes with `font-editor`.

- [ ] **Step 3: Build smoke**

Run: `cd src/frontend && npm run build`

Expected: PASS. Optionally confirm compiled CSS contains `[data-ui-scale="large"]` and `.font-editor`.

- [ ] **Step 4: Commit**

```bash
git add src/frontend/src/style.css
git commit -m "feat(ui): add UI text scale tokens and font-editor utility"
```

---

### Task 3: ThemeProvider load / persist / apply font prefs

**Files:**
- Modify: `src/frontend/src/hooks/useTheme.tsx`

**Interfaces:**
- Consumes:
  - `parseUiFontScale`, `parseEditorFontSize`, `editorFontSizePx`
  - `SettingKey.UiFontScale`, `SettingKey.EditorFontSize`
- Produces (extend `ThemeContextValue`):
  - `uiFontScale: UiFontScale`
  - `editorFontSize: EditorFontSize`
  - `setUiFontScale: (scale: UiFontScale) => Promise<void>`
  - `setEditorFontSize: (size: EditorFontSize) => Promise<void>`

- [ ] **Step 1: Extend imports, context type, and apply helper**

Update imports from `../lib/theme` to include:

```ts
parseUiFontScale,
parseEditorFontSize,
editorFontSizePx,
type UiFontScale,
type EditorFontSize,
```

Extend context:

```ts
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
```

Replace `applyResolved` with:

```ts
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
```

- [ ] **Step 2: State, load, setters, apply effect**

Add state:

```ts
const [uiFontScale, setUiFontScaleState] = useState<UiFontScale>('default')
const [editorFontSize, setEditorFontSizeState] = useState<EditorFontSize>('14')
```

Replace the load `Promise.all` with four settings:

```ts
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
```

Apply effect:

```ts
useEffect(() => {
  applyResolved(resolved, contrast, uiFontScale, editorFontSize)
}, [resolved, contrast, uiFontScale, editorFontSize])
```

Add setters mirroring `setContrast`:

```ts
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
```

Include the new fields in the context `value` `useMemo`.

- [ ] **Step 3: Typecheck + unit tests**

Run: `cd src/frontend && npx tsc --noEmit && npm test`

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add src/frontend/src/hooks/useTheme.tsx
git commit -m "feat(ui): persist and apply UI font scale and editor size"
```

---

### Task 4: Settings Appearance controls + Lint `font-editor`

**Files:**
- Modify: `src/frontend/src/pages/Settings.tsx`
- Modify: `src/frontend/src/pages/LintWorkspace.tsx`
- Modify: `src/frontend/src/components/DiffView.tsx`

**Interfaces:**
- Consumes: `uiFontScale`, `setUiFontScale`, `editorFontSize`, `setEditorFontSize`, `loading` from `useTheme()`
- Produces: Appearance rows; Lint mono uses `.font-editor`

- [ ] **Step 1: Wire Settings Appearance rows**

In `Settings.tsx`, extend the `useTheme()` destructure:

```ts
const {
  preference,
  setPreference,
  contrast,
  setContrast,
  uiFontScale,
  setUiFontScale,
  editorFontSize,
  setEditorFontSize,
  loading: themeLoading,
} = useTheme()
```

After the High contrast `GroupedRow`, add:

```tsx
<GroupedRow
  label="Interface text"
  description="Enlarge labels, buttons, and chrome across the app."
>
  <SegmentedControl
    value={uiFontScale}
    disabled={themeLoading}
    onChange={(v) => {
      void setUiFontScale(v).catch((err) => {
        setError(err instanceof Error ? err.message : 'Failed to save interface text size')
      })
    }}
    options={[
      { value: 'default', label: 'Default' },
      { value: 'large', label: 'Large' },
      { value: 'xlarge', label: 'Extra large' },
    ]}
  />
</GroupedRow>
<GroupedRow
  label="Editor text"
  description="Font size for Lint prompt, diffs, and other monospace panels."
>
  <SegmentedControl
    value={editorFontSize}
    disabled={themeLoading}
    onChange={(v) => {
      void setEditorFontSize(v).catch((err) => {
        setError(err instanceof Error ? err.message : 'Failed to save editor text size')
      })
    }}
    options={[
      { value: '12', label: '12' },
      { value: '14', label: '14' },
      { value: '16', label: '16' },
      { value: '18', label: '18' },
    ]}
  />
</GroupedRow>
```

Do not place these inside the Save form.

- [ ] **Step 2: Apply `font-editor` in LintWorkspace**

In `src/frontend/src/pages/LintWorkspace.tsx`:

1. Prompt textarea — change class from  
   `` `${fieldClass} resize-y font-mono leading-relaxed` ``  
   to  
   `` `${fieldClass} font-editor resize-y leading-relaxed` ``  
   (`fieldClass` still includes `text-sm`; `font-editor`’s `font-size` must win — see Step 2b).

2. Before/after `<pre>` bodies — replace  
   `font-mono text-xs leading-5`  
   with  
   `font-editor`  
   (drop `text-xs` / `font-mono` so editor size is not fighting UI scale).

Keep panel header captions (`text-xs text-muted`) on UI scale — they are chrome, not editor bodies.

- [ ] **Step 2b: Ensure `.font-editor` wins over `text-sm` on the textarea**

If build order lets Tailwind `text-sm` override `.font-editor`, strengthen the utility in `style.css`:

```css
.font-editor {
  font-family: var(--font-mono);
  font-size: var(--editor-font-size) !important;
  line-height: 1.5;
}
```

Prefer avoiding `!important` if placing `.font-editor` after Tailwind utilities (unlayered custom CSS already beats `@layer utilities` in this project). Verify with build/DevTools; only add `!important` if needed.

Alternatively for the textarea only: remove `text-sm` from a local class string used by the prompt, e.g. define `editorFieldClass` without `text-sm` and use that for the textarea while keeping `fieldClass` for the model select. Prefer that if it avoids `!important`:

```ts
const editorFieldClass =
  'w-full rounded-md border border-hairline bg-fill px-3 py-2 text-fg outline-none focus:border-accent focus:ring-1 focus:ring-accent disabled:cursor-not-allowed disabled:opacity-50'
```

Prompt: `` className={`${editorFieldClass} font-editor resize-y leading-relaxed`} ``

- [ ] **Step 3: Apply `font-editor` in DiffView**

In `src/frontend/src/components/DiffView.tsx`, change the `<pre>` class from:

```tsx
<pre className="max-h-[28rem] overflow-auto p-0 font-mono text-xs leading-5">
```

to:

```tsx
<pre className="font-editor max-h-[28rem] overflow-auto p-0">
```

Keep the empty-state message on `text-sm` (UI chrome). Keep the “Diff” header on `text-xs` (UI chrome).

- [ ] **Step 4: Verify**

Run: `cd src/frontend && npm test && npx tsc --noEmit && npm run build`

Manual checklist (desktop or `npm run dev`):

1. Appearance → Interface text Large / Extra large: labels/buttons grow; Lint mono unchanged when Editor text fixed.
2. Editor text 12 → 18: prompt, before/after, DiffView change; Settings API key / base URL do not.
3. Restart: both prefs persist.
4. High contrast on/off still works.

- [ ] **Step 5: Commit**

```bash
git add src/frontend/src/pages/Settings.tsx src/frontend/src/pages/LintWorkspace.tsx src/frontend/src/components/DiffView.tsx src/frontend/src/style.css
git commit -m "feat(ui): add font size Appearance controls and Lint editor sizing"
```

---

## Self-review (plan vs spec)

| Spec requirement | Task |
|------------------|------|
| Stepped Default / Large / Extra large | Tasks 1–4 |
| Editor px 12/14/16/18 | Tasks 1–4 |
| `data-ui-scale` + `--editor-font-size` | Tasks 2–3 |
| Keys `ui_font_scale`, `editor_font_size`; defaults | Task 1 |
| Immediate SetSetting | Tasks 3–4 |
| Lint-only editor scope; Settings mono excluded | Task 4 |
| No rem-root zoom | Task 2 |
| Unit parse tests | Task 1 |
| Manual verification | Task 4 |

No placeholders. Types consistent: `UiFontScale`, `EditorFontSize`, `parseUiFontScale`, `parseEditorFontSize`, `editorFontSizePx`, `setUiFontScale`, `setEditorFontSize`.
