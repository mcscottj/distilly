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
        {page === 'lint' && (
          <LintWorkspace
            preferredModel={lintModel}
            onPreferredModelConsumed={() => setLintModel(undefined)}
          />
        )}
        {page === 'dashboard' && <Dashboard onOpenRequest={openRequestInLint} />}
        {page === 'settings' && <Settings />}
      </main>
    </div>
  )
}

export default App
