type ScoreCardProps = {
  score: number
  issues: string[] | null | undefined
  inputTokens?: number
  potentialSavings?: number
  costKnown?: boolean
  estimatedCostUsd?: number
  estimatedSavingsUsd?: number
}

function scoreTone(score: number): string {
  if (score >= 80) return 'text-emerald-300'
  if (score >= 50) return 'text-amber-300'
  return 'text-rose-300'
}

export function ScoreCard({
  score,
  issues,
  inputTokens,
  potentialSavings,
  costKnown,
  estimatedCostUsd,
  estimatedSavingsUsd,
}: ScoreCardProps) {
  const issueList = issues ?? []

  return (
    <section className="rounded-lg border border-white/10 bg-black/20 p-4">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <p className="text-xs uppercase tracking-[0.15em] text-slate-400">Prompt score</p>
          <p className={`mt-1 text-4xl font-semibold tabular-nums ${scoreTone(score)}`}>
            {score}
            <span className="text-lg font-normal text-slate-500"> / 100</span>
          </p>
        </div>
        <div className="text-right text-sm text-slate-300">
          {inputTokens != null && (
            <p>
              <span className="text-slate-500">Input tokens</span>{' '}
              <span className="tabular-nums text-slate-100">{inputTokens.toLocaleString()}</span>
            </p>
          )}
          {potentialSavings != null && potentialSavings > 0 && (
            <p>
              <span className="text-slate-500">Potential savings</span>{' '}
              <span className="tabular-nums text-sky-200">{potentialSavings.toFixed(1)}%</span>
            </p>
          )}
          {costKnown && estimatedCostUsd != null && (
            <p>
              <span className="text-slate-500">Est. cost</span>{' '}
              <span className="tabular-nums text-slate-100">${estimatedCostUsd.toFixed(4)}</span>
              {estimatedSavingsUsd != null && estimatedSavingsUsd > 0 && (
                <span className="text-emerald-300/90">
                  {' '}
                  (−${estimatedSavingsUsd.toFixed(4)})
                </span>
              )}
            </p>
          )}
        </div>
      </div>

      {issueList.length > 0 ? (
        <ul className="mt-4 space-y-1.5 border-t border-white/10 pt-3">
          {issueList.map((issue) => (
            <li key={issue} className="text-sm text-slate-300">
              <span className="mr-2 text-amber-400/80">•</span>
              {issue}
            </li>
          ))}
        </ul>
      ) : (
        <p className="mt-4 border-t border-white/10 pt-3 text-sm text-slate-400">
          No scoring issues detected.
        </p>
      )}
    </section>
  )
}
