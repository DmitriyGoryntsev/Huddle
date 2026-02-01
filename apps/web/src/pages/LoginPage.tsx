import { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { authApi } from '@/lib/api'
import { useAuthStore } from '@/stores/authStore'

export default function LoginPage() {
  const location = useLocation()
  const message = (location.state as { message?: string })?.message
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const setTokens = useAuthStore((s) => s.setTokens)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const tokens = await authApi.login(email, password)
      setTokens(tokens)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка входа')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-huddle-50 to-huddle-100 px-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-huddle-700">Huddle</h1>
          <p className="text-stone-600 mt-1">Найди компанию для игры</p>
        </div>

        <form onSubmit={handleSubmit} className="bg-white rounded-2xl shadow-xl p-6 space-y-4">
          <h2 className="text-xl font-semibold text-stone-800">Вход</h2>

          {message && (
            <div className="p-3 rounded-lg bg-huddle-50 text-huddle-700 text-sm">{message}</div>
          )}
          {error && (
            <div className="p-3 rounded-lg bg-red-50 text-red-700 text-sm">{error}</div>
          )}

          <div>
            <label className="block text-sm font-medium text-stone-700 mb-1">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-2 rounded-lg border border-stone-300 focus:ring-2 focus:ring-huddle-500 focus:border-transparent"
              required
              autoComplete="email"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-stone-700 mb-1">Пароль</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-2 rounded-lg border border-stone-300 focus:ring-2 focus:ring-huddle-500 focus:border-transparent"
              required
              autoComplete="current-password"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full py-2.5 rounded-lg bg-huddle-600 text-white font-medium hover:bg-huddle-700 disabled:opacity-50"
          >
            {loading ? 'Вход...' : 'Войти'}
          </button>

          <p className="text-center text-sm text-stone-600">
            Нет аккаунта?{' '}
            <Link to="/register" className="text-huddle-600 font-medium hover:underline">
              Регистрация
            </Link>
          </p>
        </form>
      </div>
    </div>
  )
}
