import { useState } from 'react'
import { store } from '../wailsjs/go/models'
import { SidebarNav, type AppPage } from './components/SidebarNav'
import { ContextWorkspace } from './pages/ContextWorkspace'
import { Dashboard } from './pages/Dashboard'
import { LintWorkspace } from './pages/LintWorkspace'
import { Settings } from './pages/Settings'

function App() {
  const [page, setPage] = useState<AppPage>('lint')
  const [lintModel, setLintModel] = useState<string | undefined>(undefined)
  const [lintPrompt, setLintPrompt] = useState<string | undefined>(undefined)

  function openRequestInLint(request: store.Request) {
    if (request.model) {
      setLintModel(request.model)
    }
    setPage('lint')
  }

  function openContextInLint(markdown: string) {
    setLintPrompt(markdown)
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
            preferredPrompt={lintPrompt}
            onPreferredPromptConsumed={() => setLintPrompt(undefined)}
          />
        </div>
        <div
          className={page === 'context' ? undefined : 'hidden'}
          aria-hidden={page !== 'context'}
        >
          <ContextWorkspace
            active={page === 'context'}
            onOpenInLint={openContextInLint}
          />
        </div>
        <div
          className={page === 'dashboard' ? undefined : 'hidden'}
          aria-hidden={page !== 'dashboard'}
        >
          <Dashboard active={page === 'dashboard'} onOpenRequest={openRequestInLint} />
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
