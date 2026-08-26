import { FormEvent, useEffect, useState } from 'react'
import { GetSetting } from '../../wailsjs/go/main/App'
import { useContextSelect } from '../hooks/useContextSelect'
import { SettingKey, withSettingDefault } from '../lib/settings'

const fieldBaseClass =
  'w-full rounded-md border border-hairline bg-fill px-3 py-2 text-fg outline-none focus:border-accent focus:ring-1 focus:ring-accent disabled:cursor-not-allowed disabled:opacity-50'

const fieldClass = `${fieldBaseClass} text-sm`

const primaryButtonClass =
  'rounded-md bg-accent px-4 py-2 text-sm font-medium text-accent-fg hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50'

const secondaryButtonClass =
  'rounded-md border border-hairline bg-fill px-4 py-2 text-sm text-fg hover:bg-surface disabled:cursor-not-allowed disabled:opacity-50'

const checkboxClass =
  'size-4 rounded border-hairline bg-fill text-accent focus:ring-accent'

type ContextWorkspaceProps = {
  active?: boolean
  onOpenInLint?: (markdown: string) => void
}

function formatReasons(reasons: { kind: string; detail: string }[]): string {
  if (!reasons?.length) return ''
  return reasons
    .map((r) => {
      if (r.kind === 'import' || r.kind === 'reverse_import' || r.kind === 'symbol_match') {
        return r.detail ? `${r.kind}: ${r.detail}` : r.kind
      }
      return r.detail || r.kind
    })
    .join('; ')
}

