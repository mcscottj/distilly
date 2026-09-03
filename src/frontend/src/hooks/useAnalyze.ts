import { useCallback, useEffect, useState } from 'react'
import { Analyze, Apply, ListModels } from '../../wailsjs/go/main/App'
import { api } from '../../wailsjs/go/models'

export type ApplyToggles = {
  approveNearDuplicates: boolean
  approveJsonConversion: boolean
}

export function useAnalyze() {
  const [models, setModels] = useState<string[]>([])
  const [analysis, setAnalysis] = useState<api.AnalyzeResponse | null>(null)
  const [applyResult, setApplyResult] = useState<api.ApplyResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [applying, setApplying] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    ListModels()
      .then((list) => {
        if (!cancelled) setModels(list ?? [])
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load models')
        }
      })
    return () => {
      cancelled = true
    }
  }, [])

  const analyze = useCallback(async (prompt: string, model: string) => {
    setLoading(true)
    setError(null)
    try {
      const result = await Analyze({ prompt, model })
      setAnalysis(result)
      return result
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Analyze failed'
      setError(message)
      setAnalysis(null)
      throw err
    } finally {
      setLoading(false)
    }
  }, [])

  const apply = useCallback(async (prompt: string, toggles: ApplyToggles) => {
    setApplying(true)
    setError(null)
    try {
      const result = await Apply({
        prompt,
        approveNearDuplicates: toggles.approveNearDuplicates,
        approveJsonConversion: toggles.approveJsonConversion,
      })
      setApplyResult(result)
      return result
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Apply failed'
      setError(message)
      setApplyResult(null)
      throw err
    } finally {
      setApplying(false)
    }
  }, [])

  const clearResults = useCallback(() => {
    setAnalysis(null)
    setApplyResult(null)
    setError(null)
  }, [])

  return {
    models,
    analysis,
    applyResult,
    loading,
    applying,
    error,
    analyze,
    apply,
    clearResults,
    setError,
  }
}
