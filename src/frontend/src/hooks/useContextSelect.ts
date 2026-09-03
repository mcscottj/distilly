import { useCallback, useState } from 'react'
import { SelectContext } from '../../wailsjs/go/main/App'
import { api } from '../../wailsjs/go/models'

export function useContextSelect() {
  const [result, setResult] = useState<api.SelectContextResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const select = useCallback(async (req: api.SelectContextRequest) => {
    setLoading(true)
    setError(null)
    try {
      const resp = await SelectContext(req)
      if (resp.error) {
        setError(resp.error)
        setResult(null)
        throw new Error(resp.error)
      }
      setResult(resp)
      return resp
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'SelectContext failed'
      setError(message)
      setResult(null)
      throw err
    } finally {
      setLoading(false)
    }
  }, [])

  const clearResults = useCallback(() => {
    setResult(null)
    setError(null)
  }, [])

  return {
    result,
    loading,
    error,
    select,
    clearResults,
    setError,
  }
}
