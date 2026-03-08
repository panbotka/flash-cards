import { useQuery, useQueryClient } from '@tanstack/react-query'
import { BrowserRouter, Routes, Route, useLocation, Navigate } from 'react-router-dom'
import { authCheck } from './api/client'
import { NavBar } from './components/NavBar'
import { LoginPage } from './pages/LoginPage'
import { StudyPage } from './pages/StudyPage'
import { CardsPage } from './pages/CardsPage'
import { ImportPage } from './pages/ImportPage'
import { StatsPage } from './pages/StatsPage'
import { CardHistoryPage } from './pages/CardHistoryPage'

function LoadingScreen() {
  return (
    <div className="flex items-center justify-center min-h-screen bg-[#0a0a0a]">
      <div className="w-6 h-6 border-2 border-[#2a2a2a] border-t-[#5e9eff] rounded-full animate-spin" />
    </div>
  )
}

function AppLayout() {
  const location = useLocation()
  const showNav = location.pathname !== '/login'

  return (
    <>
      <Routes>
        <Route path="/" element={<StudyPage />} />
        <Route path="/cards" element={<CardsPage />} />
        <Route path="/cards/:id" element={<CardHistoryPage />} />
        <Route path="/import" element={<ImportPage />} />
        <Route path="/stats" element={<StatsPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
      {showNav && <NavBar />}
    </>
  )
}

export function App() {
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['auth'],
    queryFn: authCheck,
    retry: false,
    staleTime: 60_000,
  })

  if (isLoading) {
    return <LoadingScreen />
  }

  const needsAuth = data?.authRequired && !data?.authenticated

  if (needsAuth) {
    return (
      <LoginPage
        onSuccess={() => {
          queryClient.invalidateQueries({ queryKey: ['auth'] })
        }}
      />
    )
  }

  return (
    <BrowserRouter>
      <AppLayout />
    </BrowserRouter>
  )
}
