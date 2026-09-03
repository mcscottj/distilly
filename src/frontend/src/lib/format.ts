export function formatTokens(n: number): string {
  return Math.round(n).toLocaleString()
}

export function formatUsd(n: number, digits = 4): string {
  const abs = Math.abs(n)
  const fixed = abs < 0.01 && abs > 0 ? n.toFixed(Math.max(digits, 6)) : n.toFixed(digits)
  return `$${fixed}`
}

export function formatPct(n: number): string {
  return `${n.toFixed(1)}%`
}

export function formatWhen(iso: string): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString()
}
