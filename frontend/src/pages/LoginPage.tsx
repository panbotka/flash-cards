import { useState, useRef, useEffect } from 'react'
import type { FormEvent } from 'react'
import { authLogin } from '../api/client'

interface LoginPageProps {
  onSuccess: () => void
}

export function LoginPage({ onSuccess }: LoginPageProps) {
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    inputRef.current?.focus()
  }, [])

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      await authLogin(password)
      onSuccess()
    } catch {
      setError('Incorrect password')
      setPassword('')
      inputRef.current?.focus()
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-[#0a0a0a] px-6">
      <div className="w-full max-w-sm">
        <h1 className="text-4xl font-bold text-white text-center mb-2 tracking-tight">
          Flash Cards
        </h1>
        <p className="text-[#6e6e73] text-center mb-12 text-lg">
          Czech &mdash; English
        </p>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <input
            ref={inputRef}
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Password"
            autoComplete="current-password"
            className="w-full px-4 py-3.5 rounded-xl bg-[#1a1a1a] border border-[#2a2a2a] text-white text-center text-lg placeholder-[#6e6e73] outline-none focus:border-[#5e9eff] transition-colors duration-200"
          />

          {error && (
            <p className="text-[#ff453a] text-sm text-center">{error}</p>
          )}

          <button
            type="submit"
            disabled={loading || !password}
            className="w-full py-3.5 rounded-xl bg-[#5e9eff] text-white font-semibold text-base transition-all duration-150 active:scale-[0.98] hover:bg-[#4a8af0] disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? 'Signing in\u2026' : 'Sign In'}
          </button>
        </form>
      </div>
    </div>
  )
}
