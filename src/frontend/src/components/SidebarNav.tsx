import logo from '../assets/distilly-logo.png'

export type AppPage = 'lint' | 'dashboard' | 'settings'

const navItems: { id: AppPage; label: string }[] = [
  { id: 'lint', label: 'Lint' },
  { id: 'dashboard', label: 'Dashboard' },
  { id: 'settings', label: 'Settings' },
]

type SidebarNavProps = {
  page: AppPage
  onNavigate: (page: AppPage) => void
}

export function SidebarNav({ page, onNavigate }: SidebarNavProps) {
  return (
    <aside className="flex w-[220px] shrink-0 flex-col border-r border-hairline bg-sidebar">
      <div className="flex items-center gap-2 px-4 py-4">
        <img src={logo} alt="" className="size-9 object-contain" />
        <span className="text-[15px] font-semibold tracking-tight text-fg">distilly</span>
      </div>
      <nav className="flex flex-col gap-0.5 px-2 pb-4">
        {navItems.map((item) => {
          const active = page === item.id
          return (
            <button
              key={item.id}
              type="button"
              onClick={() => onNavigate(item.id)}
              className={
                active
                  ? 'rounded-md bg-accent-soft px-3 py-1.5 text-left text-sm font-medium text-accent'
                  : 'rounded-md px-3 py-1.5 text-left text-sm text-fg hover:bg-fill'
              }
            >
              {item.label}
            </button>
          )
        })}
      </nav>
    </aside>
  )
}
