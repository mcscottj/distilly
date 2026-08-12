import { store } from '../../wailsjs/go/models'
import { useDashboard } from '../hooks/useDashboard'
import { formatPct, formatTokens, formatUsd, formatWhen } from '../lib/format'

type DashboardProps = {
  onOpenRequest?: (request: store.Request) => void
}

export function Dashboard({ onOpenRequest }: DashboardProps) {
  const { stats, recent, loading, error, refresh } = useDashboard()

  const byModel = stats?.byModel ?? []

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-medium text-white">Dashboard</h2>
          <p className="mt-1 text-sm text-slate-400">
            Session and historical savings from logged manual and proxy requests.
          </p>
        </div>
        <button
          type="button"
          onClick={() => void refresh()}
          disabled={loading}
          className="rounded-md border border-white/20 px-3 py-1.5 text-sm text-slate-200 hover:bg-white/5 disabled:opacity-50"
        >
          {loading ? 'Refreshing…' : 'Refresh'}
        </button>
      </div>

      {error && (
        <p className="rounded-md border border-rose-500/40 bg-rose-950/40 px-3 py-2 text-sm text-rose-200">
          {error}
        </p>
      )}

      <div className="grid gap-3 sm:grid-cols-3">
        <StatCard
          label="Requests analyzed"
          value={loading && !stats ? '—' : formatTokens(stats?.requestCount ?? 0)}
        />
        <StatCard
          label="Tokens saved"
          value={loading && !stats ? '—' : formatTokens(stats?.tokensSaved ?? 0)}
        />
        <StatCard
          label="Estimated $ saved"
          value={loading && !stats ? '—' : formatUsd(stats?.savingsUsd ?? 0)}
        />
      </div>

      <section className="space-y-3">
        <h3 className="text-sm font-medium text-white">Per-model breakdown</h3>
        {byModel.length === 0 ? (
          <p className="rounded-lg border border-white/10 bg-black/15 px-4 py-6 text-sm text-slate-400">
            {loading ? 'Loading…' : 'No logged requests yet. Analyze a prompt or proxy traffic to populate this table.'}
          </p>
        ) : (
          <div className="overflow-hidden rounded-lg border border-white/10 bg-black/15">
            <table className="w-full text-left text-sm">
              <thead className="border-b border-white/10 text-xs uppercase tracking-[0.12em] text-slate-500">
                <tr>
                  <th className="px-4 py-2.5 font-medium">Model</th>
                  <th className="px-4 py-2.5 font-medium">Requests</th>
                  <th className="px-4 py-2.5 font-medium">Tokens saved</th>
                  <th className="px-4 py-2.5 font-medium">$ saved</th>
                </tr>
              </thead>
              <tbody>
                {byModel.map((row) => (
                  <tr key={row.model || '(unknown)'} className="border-t border-white/5">
                    <td className="px-4 py-2.5 text-slate-100">
                      {row.model || <span className="text-slate-500">unknown</span>}
                    </td>
                    <td className="px-4 py-2.5 tabular-nums text-slate-300">
                      {formatTokens(row.requestCount)}
                    </td>
                    <td className="px-4 py-2.5 tabular-nums text-slate-300">
                      {formatTokens(row.tokensSaved)}
                    </td>
                    <td className="px-4 py-2.5 tabular-nums text-emerald-300/90">
                      {formatUsd(row.savingsUsd)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <section className="space-y-3">
        <h3 className="text-sm font-medium text-white">Recent requests</h3>
        {recent.length === 0 ? (
          <p className="rounded-lg border border-white/10 bg-black/15 px-4 py-6 text-sm text-slate-400">
            {loading ? 'Loading…' : 'No recent requests.'}
          </p>
        ) : (
          <ul className="overflow-hidden rounded-lg border border-white/10 bg-black/15">
            {recent.map((req) => (
              <li key={req.id} className="border-t border-white/5 first:border-t-0">
                <button
                  type="button"
                  onClick={() => onOpenRequest?.(req)}
                  className="flex w-full flex-wrap items-center gap-x-4 gap-y-1 px-4 py-3 text-left hover:bg-white/5"
                >
                  <span className="min-w-[7rem] text-xs tabular-nums text-slate-500">
                    {formatWhen(req.createdAt)}
                  </span>
                  <span className="rounded bg-white/5 px-1.5 py-0.5 text-xs uppercase tracking-wide text-slate-400">
                    {req.source}
                  </span>
                  <span className="flex-1 text-sm text-slate-100">
                    {req.model || 'unknown model'}
                  </span>
                  <span className="text-sm tabular-nums text-slate-400">
                    {formatTokens(req.inputTokens)} → {formatTokens(req.optimizedTokens)}
                  </span>
                  <span className="text-sm tabular-nums text-sky-200">
                    {formatPct(req.savingsPct)}
                  </span>
                  <span className="text-sm tabular-nums text-emerald-300/90">
                    {formatUsd(req.savingsUsd)}
                  </span>
                </button>
              </li>
            ))}
          </ul>
        )}
        {recent.length > 0 && (
          <p className="text-xs text-slate-500">
            Click a row to open the Lint workspace with that model selected. Prompt text is not
            stored in the request log.
          </p>
        )}
      </section>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-white/10 bg-black/20 p-4">
      <p className="text-xs uppercase tracking-[0.15em] text-slate-400">{label}</p>
      <p className="mt-2 text-2xl font-semibold tabular-nums text-white">{value}</p>
    </div>
  )
}
