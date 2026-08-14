import { useState } from 'react'
import { store } from '../wailsjs/go/models'
import { SidebarNav, type AppPage } from './components/SidebarNav'
import { Dashboard } from './pages/Dashboard'
import { LintWorkspace } from './pages/LintWorkspace'
import { Settings } from './pages/Settings'

function App() {
  const [page, setPage] = useState<AppPage>('lint')
  const [lintModel, setLintModel] = useState<string | undefined>(undefined)

  function openRequestInLint(request: store.Request) {
    if (request.model) {
      setLintModel(request.model)
    }
    setPage('lint')
  }

  return (
    <div className="flex h-full min-h-0 bg-window text-fg">
      <SidebarNav page={page} onNavigate={setPage} />
      <main className="min-w-0 flex-1 overflow-auto p-6">
        <div className={page === 'lint' ? undefined : 'hidden'} aria-hidden={page !== 'lint'}>
          <LintWorkspace
            preferredModel={lintModel}
            onPreferredModelConsumed={() => setLintModel(undefined)}
          />
        </div>
        <div
          className={page === 'dashboard' ? undefined : 'hidden'}
          aria-hidden={page !== 'dashboard'}
        >
          <Dashboard onOpenRequest={openRequestInLint} />
        </div>
        <div
          className={page === 'settings' ? undefined : 'hidden'}
          aria-hidden={page !== 'settings'}
        >
          <Settings />
        </div>
      </main>
    </div>
  )
}

export default App
