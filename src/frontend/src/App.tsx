import { useState } from 'react'
import { store } from '../wailsjs/go/models'
import { Dashboard } from './pages/Dashboard'
import { LintWorkspace } from './pages/LintWorkspace'
import { Settings } from './pages/Settings'

type Page = 'lint' | 'dashboard' | 'settings'

const navItems: { id: Page; label: string }[] = [
  { id: 'lint', label: 'Lint' },
  { id: 'dashboard', label: 'Dashboard' },
  { id: 'settings', label: 'Settings' },
]

function App() {
  const [page, setPage] = useState<Page>('lint')
  const [lintModel, setLintModel] = useState<string | undefined>(undefined)

  function openRequestInLint(request: store.Request) {
    if (request.model) {
      setLintModel(request.model)
    }
    setPage('lint')
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
          <LintWorkspace preferredModel={lintModel} onPreferredModelConsumed={() => setLintModel(undefined)} />
        )}
        {page === 'dashboard' && <Dashboard onOpenRequest={openRequestInLint} />}
        {page === 'settings' && <Settings />}
      </main>
    </div>
  )
}

export default App
