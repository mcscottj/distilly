import { FormEvent, useEffect, useState } from 'react'
import { GetSetting } from '../../wailsjs/go/main/App'
import { DiffView } from '../components/DiffView'
import { ScoreCard } from '../components/ScoreCard'
import { SectionBreakdown } from '../components/SectionBreakdown'
import { SuggestionList } from '../components/SuggestionList'
import { ApplyToggles, useAnalyze } from '../hooks/useAnalyze'
import { SettingKey, parseBoolSetting } from '../lib/settings'

const fieldBaseClass =
  'w-full rounded-md border border-hairline bg-fill px-3 py-2 text-fg outline-none focus:border-accent focus:ring-1 focus:ring-accent disabled:cursor-not-allowed disabled:opacity-50'

const fieldClass = `${fieldBaseClass} text-sm`

const editorFieldClass = fieldBaseClass

const primaryButtonClass =
  'rounded-md bg-accent px-4 py-2 text-sm font-medium text-accent-fg hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50'

const secondaryButtonClass =
  'rounded-md border border-hairline bg-fill px-4 py-2 text-sm text-fg hover:bg-surface disabled:cursor-not-allowed disabled:opacity-50'

const checkboxClass =
  'size-4 rounded border-hairline bg-fill text-accent focus:ring-accent'

type LintWorkspaceProps = {
  preferredModel?: string
  onPreferredModelConsumed?: () => void
  preferredPrompt?: string
  onPreferredPromptConsumed?: () => void
}

