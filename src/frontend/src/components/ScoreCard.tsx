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
  if (score >= 80) return 'text-success'
  if (score >= 50) return 'text-warning'
  return 'text-danger'
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
    <section className="rounded-xl border border-hairline bg-surface p-4">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <p className="text-xs text-muted">Prompt score</p>
          <p className={`mt-1 text-4xl font-semibold tabular-nums ${scoreTone(score)}`}>
            {score}
            <span className="text-lg font-normal text-muted"> / 100</span>
          </p>
        </div>
        <div className="text-right text-sm text-fg">
          {inputTokens != null && (
            <p>
              <span className="text-muted">Input tokens</span>{' '}
              <span className="tabular-nums text-fg">{inputTokens.toLocaleString()}</span>
            </p>
          )}
          {potentialSavings != null && potentialSavings > 0 && (
            <p>
              <span className="text-muted">Potential savings</span>{' '}
              <span className="tabular-nums text-accent">{potentialSavings.toFixed(1)}%</span>
            </p>
          )}
          {costKnown && estimatedCostUsd != null && (
            <p>
              <span className="text-muted">Est. cost</span>{' '}
              <span className="tabular-nums text-fg">${estimatedCostUsd.toFixed(4)}</span>
              {estimatedSavingsUsd != null && estimatedSavingsUsd > 0 && (
                <span className="text-success">
                  {' '}
                  (−${estimatedSavingsUsd.toFixed(4)})
                </span>
              )}
            </p>
          )}
        </div>
      </div>

      {issueList.length > 0 ? (
        <ul className="mt-4 space-y-1.5 border-t border-hairline pt-3">
          {issueList.map((issue) => (
            <li key={issue} className="text-sm text-fg">
              <span className="mr-2 text-warning">•</span>
              {issue}
            </li>
          ))}
        </ul>
      ) : (
        <p className="mt-4 border-t border-hairline pt-3 text-sm text-muted">
          No scoring issues detected.
        </p>
      )}
    </section>
  )
}
