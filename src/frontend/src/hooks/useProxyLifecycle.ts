import { useCallback, useEffect, useState } from 'react'
import { GetProxyStatus, StartProxy, StopProxy } from '../../wailsjs/go/main/App'

export type ProxyStatus = {
  running: boolean
  addr: string
}

const emptyStatus = (): ProxyStatus => ({ running: false, addr: '' })

function normalizeStatus(raw: { running?: boolean; addr?: string } | null | undefined): ProxyStatus {
  return {
    running: Boolean(raw?.running),
    addr: raw?.addr ?? '',
  }
}

export function useProxyLifecycle(pollMs = 2000) {
  const [status, setStatus] = useState<ProxyStatus>(emptyStatus)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    try {
      const next = await GetProxyStatus()
      setStatus(normalizeStatus(next))
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to read proxy status')
    }
  }, [])

  useEffect(() => {
    void refresh()
    const id = window.setInterval(() => {
      void refresh()
    }, pollMs)
    return () => window.clearInterval(id)
  }, [pollMs, refresh])

  const start = useCallback(async () => {
    setBusy(true)
    setError(null)
    try {
      await StartProxy()
      await refresh()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to start proxy')
      throw err
    } finally {
      setBusy(false)
    }
  }, [refresh])

  const stop = useCallback(async () => {
    setBusy(true)
    setError(null)
    try {
      await StopProxy()
      await refresh()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to stop proxy')
      throw err
    } finally {
      setBusy(false)
    }
  }, [refresh])

  return {
    status,
    busy,
    error,
    setError,
    refresh,
    start,
    stop,
  }
}
