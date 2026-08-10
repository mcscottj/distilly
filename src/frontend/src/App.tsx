import { FormEvent, useState } from 'react'
import { Greet } from '../wailsjs/go/main/App'

type Page = 'lint' | 'dashboard' | 'settings'

const navItems: { id: Page; label: string }[] = [
  { id: 'lint', label: 'Lint' },
  { id: 'dashboard', label: 'Dashboard' },
  { id: 'settings', label: 'Settings' },
]

function App() {
  const [page, setPage] = useState<Page>('lint')
  const [name, setName] = useState('')
  const [greeting, setGreeting] = useState('Wails scaffold ready, yo.')

  async function onGreet(e: FormEvent) {
    e.preventDefault()
    const result = await Greet(name || 'friend')
    setGreeting(result)
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      <header className="flex items-center justify-between border-b border-white/10 px-6 py-4">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-sky-300/80">Distilly</p>
          <h1 className="text-xl font-semibold tracking-tight text-white">Desktop</h1>
        </div>
        <nav className="flex gap-1">
          {navItems.map((item) => (
            <button
              key={item.id}
              type="button"
              onClick={() => setPage(item.id)}
              className={
                page === item.id
                  ? 'rounded-md bg-sky-500/20 px-3 py-1.5 text-sm text-sky-100'
                  : 'rounded-md px-3 py-1.5 text-sm text-slate-300 hover:bg-white/5'
              }
            >
              {item.label}
            </button>
          ))}
        </nav>
      </header>

      <main className="flex-1 overflow-auto p-6">
        {page === 'lint' && (
          <section className="mx-auto max-w-xl space-y-4">
            <h2 className="text-lg font-medium text-white">Lint workspace</h2>
            <p className="text-sm text-slate-300">
              Scaffold only — paste/analyze UI lands in the next slice. Binding smoke test below.
            </p>
            <form onSubmit={onGreet} className="flex gap-2">
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Your name"
                className="flex-1 rounded-md border border-white/15 bg-black/20 px-3 py-2 text-sm text-white outline-none focus:border-sky-400"
              />
              <button
                type="submit"
                className="rounded-md bg-sky-500 px-4 py-2 text-sm font-medium text-slate-950 hover:bg-sky-400"
              >
                Greet
              </button>
            </form>
            <p className="rounded-md border border-white/10 bg-black/15 px-3 py-2 text-sm text-sky-100">
              {greeting}
            </p>
          </section>
        )}

        {page === 'dashboard' && (
          <section className="mx-auto max-w-xl space-y-2">
            <h2 className="text-lg font-medium text-white">Dashboard</h2>
            <p className="text-sm text-slate-300">Cost aggregates arrive with the SQLite store slice.</p>
          </section>
        )}

        {page === 'settings' && (
          <section className="mx-auto max-w-xl space-y-2">
            <h2 className="text-lg font-medium text-white">Settings</h2>
            <p className="text-sm text-slate-300">Upstream URL, API key, and proxy toggles come next.</p>
          </section>
        )}
      </main>
    </div>
  )
}

export default App
