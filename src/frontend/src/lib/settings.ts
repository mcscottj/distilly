/** SQLite settings keys (must match store / future proxy reads). */
export const SettingKey = {
  UpstreamURL: 'upstream_url',
  ApiKey: 'api_key',
  DefaultModel: 'default_model',
  ProxyPort: 'proxy_port',
  ApproveNearDuplicates: 'approve_near_duplicates',
  ApproveJsonConversion: 'approve_json_conversion',
  Passthrough: 'passthrough',
  Theme: 'theme',
  ThemeContrast: 'theme_contrast',
  UiFontScale: 'ui_font_scale',
  EditorFontSize: 'editor_font_size',
} as const

export type SettingKeyName = (typeof SettingKey)[keyof typeof SettingKey]

export const SETTING_DEFAULTS: Record<SettingKeyName, string> = {
  [SettingKey.UpstreamURL]: 'https://api.openai.com',
  [SettingKey.ApiKey]: '',
  [SettingKey.DefaultModel]: '',
  [SettingKey.ProxyPort]: '8787',
  [SettingKey.ApproveNearDuplicates]: 'false',
  [SettingKey.ApproveJsonConversion]: 'false',
  [SettingKey.Passthrough]: 'false',
  [SettingKey.Theme]: 'light',
  [SettingKey.ThemeContrast]: 'normal',
  [SettingKey.UiFontScale]: 'default',
  [SettingKey.EditorFontSize]: '14',
}

export function parseBoolSetting(value: string | null | undefined): boolean {
  return value === 'true' || value === '1'
}

export function serializeBoolSetting(value: boolean): string {
  return value ? 'true' : 'false'
}

/** Resolve stored value or documented default. Empty string for optional fields stays empty. */
export function withSettingDefault(key: SettingKeyName, value: string | null | undefined): string {
  if (value == null || value === '') {
    // Keep empty for secrets / optional selectors; apply defaults only when meaningful.
    if (key === SettingKey.ApiKey || key === SettingKey.DefaultModel) {
      return ''
    }
    return SETTING_DEFAULTS[key]
  }
  return value
}

export function proxyBaseURL(port: string): string {
  const trimmed = (port || SETTING_DEFAULTS[SettingKey.ProxyPort]).trim()
  return `http://127.0.0.1:${trimmed}/v1`
}
