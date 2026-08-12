import { useCallback, useEffect, useState } from 'react'
import { GetDashboardStats, GetRecentRequests } from '../../wailsjs/go/main/App'
import { store } from '../../wailsjs/go/models'

const RECENT_LIMIT = 25

export function useDashboard() {
  const [stats, setStats] = useState<store.DashboardStats | null>(null)
  const [recent, setRecent] = useState<store.Request[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [nextStats, nextRecent] = await Promise.all([
        GetDashboardStats(),
        GetRecentRequests(RECENT_LIMIT),
      ])
      setStats(
        nextStats ?? {
          requestCount: 0,
          tokensSaved: 0,
          savingsUsd: 0,
          byModel: [],
        },
      )
      setRecent(nextRecent ?? [])
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to load dashboard')
      setStats(null)
      setRecent([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void refresh()
  }, [refresh])

  return { stats, recent, loading, error, refresh }
}
