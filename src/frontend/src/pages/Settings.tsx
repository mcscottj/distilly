import { FormEvent, useState } from 'react'
import { useProxyLifecycle } from '../hooks/useProxyLifecycle'
import { useSettings } from '../hooks/useSettings'
import { proxyBaseURL } from '../lib/settings'

export function Settings() {
  const { settings, models, loading, saving, error, savedAt, update, save, setError } = useSettings()
  const {
    status: proxyStatus,
    busy: proxyBusy,
    error: proxyError,
    start: startProxy,
    stop: stopProxy,
  } = useProxyLifecycle()
  const [copied, setCopied] = useState(false)

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
      // Fallback: select via prompt for environments without clipboard permission
      window.prompt('Copy proxy base URL', url)
    }
  }

  async function onStartProxy() {
    setError(null)
    try {
      // Persist port (and other fields) before binding so StartProxy reads SQLite.
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

  return (
    <div className="mx-auto flex max-w-2xl flex-col gap-6">
      <div>
        <h2 className="text-lg font-medium text-white">Settings</h2>
        <p className="mt-1 text-sm text-slate-400">
          Local API credentials, proxy defaults, and optimization opt-ins. Values stay in SQLite on
          this machine.
        </p>
      </div>

      {displayError && (
        <p className="rounded-md border border-rose-500/40 bg-rose-950/40 px-3 py-2 text-sm text-rose-200">
          {displayError}
        </p>
      )}

      <form onSubmit={(e) => void onSubmit(e)} className="space-y-6">
        <fieldset disabled={loading || saving} className="space-y-4">
          <legend className="mb-1 text-sm font-medium text-white">Upstream</legend>

          <label className="flex flex-col gap-1.5 text-sm">
            <span className="text-slate-400">Upstream base URL</span>
            <input
              type="url"
              value={settings.upstreamURL}
              onChange={(e) => update('upstreamURL', e.target.value)}
              placeholder="https://api.openai.com"
              className="rounded-md border border-white/15 bg-black/25 px-3 py-2 text-sm text-white outline-none focus:border-sky-400"
            />
            <span className="text-xs text-slate-500">
              OpenAI-compatible endpoints work (OpenRouter, Azure gateways, local proxies).
            </span>
          </label>

          <label className="flex flex-col gap-1.5 text-sm">
            <span className="text-slate-400">API key</span>
            <input
              type="password"
              value={settings.apiKey}
              onChange={(e) => update('apiKey', e.target.value)}
              placeholder="sk-…"
              autoComplete="off"
              spellCheck={false}
              className="rounded-md border border-white/15 bg-black/25 px-3 py-2 font-mono text-sm text-white outline-none focus:border-sky-400"
            />
            <span className="text-xs text-slate-500">
              Stored locally only. Never logged. Used as Authorization Bearer for proxied requests.
            </span>
          </label>

          <label className="flex flex-col gap-1.5 text-sm">
            <span className="text-slate-400">Default model</span>
            <select
              value={settings.defaultModel}
              onChange={(e) => update('defaultModel', e.target.value)}
              className="rounded-md border border-white/15 bg-black/25 px-3 py-2 text-sm text-white outline-none focus:border-sky-400"
            >
              <option value="">Use first listed model</option>
              {models.map((m) => (
                <option key={m} value={m}>
                  {m}
                </option>
              ))}
            </select>
          </label>
        </fieldset>

        <fieldset disabled={loading || saving || proxyBusy} className="space-y-4">
          <legend className="mb-1 text-sm font-medium text-white">Local proxy</legend>

          <div className="flex flex-wrap items-center gap-3 rounded-lg border border-white/10 bg-black/15 px-3 py-3">
            <div className="min-w-0 flex-1">
              <p className="text-xs uppercase tracking-[0.12em] text-slate-500">Status</p>
              <p className="mt-0.5 text-sm text-slate-100">
                {proxyRunning ? (
                  <>
                    Running
                    {proxyStatus.addr ? (
                      <span className="ml-2 font-mono text-sky-200">{proxyStatus.addr}</span>
                    ) : null}
                  </>
                ) : (
                  'Stopped'
                )}
              </p>
            </div>
            {proxyRunning ? (
              <button
                type="button"
                onClick={() => void onStopProxy()}
                disabled={proxyBusy}
                className="rounded-md border border-white/20 px-3 py-1.5 text-sm text-slate-200 hover:bg-white/5 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {proxyBusy ? 'Stopping…' : 'Stop proxy'}
              </button>
            ) : (
              <button
                type="button"
                onClick={() => void onStartProxy()}
                disabled={proxyBusy || loading || saving}
                className="rounded-md bg-sky-500 px-3 py-1.5 text-sm font-medium text-slate-950 hover:bg-sky-400 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {proxyBusy || saving ? 'Starting…' : 'Start proxy'}
              </button>
            )}
          </div>

          <label className="flex flex-col gap-1.5 text-sm">
            <span className="text-slate-400">Proxy port</span>
            <input
              type="text"
              inputMode="numeric"
              pattern="[0-9]*"
              value={settings.proxyPort}
              onChange={(e) => update('proxyPort', e.target.value.replace(/[^\d]/g, ''))}
              placeholder="8787"
              disabled={proxyRunning}
              className="w-40 rounded-md border border-white/15 bg-black/25 px-3 py-2 text-sm text-white outline-none focus:border-sky-400 disabled:cursor-not-allowed disabled:opacity-60"
            />
            {proxyRunning && (
              <span className="text-xs text-slate-500">Stop the proxy to change the port.</span>
            )}
          </label>

          <div className="flex flex-wrap items-center gap-2 rounded-lg border border-white/10 bg-black/15 px-3 py-3">
            <div className="min-w-0 flex-1">
              <p className="text-xs uppercase tracking-[0.12em] text-slate-500">Base URL</p>
              <p className="mt-0.5 truncate font-mono text-sm text-sky-200">{baseURL}</p>
            </div>
            <button
              type="button"
              onClick={() => void copyBaseURL()}
              className="rounded-md border border-white/20 px-3 py-1.5 text-sm text-slate-200 hover:bg-white/5"
            >
              {copied ? 'Copied' : 'Copy'}
            </button>
          </div>

          <p className="text-xs text-slate-500">
            Clients must send <code className="text-slate-400">stream: false</code>. Streaming is
            not supported in the M4 MVP.
          </p>

          <label className="flex cursor-pointer items-start gap-2 text-sm text-slate-300">
            <input
              type="checkbox"
              checked={settings.passthrough}
              onChange={(e) => update('passthrough', e.target.checked)}
              className="mt-0.5 size-4 rounded border-white/20 bg-black/40 text-sky-500 focus:ring-sky-400"
            />
            <span>
              <span className="text-slate-100">Passthrough mode</span>
              <span className="mt-0.5 block text-xs text-slate-500">
                Analyze and log only; forward the original request body unchanged.
              </span>
            </span>
          </label>
        </fieldset>

        <fieldset disabled={loading || saving} className="space-y-3">
          <legend className="mb-1 text-sm font-medium text-white">Optimization defaults</legend>
          <p className="text-xs text-slate-500">
            Exact duplicates always apply. Near-duplicates and JSON conversion stay off unless you
            opt in (same as the CLI).
          </p>

          <label className="flex cursor-pointer items-center gap-2 text-sm text-slate-300">
            <input
              type="checkbox"
              checked={settings.approveNearDuplicates}
              onChange={(e) => update('approveNearDuplicates', e.target.checked)}
              className="size-4 rounded border-white/20 bg-black/40 text-sky-500 focus:ring-sky-400"
            />
            Approve near-duplicates by default
          </label>

          <label className="flex cursor-pointer items-center gap-2 text-sm text-slate-300">
            <input
              type="checkbox"
              checked={settings.approveJsonConversion}
              onChange={(e) => update('approveJsonConversion', e.target.checked)}
              className="size-4 rounded border-white/20 bg-black/40 text-sky-500 focus:ring-sky-400"
            />
            Approve JSON conversion by default
          </label>
        </fieldset>

        <div className="flex flex-wrap items-center gap-3">
          <button
            type="submit"
            disabled={loading || saving}
            className="rounded-md bg-sky-500 px-4 py-2 text-sm font-medium text-slate-950 hover:bg-sky-400 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {saving ? 'Saving…' : loading ? 'Loading…' : 'Save settings'}
          </button>
          {savedAt != null && (
            <span className="text-sm text-emerald-300/90">Saved</span>
          )}
        </div>
      </form>
    </div>
  )
}