export function ContextWorkspace({ active = true, onOpenInLint }: ContextWorkspaceProps) {
  const { result, loading, error, select, clearResults } = useContextSelect()

  const [repoRoot, setRepoRoot] = useState('')
  const [seedFile, setSeedFile] = useState('')
  const [question, setQuestion] = useState('')
  const [maxDepth, setMaxDepth] = useState('2')
  const [maxTokens, setMaxTokens] = useState('32000')
  const [includeTests, setIncludeTests] = useState(false)
  const [prefsLoaded, setPrefsLoaded] = useState(false)
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (!active) return

    let cancelled = false
    setPrefsLoaded(false)
    Promise.all([
      GetSetting(SettingKey.RepoRoot),
      GetSetting(SettingKey.ContextMaxDepth),
      GetSetting(SettingKey.ContextMaxTokens),
    ])
      .then(([root, depth, tokens]) => {
        if (cancelled) return
        setRepoRoot(root ?? '')
        setMaxDepth(withSettingDefault(SettingKey.ContextMaxDepth, depth))
        setMaxTokens(withSettingDefault(SettingKey.ContextMaxTokens, tokens))
        setPrefsLoaded(true)
      })
      .catch(() => {
        if (!cancelled) setPrefsLoaded(true)
      })
    return () => {
      cancelled = true
    }
  }, [active])

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!repoRoot.trim() || !seedFile.trim()) return

    const depth = parseInt(maxDepth, 10)
    const tokens = parseInt(maxTokens, 10)

    await select({
      repoRoot: repoRoot.trim(),
      seedFile: seedFile.trim(),
      question: question.trim(),
      maxDepth: Number.isFinite(depth) && depth > 0 ? depth : 0,
      maxTokens: Number.isFinite(tokens) && tokens >= 0 ? tokens : 0,
      includeTests,
    })
  }

  async function copyMarkdown() {
    if (!result?.markdown) return
    try {
      await navigator.clipboard.writeText(result.markdown)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 2000)
    } catch {
      window.prompt('Copy markdown context', result.markdown)
    }
  }

  function onClear() {
    setQuestion('')
    clearResults()
  }

  return (
    <div className="mx-auto flex max-w-6xl flex-col gap-6">
      <div>
        <h2 className="text-lg font-medium text-fg">Context workspace</h2>
        <p className="mt-1 text-sm text-muted">
          Select relevant Go files from a repo for LLM context — deterministic import-graph
          analysis, no embeddings.
        </p>
      </div>

      <form onSubmit={(e) => void onSubmit(e)} className="space-y-3">
        <div className="grid gap-3 sm:grid-cols-2">
          <label className="flex flex-col gap-1.5 text-sm sm:col-span-2">
            <span className="text-muted">Repo root</span>
            <input
              type="text"
              value={repoRoot}
              onChange={(e) => setRepoRoot(e.target.value)}
              placeholder="/path/to/your/go/module"
              disabled={!prefsLoaded || loading}
              className={fieldClass}
            />
          </label>

          <label className="flex flex-col gap-1.5 text-sm sm:col-span-2">
            <span className="text-muted">Seed file (repo-relative)</span>
            <input
              type="text"
              value={seedFile}
              onChange={(e) => setSeedFile(e.target.value)}
              placeholder="internal/lint/apply.go"
              disabled={!prefsLoaded || loading}
              className={`${fieldClass} font-mono`}
            />
          </label>

          <label className="flex flex-col gap-1.5 text-sm">
            <span className="text-muted">Max depth</span>
            <input
              type="number"
              min={1}
              value={maxDepth}
              onChange={(e) => setMaxDepth(e.target.value)}
              disabled={!prefsLoaded || loading}
              className={fieldClass}
            />
          </label>

          <label className="flex flex-col gap-1.5 text-sm">
            <span className="text-muted">Max tokens</span>
            <input
              type="number"
              min={0}
              value={maxTokens}
              onChange={(e) => setMaxTokens(e.target.value)}
              disabled={!prefsLoaded || loading}
              className={fieldClass}
            />
          </label>
        </div>

        <label className="flex flex-col gap-1.5 text-sm">
          <span className="text-muted">Question</span>
          <textarea
            value={question}
            onChange={(e) => setQuestion(e.target.value)}
            rows={3}
            placeholder="What are you trying to understand or change?"
            disabled={!prefsLoaded || loading}
            className={fieldClass}
          />
        </label>

        <label className="flex cursor-pointer items-center gap-2 text-sm text-fg">
          <input
            type="checkbox"
            checked={includeTests}
            onChange={(e) => setIncludeTests(e.target.checked)}
            disabled={!prefsLoaded || loading}
            className={checkboxClass}
          />
          Include test files (*_test.go)
        </label>

        <div className="flex flex-wrap items-center justify-end gap-2">
          <button
            type="button"
            onClick={onClear}
            disabled={loading || (!question && !result && !error)}
            className={secondaryButtonClass}
          >
            Clear
          </button>
          <button
            type="submit"
            disabled={loading || !repoRoot.trim() || !seedFile.trim()}
            className={primaryButtonClass}
          >
            {loading ? 'Selecting…' : 'Select context'}
          </button>
        </div>
      </form>

      {error && (
        <p className="rounded-md border border-danger bg-danger-soft px-3 py-2 text-sm text-danger">
          {error}
        </p>
      )}

      {result && (
        <section className="space-y-4 rounded-xl border border-hairline bg-surface p-4">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h3 className="text-sm font-medium text-fg">
                Selected files ({result.files?.length ?? 0})
              </h3>
              <p className="mt-0.5 text-xs text-muted">
                {result.totalTokens?.toLocaleString()} tokens total
                {result.excludedCount > 0 &&
                  ` · ${result.excludedCount} excluded due to budget`}
              </p>
            </div>
            <div className="flex flex-wrap gap-2">
              {result.markdown && (
                <button type="button" onClick={() => void copyMarkdown()} className={secondaryButtonClass}>
                  {copied ? 'Copied' : 'Copy markdown'}
                </button>
              )}
              {result.markdown && onOpenInLint && (
                <button
                  type="button"
                  onClick={() => onOpenInLint(result.markdown)}
                  className={secondaryButtonClass}
                >
                  Open in Lint
                </button>
              )}
            </div>
          </div>

          {result.warnings?.length > 0 && (
            <ul className="list-inside list-disc text-sm text-muted">
              {result.warnings.map((w) => (
                <li key={w}>{w}</li>
              ))}
            </ul>
          )}

          <div className="overflow-x-auto rounded-lg border border-hairline">
            <table className="w-full min-w-[32rem] text-left text-sm">
              <thead className="border-b border-hairline bg-fill text-xs text-muted">
                <tr>
                  <th className="px-3 py-2 font-medium">Path</th>
                  <th className="px-3 py-2 font-medium text-right">Tokens</th>
                  <th className="px-3 py-2 font-medium">Reasons</th>
                </tr>
              </thead>
              <tbody>
                {(result.files ?? []).map((f) => (
                  <tr key={f.path} className="border-b border-hairline last:border-0">
                    <td className="px-3 py-2 font-mono text-fg">{f.path}</td>
                    <td className="px-3 py-2 text-right tabular-nums text-fg">
                      {f.tokens?.toLocaleString()}
                    </td>
                    <td className="px-3 py-2 text-muted">{formatReasons(f.reasons ?? [])}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {result.markdown && (
            <div className="overflow-hidden rounded-lg border border-hairline bg-fill">
              <div className="border-b border-hairline px-3 py-2 text-xs text-muted">
                Markdown preview
              </div>
              <pre className="font-editor max-h-96 overflow-auto whitespace-pre-wrap break-all p-3 text-fg">
                {result.markdown}
              </pre>
            </div>
          )}
        </section>
      )}
    </div>
  )
}
