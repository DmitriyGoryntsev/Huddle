import { Outlet } from 'react-router-dom'
import { Link, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'

export default function Layout() {
  const location = useLocation()
  const clearAuth = useAuthStore((s) => s.clearAuth)

  return (
    <div className="min-h-screen flex flex-col">
      <header className="bg-white border-b border-stone-200 sticky top-0 z-20">
        <div className="max-w-7xl mx-auto px-4 h-14 flex items-center justify-between">
          <Link to="/" className="text-xl font-bold text-huddle-700">
            Huddle
          </Link>
          <nav className="flex items-center gap-4">
            <Link
              to="/"
              className={`text-sm font-medium ${location.pathname === '/' ? 'text-huddle-600' : 'text-stone-600 hover:text-stone-900'}`}
            >
              Карта
            </Link>
            <Link
              to="/my-events"
              className={`text-sm font-medium ${location.pathname === '/my-events' ? 'text-huddle-600' : 'text-stone-600 hover:text-stone-900'}`}
            >
              Мои события
            </Link>
            <button
              onClick={clearAuth}
              className="text-sm text-stone-500 hover:text-stone-700"
            >
              Выйти
            </button>
          </nav>
        </div>
      </header>

      <main className="flex-1">
        <Outlet />
      </main>
    </div>
  )
}