export function LintWorkspace({
  preferredModel,
  onPreferredModelConsumed,
  preferredPrompt,
  onPreferredPromptConsumed,
}: LintWorkspaceProps) {
  const {
    models,
    analysis,
    applyResult,
    loading,
    applying,
    error,
    analyze,
    apply,
    clearResults,
  } = useAnalyze()

  const [prompt, setPrompt] = useState('')
  const [model, setModel] = useState('')
  const [defaultModel, setDefaultModel] = useState('')
  const [toggles, setToggles] = useState<ApplyToggles>({
    approveNearDuplicates: false,
    approveJsonConversion: false,
  })
  const [prefsLoaded, setPrefsLoaded] = useState(false)

  useEffect(() => {
    let cancelled = false
    Promise.all([
      GetSetting(SettingKey.DefaultModel),
      GetSetting(SettingKey.ApproveNearDuplicates),
      GetSetting(SettingKey.ApproveJsonConversion),
    ])
      .then(([savedModel, near, json]) => {
        if (cancelled) return
        setDefaultModel(savedModel ?? '')
        setToggles({
          approveNearDuplicates: parseBoolSetting(near),
          approveJsonConversion: parseBoolSetting(json),
        })
        setPrefsLoaded(true)
      })
      .catch(() => {
        if (!cancelled) setPrefsLoaded(true)
      })
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (preferredModel) {
      setModel(preferredModel)
      onPreferredModelConsumed?.()
      return
    }
    if (!prefsLoaded || model) return
    if (defaultModel && (models.length === 0 || models.includes(defaultModel))) {
      setModel(defaultModel)
      return
    }
    if (models.length > 0) {
      setModel(models[0])
    }
  }, [preferredModel, onPreferredModelConsumed, prefsLoaded, defaultModel, models, model])

  useEffect(() => {
    if (preferredPrompt) {
      setPrompt(preferredPrompt)
      onPreferredPromptConsumed?.()
    }
  }, [preferredPrompt, onPreferredPromptConsumed])
  async function onAnalyze(e: FormEvent) {
    e.preventDefault()
    if (!prompt.trim()) return
    await analyze(prompt, model)
  }

  async function onApply() {
    if (!prompt.trim()) return
    await apply(prompt, toggles)
  }

  async function onToggleChange(next: ApplyToggles) {
    setToggles(next)
    if (applyResult && prompt.trim()) {
      await apply(prompt, next)
    }
  }

  function useOptimized() {
    if (!applyResult?.optimized) return
    setPrompt(applyResult.optimized)
    clearResults()
  }

  function onClear() {
    setPrompt('')
    clearResults()
  }

  return (
    <div className="mx-auto flex max-w-6xl flex-col gap-6">
      <div>
        <h2 className="text-lg font-medium text-fg">Lint workspace</h2>
        <p className="mt-1 text-sm text-muted">
          Paste a prompt, analyze for savings, then preview and apply optimizations.
        </p>
      </div>

      <form onSubmit={onAnalyze} className="space-y-3">
        <label className="flex min-w-[12rem] flex-col gap-1.5 text-sm sm:max-w-xs">
          <span className="text-muted">Model</span>
          <select
            value={model}
            onChange={(e) => setModel(e.target.value)}
            className={fieldClass}
          >
            {models.length === 0 && !model && <option value="">Loading models…</option>}
            {model && !models.includes(model) && (
              <option value={model}>{model}</option>
            )}
            {models.map((m) => (
              <option key={m} value={m}>
                {m}
              </option>
            ))}
          </select>
        </label>

        <label className="flex flex-col gap-1.5 text-sm">
          <span className="text-muted">Prompt</span>
          <textarea
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            rows={14}
            placeholder="Paste your full prompt here…"
            spellCheck={false}
            className={`${editorFieldClass} font-editor resize-y`}
          />
        </label>

        <div className="flex flex-wrap items-center justify-end gap-2">
          <button
            type="button"
            onClick={onClear}
            disabled={loading || (!prompt && !analysis && !applyResult && !error)}
            className={secondaryButtonClass}
          >
            Clear
          </button>
          <button
            type="submit"
            disabled={loading || !prompt.trim()}
            className={primaryButtonClass}
          >
            {loading ? 'Analyzing…' : 'Analyze'}
          </button>
        </div>
      </form>

      {error && (
        <p className="rounded-md border border-danger bg-danger-soft px-3 py-2 text-sm text-danger">
          {error}
        </p>
      )}

      {analysis && (
        <div className="grid gap-4 lg:grid-cols-2">
          <ScoreCard
            score={analysis.score}
            issues={analysis.issues}
            inputTokens={analysis.inputTokens}
            potentialSavings={analysis.potentialSavings}
            costKnown={analysis.costKnown}
            estimatedCostUsd={analysis.estimatedCostUsd}
            estimatedSavingsUsd={analysis.estimatedSavingsUsd}
          />
          <SectionBreakdown
            sections={
              analysis.sections ?? {
                system: 0,
                examples: 0,
                history: 0,
                question: 0,
              }
            }
          />
          <div className="lg:col-span-2">
            <SuggestionList suggestions={analysis.suggestions} />
          </div>
        </div>
      )}

      <section className="space-y-4 rounded-xl border border-hairline bg-surface p-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h3 className="text-sm font-medium text-fg">Optimize &amp; diff</h3>
            <p className="mt-0.5 text-xs text-muted">
              Exact duplicates always apply. Opt in to near-duplicates and JSON conversion.
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={onApply}
              disabled={applying || !prompt.trim()}
              className={primaryButtonClass}
            >
              {applying ? 'Applying…' : applyResult ? 'Refresh diff' : 'Preview apply'}
            </button>
            {applyResult && (
              <button
                type="button"
                onClick={useOptimized}
                className={secondaryButtonClass}
              >
                Use optimized in editor
              </button>
            )}
          </div>
        </div>

        <div className="flex flex-wrap gap-6 text-sm text-fg">
          <label className="flex cursor-pointer items-center gap-2">
            <input
              type="checkbox"
              checked={toggles.approveNearDuplicates}
              onChange={(e) =>
                onToggleChange({
                  ...toggles,
                  approveNearDuplicates: e.target.checked,
                })
              }
              className={checkboxClass}
            />
            Approve near-duplicates
          </label>
          <label className="flex cursor-pointer items-center gap-2">
            <input
              type="checkbox"
              checked={toggles.approveJsonConversion}
              onChange={(e) =>
                onToggleChange({
                  ...toggles,
                  approveJsonConversion: e.target.checked,
                })
              }
              className={checkboxClass}
            />
            Approve JSON conversion
          </label>
        </div>

        {applyResult && (
          <div className="grid gap-4 lg:grid-cols-2">
            <div className="overflow-hidden rounded-lg border border-hairline bg-fill">
              <div className="border-b border-hairline px-3 py-2 text-xs text-muted">
                Before
              </div>
              <pre className="font-editor max-h-64 overflow-auto whitespace-pre-wrap break-all p-3 text-fg">
                {prompt}
              </pre>
            </div>
            <div className="overflow-hidden rounded-lg border border-hairline bg-fill">
              <div className="border-b border-hairline px-3 py-2 text-xs text-muted">
                After
              </div>
              <pre className="font-editor max-h-64 overflow-auto whitespace-pre-wrap break-all p-3 text-fg">
                {applyResult.optimized}
              </pre>
            </div>
          </div>
        )}

        <DiffView lines={applyResult?.fullDiff} />
      </section>
    </div>
  )
}
