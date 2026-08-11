export type DiffLine = {
  marker: string
  content: string
}

type DiffViewProps = {
  lines: DiffLine[] | null | undefined
  emptyMessage?: string
}

function lineClass(marker: string): string {
  if (marker === '-') return 'bg-rose-950/50 text-rose-200'
  if (marker === '+') return 'bg-emerald-950/50 text-emerald-200'
  return 'text-slate-300'
}

function markerLabel(marker: string): string {
  if (marker === '-') return '−'
  if (marker === '+') return '+'
  return ' '
}

export function DiffView({
  lines,
  emptyMessage = 'Run Apply to preview the optimized diff.',
}: DiffViewProps) {
  const rows = lines ?? []

  if (rows.length === 0) {
    return (
      <div className="rounded-lg border border-dashed border-white/15 bg-black/10 px-4 py-8 text-center text-sm text-slate-400">
        {emptyMessage}
      </div>
    )
  }

  return (
    <div className="overflow-hidden rounded-lg border border-white/10 bg-black/30">
      <div className="border-b border-white/10 px-3 py-2 text-xs uppercase tracking-[0.12em] text-slate-400">
        Diff
      </div>
      <pre className="max-h-[28rem] overflow-auto p-0 font-mono text-xs leading-5">
        {rows.map((line, i) => (
          <div
            key={`${i}-${line.marker}-${line.content.slice(0, 24)}`}
            className={`flex whitespace-pre-wrap break-all px-3 py-0.5 ${lineClass(line.marker)}`}
          >
            <span className="mr-3 w-3 shrink-0 select-none opacity-70">
              {markerLabel(line.marker)}
            </span>
            <span>{line.content || ' '}</span>
          </div>
        ))}
      </pre>
    </div>
  )
}
