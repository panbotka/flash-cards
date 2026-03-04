import type { ReactNode } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'

interface Tab {
  label: string
  path: string
  icon: (active: boolean) => ReactNode
}

const tabs: Tab[] = [
  {
    label: 'Study',
    path: '/',
    icon: (active) => (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke={active ? '#5e9eff' : '#6e6e73'} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <rect x="4" y="3" width="16" height="18" rx="2" />
        <path d="M12 3v18" />
        <path d="M4 9h8" />
        <path d="M12 9h8" />
      </svg>
    ),
  },
  {
    label: 'Cards',
    path: '/cards',
    icon: (active) => (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke={active ? '#5e9eff' : '#6e6e73'} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="5" width="18" height="14" rx="2" />
        <path d="M3 9h18" />
        <path d="M3 13h18" />
        <path d="M3 17h18" />
      </svg>
    ),
  },
  {
    label: 'Stats',
    path: '/stats',
    icon: (active) => (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke={active ? '#5e9eff' : '#6e6e73'} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <rect x="4" y="12" width="4" height="8" rx="1" />
        <rect x="10" y="8" width="4" height="12" rx="1" />
        <rect x="16" y="4" width="4" height="16" rx="1" />
      </svg>
    ),
  },
]

export function NavBar() {
  const location = useLocation()
  const navigate = useNavigate()

  return (
    <nav
      className="fixed bottom-0 left-0 right-0 bg-[#0a0a0a] border-t border-[#2a2a2a] z-50"
      style={{ paddingBottom: 'env(safe-area-inset-bottom, 0px)' }}
    >
      <div className="flex items-center justify-around h-16 px-4">
        {tabs.map((tab) => {
          const active = location.pathname === tab.path
          return (
            <button
              key={tab.path}
              onClick={() => navigate(tab.path)}
              className="flex flex-col items-center justify-center gap-1 flex-1 h-full transition-colors duration-150"
            >
              {tab.icon(active)}
              <span
                className="text-[10px] font-medium"
                style={{ color: active ? '#5e9eff' : '#6e6e73' }}
              >
                {tab.label}
              </span>
            </button>
          )
        })}
      </div>
    </nav>
  )
}
