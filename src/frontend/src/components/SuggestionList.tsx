type SuggestionListProps = {
  suggestions: string[] | null | undefined
}

export function SuggestionList({ suggestions }: SuggestionListProps) {
  const items = suggestions ?? []

  return (
    <section className="rounded-lg border border-white/10 bg-black/20 p-4">
      <h3 className="text-sm font-medium text-white">Suggestions</h3>
      {items.length === 0 ? (
        <p className="mt-2 text-sm text-slate-400">No suggestions — prompt looks clean.</p>
      ) : (
        <ol className="mt-3 list-decimal space-y-2 pl-5 text-sm text-slate-300">
          {items.map((s) => (
            <li key={s}>{s}</li>
          ))}
        </ol>
      )}
    </section>
  )
}
