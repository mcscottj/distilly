import { useCallback, useEffect, useState } from 'react'
import { GetSetting, ListModels, SetSetting } from '../../wailsjs/go/main/App'
import {
  SettingKey,
  parseBoolSetting,
  serializeBoolSetting,
  withSettingDefault,
} from '../lib/settings'

export type AppSettings = {
  upstreamURL: string
  apiKey: string
  defaultModel: string
  proxyPort: string
  approveNearDuplicates: boolean
  approveJsonConversion: boolean
  passthrough: boolean
  repoRoot: string
  contextMaxDepth: string
  contextMaxTokens: string
  enableCodeContext: boolean
}

const emptySettings = (): AppSettings => ({
  upstreamURL: withSettingDefault(SettingKey.UpstreamURL, ''),
  apiKey: '',
  defaultModel: '',
  proxyPort: withSettingDefault(SettingKey.ProxyPort, ''),
  approveNearDuplicates: false,
  approveJsonConversion: false,
  passthrough: false,
  repoRoot: '',
  contextMaxDepth: withSettingDefault(SettingKey.ContextMaxDepth, ''),
  contextMaxTokens: withSettingDefault(SettingKey.ContextMaxTokens, ''),
  enableCodeContext: false,
})

async function loadSettings(): Promise<AppSettings> {
  const [
    upstreamURL,
    apiKey,
    defaultModel,
    proxyPort,
    near,
    json,
    passthrough,
    repoRoot,
    contextMaxDepth,
    contextMaxTokens,
    enableCodeContext,
  ] = await Promise.all([
    GetSetting(SettingKey.UpstreamURL),
    GetSetting(SettingKey.ApiKey),
    GetSetting(SettingKey.DefaultModel),
    GetSetting(SettingKey.ProxyPort),
    GetSetting(SettingKey.ApproveNearDuplicates),
    GetSetting(SettingKey.ApproveJsonConversion),
    GetSetting(SettingKey.Passthrough),
    GetSetting(SettingKey.RepoRoot),
    GetSetting(SettingKey.ContextMaxDepth),
    GetSetting(SettingKey.ContextMaxTokens),
    GetSetting(SettingKey.EnableCodeContext),
  ])

  return {
    upstreamURL: withSettingDefault(SettingKey.UpstreamURL, upstreamURL),
    apiKey: apiKey ?? '',
    defaultModel: defaultModel ?? '',
    proxyPort: withSettingDefault(SettingKey.ProxyPort, proxyPort),
    approveNearDuplicates: parseBoolSetting(near),
    approveJsonConversion: parseBoolSetting(json),
    passthrough: parseBoolSetting(passthrough),
    repoRoot: repoRoot ?? '',
    contextMaxDepth: withSettingDefault(SettingKey.ContextMaxDepth, contextMaxDepth),
    contextMaxTokens: withSettingDefault(SettingKey.ContextMaxTokens, contextMaxTokens),
    enableCodeContext: parseBoolSetting(enableCodeContext),
  }
}

export function useSettings() {
  const [settings, setSettings] = useState<AppSettings>(emptySettings)
  const [models, setModels] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [savedAt, setSavedAt] = useState<number | null>(null)

  const refresh = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [next, modelList] = await Promise.all([loadSettings(), ListModels()])
      setSettings(next)
      setModels(modelList ?? [])
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to load settings')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void refresh()
  }, [refresh])

  const update = useCallback(<K extends keyof AppSettings>(key: K, value: AppSettings[K]) => {
    setSettings((prev) => ({ ...prev, [key]: value }))
    setSavedAt(null)
  }, [])

  const save = useCallback(async () => {
    setSaving(true)
    setError(null)
    try {
      const port = settings.proxyPort.trim() || withSettingDefault(SettingKey.ProxyPort, '')
      const upstream =
        settings.upstreamURL.trim() || withSettingDefault(SettingKey.UpstreamURL, '')

      await Promise.all([
        SetSetting(SettingKey.UpstreamURL, upstream),
        SetSetting(SettingKey.ApiKey, settings.apiKey),
        SetSetting(SettingKey.DefaultModel, settings.defaultModel),
        SetSetting(SettingKey.ProxyPort, port),
        SetSetting(
          SettingKey.ApproveNearDuplicates,
          serializeBoolSetting(settings.approveNearDuplicates),
        ),
        SetSetting(
          SettingKey.ApproveJsonConversion,
          serializeBoolSetting(settings.approveJsonConversion),
        ),
        SetSetting(SettingKey.Passthrough, serializeBoolSetting(settings.passthrough)),
        SetSetting(SettingKey.RepoRoot, settings.repoRoot.trim()),
        SetSetting(
          SettingKey.ContextMaxDepth,
          settings.contextMaxDepth.trim() ||
            withSettingDefault(SettingKey.ContextMaxDepth, ''),
        ),
        SetSetting(
          SettingKey.ContextMaxTokens,
          settings.contextMaxTokens.trim() ||
            withSettingDefault(SettingKey.ContextMaxTokens, ''),
        ),
        SetSetting(
          SettingKey.EnableCodeContext,
          serializeBoolSetting(settings.enableCodeContext),
        ),
      ])
      setSettings((prev) => ({ ...prev, upstreamURL: upstream, proxyPort: port }))
      setSavedAt(Date.now())
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to save settings')
      throw err
    } finally {
      setSaving(false)
    }
  }, [settings])

  return {
    settings,
    models,
    loading,
    saving,
    error,
    savedAt,
    update,
    save,
    refresh,
    setError,
  }
}
