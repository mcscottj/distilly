type Sections = {
  system: number
  examples: number
  history: number
  question: number
}

type SectionBreakdownProps = {
  sections: Sections
}

const LABELS: { key: keyof Sections; label: string }[] = [
  { key: 'system', label: 'System' },
  { key: 'examples', label: 'Examples' },
  { key: 'history', label: 'History' },
  { key: 'question', label: 'Question' },
]

export function SectionBreakdown({ sections }: SectionBreakdownProps) {
  const total = LABELS.reduce((sum, { key }) => sum + (sections[key] ?? 0), 0)
  const max = Math.max(total, 1)

  return (
    <section className="rounded-lg border border-white/10 bg-black/20 p-4">
      <h3 className="text-sm font-medium text-white">Section tokens</h3>
      <p className="mt-0.5 text-xs text-slate-400">
        {total.toLocaleString()} tokens across sections
      </p>
      <ul className="mt-4 space-y-3">
        {LABELS.map(({ key, label }) => {
          const value = sections[key] ?? 0
          const pct = (value / max) * 100
          return (
            <li key={key}>
              <div className="mb-1 flex items-baseline justify-between text-sm">
                <span className="text-slate-300">{label}</span>
                <span className="tabular-nums text-slate-100">{value.toLocaleString()}</span>
              </div>
              <div className="h-1.5 overflow-hidden rounded-full bg-white/10">
                <div
                  className="h-full rounded-full bg-sky-500/70"
                  style={{ width: `${pct}%` }}
                />
              </div>
            </li>
          )
        })}
      </ul>
    </section>
  )
}
