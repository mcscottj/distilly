export type DiffLine = {
  marker: string
  content: string
}

type DiffViewProps = {
  lines: DiffLine[] | null | undefined
  emptyMessage?: string
}

function lineClass(marker: string): string {
  if (marker === '-') return 'bg-danger-soft text-danger'
  if (marker === '+') return 'bg-success-soft text-success'
  return 'text-fg'
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
      <div className="rounded-lg border border-dashed border-hairline bg-fill px-4 py-8 text-center text-sm text-muted">
        {emptyMessage}
      </div>
    )
  }

  return (
    <div className="overflow-hidden rounded-lg border border-hairline bg-surface">
      <div className="border-b border-hairline px-3 py-2 text-xs text-muted">
        Diff
      </div>
      <pre className="font-editor max-h-[28rem] overflow-auto p-0">
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
