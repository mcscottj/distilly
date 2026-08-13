import { FormEvent, useState } from 'react'
import { GroupedList, GroupedRow } from '../components/GroupedList'
import { SegmentedControl } from '../components/SegmentedControl'
import { useProxyLifecycle } from '../hooks/useProxyLifecycle'
import { useSettings } from '../hooks/useSettings'
import { useTheme } from '../hooks/useTheme'
import { proxyBaseURL } from '../lib/settings'

const fieldClass =
  'w-full rounded-md border border-hairline bg-fill px-3 py-2 text-sm text-fg outline-none focus:border-accent focus:ring-1 focus:ring-accent disabled:cursor-not-allowed disabled:opacity-50'

const secondaryButtonClass =
  'rounded-md border border-hairline bg-fill px-3 py-1.5 text-sm text-fg hover:bg-surface disabled:cursor-not-allowed disabled:opacity-50'

const primaryButtonClass =
  'rounded-md bg-accent px-3 py-1.5 text-sm font-medium text-accent-fg hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50'

const checkboxClass =
  'size-4 rounded border-hairline bg-fill text-accent focus:ring-accent'

export function Settings() {
  const { settings, models, loading, saving, error, savedAt, update, save, setError } = useSettings()
  const { preference, setPreference, loading: themeLoading } = useTheme()
  const {
    status: proxyStatus,
    busy: proxyBusy,
    error: proxyError,
    start: startProxy,
    stop: stopProxy,
  } = useProxyLifecycle()
  const [copied, setCopied] = useState(false)

  const formDisabled = loading || saving

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    await save()
  }

  async function copyBaseURL() {
    const url = proxyBaseURL(settings.proxyPort)
    try {
      await navigator.clipboard.writeText(url)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 2000)
    } catch {
      window.prompt('Copy proxy base URL', url)
    }
  }

  async function onStartProxy() {
    setError(null)
    try {
      await save()
      await startProxy()
    } catch {
      // Errors are surfaced via useSettings / useProxyLifecycle state.
    }
  }

  async function onStopProxy() {
    setError(null)
    try {
      await stopProxy()
    } catch {
      // Error state set in hook.
    }
  }

  const baseURL = proxyBaseURL(settings.proxyPort)
  const displayError = error ?? proxyError
  const proxyRunning = proxyStatus.running

  const statusDescription = proxyRunning
    ? proxyStatus.addr
      ? `Running — ${proxyStatus.addr}`
      : 'Running'
    : 'Stopped'

  return (
    <div className="mx-auto flex max-w-2xl flex-col gap-6">
      <div>
        <h2 className="text-lg font-medium text-fg">Settings</h2>
        <p className="mt-1 text-sm text-muted">
          Local API credentials, proxy defaults, and optimization opt-ins. Values stay in SQLite on
          this machine.
        </p>
      </div>

      {displayError && (
        <p className="rounded-md border border-danger bg-danger-soft px-3 py-2 text-sm text-danger">
          {displayError}
        </p>
      )}

      <GroupedList caption="Appearance">
        <GroupedRow label="Theme" description="Choose light, dark, or match the system.">
          <SegmentedControl
            value={preference}
            disabled={themeLoading}
            onChange={(v) => {
              void setPreference(v).catch((err) => {
                setError(err instanceof Error ? err.message : 'Failed to save theme preference')
              })
            }}
            options={[
              { value: 'light', label: 'Light' },
              { value: 'dark', label: 'Dark' },
              { value: 'system', label: 'System' },
            ]}
          />
        </GroupedRow>
      </GroupedList>

      <form onSubmit={(e) => void onSubmit(e)} className="space-y-6">
        <GroupedList caption="Upstream">
          <GroupedRow
            label="Upstream base URL"
            description="OpenAI-compatible endpoints work (OpenRouter, Azure gateways, local proxies)."
          >
            <input
              type="url"
              value={settings.upstreamURL}
              onChange={(e) => update('upstreamURL', e.target.value)}
              placeholder="https://api.openai.com"
              disabled={formDisabled}
              className={fieldClass}
            />
          </GroupedRow>

          <GroupedRow
            label="API key"
            description="Stored locally only. Never logged. Used as Authorization Bearer for proxied requests."
          >
            <input
              type="password"
              value={settings.apiKey}
              onChange={(e) => update('apiKey', e.target.value)}
              placeholder="sk-…"
              autoComplete="off"
              spellCheck={false}
              disabled={formDisabled}
              className={`${fieldClass} font-mono`}
            />
          </GroupedRow>

          <GroupedRow label="Default model">
            <select
              value={settings.defaultModel}
              onChange={(e) => update('defaultModel', e.target.value)}
              disabled={formDisabled}
              className={fieldClass}
            >
              <option value="">Use first listed model</option>
              {models.map((m) => (
                <option key={m} value={m}>
                  {m}
                </option>
              ))}
            </select>
          </GroupedRow>
        </GroupedList>

        <GroupedList
          caption="Local proxy"
          footer={
            <>
              Clients must send <code className="text-fg">stream: false</code>. Streaming is not
              supported in the M4 MVP.
            </>
          }
        >
          <GroupedRow label="Status" description={statusDescription}>
            {proxyRunning ? (
              <button
                type="button"
                onClick={() => void onStopProxy()}
                disabled={proxyBusy}
                className={secondaryButtonClass}
              >
                {proxyBusy ? 'Stopping…' : 'Stop proxy'}
              </button>
            ) : (
              <button
                type="button"
                onClick={() => void onStartProxy()}
                disabled={proxyBusy || formDisabled}
                className={primaryButtonClass}
              >
                {proxyBusy || saving ? 'Starting…' : 'Start proxy'}
              </button>
            )}
          </GroupedRow>

          <GroupedRow
            label="Proxy port"
            description={proxyRunning ? 'Stop the proxy to change the port.' : undefined}
          >
            <input
              type="text"
              inputMode="numeric"
              pattern="[0-9]*"
              value={settings.proxyPort}
              onChange={(e) => update('proxyPort', e.target.value.replace(/[^\d]/g, ''))}
              placeholder="8787"
              disabled={formDisabled || proxyRunning || proxyBusy}
              className={`${fieldClass} sm:ml-auto sm:w-40`}
            />
          </GroupedRow>

          <GroupedRow label="Base URL">
            <div className="flex flex-wrap items-center justify-end gap-2">
              <span className="truncate font-mono text-sm text-accent">{baseURL}</span>
              <button type="button" onClick={() => void copyBaseURL()} className={secondaryButtonClass}>
                {copied ? 'Copied' : 'Copy'}
              </button>
            </div>
          </GroupedRow>

          <GroupedRow
            label="Passthrough mode"
            description="Analyze and log only; forward the original request body unchanged."
          >
            <input
              type="checkbox"
              checked={settings.passthrough}
              onChange={(e) => update('passthrough', e.target.checked)}
              disabled={formDisabled || proxyBusy}
              className={checkboxClass}
            />
          </GroupedRow>
        </GroupedList>

        <GroupedList
          caption="Optimization defaults"
          footer="Exact duplicates always apply. Near-duplicates and JSON conversion stay off unless you opt in (same as the CLI)."
        >
          <GroupedRow label="Approve near-duplicates by default">
            <input
              type="checkbox"
              checked={settings.approveNearDuplicates}
              onChange={(e) => update('approveNearDuplicates', e.target.checked)}
              disabled={formDisabled}
              className={checkboxClass}
            />
          </GroupedRow>

          <GroupedRow label="Approve JSON conversion by default">
            <input
              type="checkbox"
              checked={settings.approveJsonConversion}
              onChange={(e) => update('approveJsonConversion', e.target.checked)}
              disabled={formDisabled}
              className={checkboxClass}
            />
          </GroupedRow>
        </GroupedList>

        <div className="flex flex-wrap items-center gap-3">
          <button
            type="submit"
            disabled={formDisabled}
            className="rounded-md bg-accent px-4 py-2 text-sm font-medium text-accent-fg hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {saving ? 'Saving…' : loading ? 'Loading…' : 'Save settings'}
          </button>
          {savedAt != null && <span className="text-sm text-success">Saved</span>}
        </div>
      </form>
    </div>
  )
}
