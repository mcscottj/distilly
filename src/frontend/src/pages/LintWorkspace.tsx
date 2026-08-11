import { FormEvent, useEffect, useState } from 'react'
import { DiffView } from '../components/DiffView'
import { ScoreCard } from '../components/ScoreCard'
import { SectionBreakdown } from '../components/SectionBreakdown'
import { SuggestionList } from '../components/SuggestionList'
import { ApplyToggles, useAnalyze } from '../hooks/useAnalyze'

export function LintWorkspace() {
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
  const [toggles, setToggles] = useState<ApplyToggles>({
    approveNearDuplicates: false,
    approveJsonConversion: false,
  })

  useEffect(() => {
    if (!model && models.length > 0) {
      setModel(models[0])
    }
  }, [models, model])

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

  return (
    <div className="mx-auto flex max-w-6xl flex-col gap-6">
      <div>
        <h2 className="text-lg font-medium text-white">Lint workspace</h2>
        <p className="mt-1 text-sm text-slate-400">
          Paste a prompt, analyze for savings, then preview and apply optimizations.
        </p>
      </div>

      <form onSubmit={onAnalyze} className="space-y-3">
        <div className="flex flex-wrap items-end gap-3">
          <label className="flex min-w-[12rem] flex-1 flex-col gap-1.5 text-sm">
            <span className="text-slate-400">Model</span>
            <select
              value={model}
              onChange={(e) => setModel(e.target.value)}
              className="rounded-md border border-white/15 bg-black/25 px-3 py-2 text-sm text-white outline-none focus:border-sky-400"
            >
              {models.length === 0 && <option value="">Loading models…</option>}
              {models.map((m) => (
                <option key={m} value={m}>
                  {m}
                </option>
              ))}
            </select>
          </label>
          <button
            type="submit"
            disabled={loading || !prompt.trim()}
            className="rounded-md bg-sky-500 px-4 py-2 text-sm font-medium text-slate-950 hover:bg-sky-400 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {loading ? 'Analyzing…' : 'Analyze'}
          </button>
        </div>

        <label className="flex flex-col gap-1.5 text-sm">
          <span className="text-slate-400">Prompt</span>
          <textarea
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            rows={14}
            placeholder="Paste your full prompt here…"
            spellCheck={false}
            className="w-full resize-y rounded-md border border-white/15 bg-black/25 px-3 py-2 font-mono text-sm leading-relaxed text-slate-100 outline-none focus:border-sky-400"
          />
        </label>
      </form>

      {error && (
        <p className="rounded-md border border-rose-500/40 bg-rose-950/40 px-3 py-2 text-sm text-rose-200">
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

      <section className="space-y-4 rounded-lg border border-white/10 bg-black/15 p-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h3 className="text-sm font-medium text-white">Optimize &amp; diff</h3>
            <p className="mt-0.5 text-xs text-slate-400">
              Exact duplicates always apply. Opt in to near-duplicates and JSON conversion.
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={onApply}
              disabled={applying || !prompt.trim()}
              className="rounded-md bg-sky-500 px-4 py-2 text-sm font-medium text-slate-950 hover:bg-sky-400 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {applying ? 'Applying…' : applyResult ? 'Refresh diff' : 'Preview apply'}
            </button>
            {applyResult && (
              <button
                type="button"
                onClick={useOptimized}
                className="rounded-md border border-white/20 px-4 py-2 text-sm text-slate-200 hover:bg-white/5"
              >
                Use optimized in editor
              </button>
            )}
          </div>
        </div>

        <div className="flex flex-wrap gap-6 text-sm text-slate-300">
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
              className="size-4 rounded border-white/20 bg-black/40 text-sky-500 focus:ring-sky-400"
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
              className="size-4 rounded border-white/20 bg-black/40 text-sky-500 focus:ring-sky-400"
            />
            Approve JSON conversion
          </label>
        </div>

        {applyResult && (
          <div className="grid gap-4 lg:grid-cols-2">
            <div className="overflow-hidden rounded-lg border border-white/10 bg-black/30">
              <div className="border-b border-white/10 px-3 py-2 text-xs uppercase tracking-[0.12em] text-slate-400">
                Before
              </div>
              <pre className="max-h-64 overflow-auto whitespace-pre-wrap break-all p-3 font-mono text-xs leading-5 text-slate-300">
                {prompt}
              </pre>
            </div>
            <div className="overflow-hidden rounded-lg border border-white/10 bg-black/30">
              <div className="border-b border-white/10 px-3 py-2 text-xs uppercase tracking-[0.12em] text-slate-400">
                After
              </div>
              <pre className="max-h-64 overflow-auto whitespace-pre-wrap break-all p-3 font-mono text-xs leading-5 text-slate-300">
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
