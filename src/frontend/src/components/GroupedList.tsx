import type { ReactNode } from 'react'

export function GroupedList({
  caption,
  footer,
  children,
}: {
  caption?: string
  footer?: ReactNode
  children: ReactNode
}) {
  return (
    <section className="space-y-2">
      {caption ? (
        <h3 className="px-1 text-xs font-medium text-muted">
          {caption}
        </h3>
      ) : null}
      <div className="overflow-hidden rounded-xl border border-hairline bg-surface">{children}</div>
      {footer ? <div className="px-1 text-xs text-muted">{footer}</div> : null}
    </section>
  )
}

export function GroupedRow({
  label,
  description,
  children,
}: {
  label: string
  description?: string
  children: ReactNode
}) {
  return (
    <div className="flex flex-col gap-2 border-t border-hairline px-4 py-3 first:border-t-0 sm:flex-row sm:items-center sm:justify-between">
      <div className="min-w-0 sm:max-w-[45%]">
        <p className="text-sm text-fg">{label}</p>
        {description ? <p className="mt-0.5 text-xs text-muted">{description}</p> : null}
      </div>
      <div className="min-w-0 sm:max-w-[55%] sm:text-right">{children}</div>
    </div>
  )
}
