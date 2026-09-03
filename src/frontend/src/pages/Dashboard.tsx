import { store } from '../../wailsjs/go/models'
import { useDashboard } from '../hooks/useDashboard'
import { formatPct, formatTokens, formatUsd, formatWhen } from '../lib/format'

const secondaryButtonClass =
  'rounded-md border border-hairline bg-fill px-3 py-1.5 text-sm text-fg hover:bg-surface disabled:cursor-not-allowed disabled:opacity-50'

type DashboardProps = {
  active?: boolean
  onOpenRequest?: (request: store.Request) => void
}

export function Dashboard({ active = true, onOpenRequest }: DashboardProps) {
  const { stats, recent, loading, error, refresh } = useDashboard(active)

  const byModel = stats?.byModel ?? []

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-medium text-fg">Dashboard</h2>
          <p className="mt-1 text-sm text-muted">
            Session and historical savings from logged manual and proxy requests.
          </p>
        </div>
        <button
          type="button"
          onClick={() => void refresh()}
          disabled={loading}
          className={secondaryButtonClass}
        >
          {loading ? 'Refreshing…' : 'Refresh'}
        </button>
      </div>

      {error && (
        <p className="rounded-md border border-danger bg-danger-soft px-3 py-2 text-sm text-danger">
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
        <h3 className="text-sm font-medium text-fg">Per-model breakdown</h3>
        {byModel.length === 0 ? (
          <p className="rounded-xl border border-hairline bg-surface px-4 py-6 text-sm text-muted">
            {loading ? 'Loading…' : 'No logged requests yet. Analyze a prompt or proxy traffic to populate this table.'}
          </p>
        ) : (
          <div className="overflow-hidden rounded-xl border border-hairline bg-surface">
            <table className="w-full text-left text-sm">
              <thead className="border-b border-hairline text-xs font-medium text-muted">
                <tr>
                  <th className="px-4 py-2.5">Model</th>
                  <th className="px-4 py-2.5">Requests</th>
                  <th className="px-4 py-2.5">Tokens saved</th>
                  <th className="px-4 py-2.5">$ saved</th>
                </tr>
              </thead>
              <tbody>
                {byModel.map((row) => (
                  <tr key={row.model || '(unknown)'} className="border-t border-hairline">
                    <td className="px-4 py-2.5 text-fg">
                      {row.model || <span className="text-muted">unknown</span>}
                    </td>
                    <td className="px-4 py-2.5 tabular-nums text-fg">
                      {formatTokens(row.requestCount)}
                    </td>
                    <td className="px-4 py-2.5 tabular-nums text-fg">
                      {formatTokens(row.tokensSaved)}
                    </td>
                    <td className="px-4 py-2.5 tabular-nums text-success">
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
        <h3 className="text-sm font-medium text-fg">Recent requests</h3>
        {recent.length === 0 ? (
          <p className="rounded-xl border border-hairline bg-surface px-4 py-6 text-sm text-muted">
            {loading ? 'Loading…' : 'No recent requests.'}
          </p>
        ) : (
          <ul className="overflow-hidden rounded-xl border border-hairline bg-surface">
            {recent.map((req) => (
              <li key={req.id} className="border-t border-hairline first:border-t-0">
                <button
                  type="button"
                  onClick={() => onOpenRequest?.(req)}
                  className="flex w-full flex-wrap items-center gap-x-4 gap-y-1 px-4 py-3 text-left hover:bg-fill"
                >
                  <span className="min-w-[7rem] text-xs tabular-nums text-muted">
                    {formatWhen(req.createdAt)}
                  </span>
                  <span className="rounded bg-fill px-1.5 py-0.5 text-xs text-accent">
                    {req.source}
                  </span>
                  <span className="flex-1 text-sm text-fg">
                    {req.model || 'unknown model'}
                  </span>
                  <span className="text-sm tabular-nums text-muted">
                    {formatTokens(req.inputTokens)} → {formatTokens(req.optimizedTokens)}
                  </span>
                  <span className="text-sm tabular-nums text-accent">
                    {formatPct(req.savingsPct)}
                  </span>
                  <span className="text-sm tabular-nums text-success">
                    {formatUsd(req.savingsUsd)}
                  </span>
                </button>
              </li>
            ))}
          </ul>
        )}
        {recent.length > 0 && (
          <p className="text-xs text-muted">
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
    <div className="rounded-xl border border-hairline bg-surface p-4">
      <p className="text-xs text-muted">{label}</p>
      <p className="mt-2 text-2xl font-semibold tabular-nums text-fg">{value}</p>
    </div>
  )
}
