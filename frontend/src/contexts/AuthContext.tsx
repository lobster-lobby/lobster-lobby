/* eslint-disable react-refresh/only-export-components */
import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  useRef,
  type ReactNode,
} from 'react'
import { useNavigate } from 'react-router-dom'

interface User {
  id: string
  username: string
  email?: string
  role?: string
}

interface AuthContextValue {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (credentials: LoginCredentials) => Promise<void>
  register: (data: RegisterData) => Promise<void>
  logout: () => void
}

interface LoginCredentials {
  identifier: string
  password: string
  rememberMe?: boolean
}

interface RegisterData {
  username: string
  email?: string
  password: string
  accountType: 'human' | 'agent'
}

interface AuthTokenPayload {
  sub?: string
  id?: string
  username: string
  email: string
  exp: number
  role?: string
}

import { getAccessToken, setAccessToken } from './authTokenStore'

const AuthContext = createContext<AuthContextValue | null>(null)

function parseJwt(token: string): AuthTokenPayload | null {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    return JSON.parse(atob(base64))
  } catch {
    return null
  }
}

function userFromPayload(payload: AuthTokenPayload): User {
  return {
    id: payload.sub || payload.id || '',
    username: payload.username,
    email: payload.email,
    role: payload.role,
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const refreshTimerRef = useRef<ReturnType<typeof setTimeout>>(undefined)
  const navigateRef = useRef<ReturnType<typeof useNavigate>>(undefined)

  const navigate = useNavigate()
  navigateRef.current = navigate

  const clearAuth = useCallback(() => {
    setAccessToken(null)
    setUser(null)
    if (refreshTimerRef.current) {
      clearTimeout(refreshTimerRef.current)
    }
  }, [])

  const scheduleRefresh = useCallback(
    (token: string) => {
      const payload = parseJwt(token)
      if (!payload?.exp) return

      // Refresh 60 seconds before expiry
      const msUntilExpiry = payload.exp * 1000 - Date.now()
      const refreshIn = Math.max(msUntilExpiry - 60_000, 5_000)

      if (refreshTimerRef.current) {
        clearTimeout(refreshTimerRef.current)
      }

      refreshTimerRef.current = setTimeout(async () => {
        try {
          const res = await fetch('/api/auth/refresh', {
            method: 'POST',
            credentials: 'include',
          })
          if (!res.ok) throw new Error('Refresh failed')
          const data = await res.json()
          setAccessToken(data.token)
          const newPayload = parseJwt(data.token)
          if (newPayload) {
            setUser(userFromPayload(newPayload))
            scheduleRefresh(data.token)
          }
        } catch {
          clearAuth()
          navigateRef.current?.('/login')
        }
      }, refreshIn)
    },
    [clearAuth]
  )

  const setAuth = useCallback(
    (token: string) => {
      setAccessToken(token)
      const payload = parseJwt(token)
      if (payload) {
        setUser(userFromPayload(payload))
        scheduleRefresh(token)
      }
    },
    [scheduleRefresh]
  )

  // Try to restore session on mount via refresh token cookie
  useEffect(() => {
    let cancelled = false
    async function tryRefresh() {
      try {
        const res = await fetch('/api/auth/refresh', {
          method: 'POST',
          credentials: 'include',
        })
        if (res.ok) {
          const data = await res.json()
          if (!cancelled) {
            setAuth(data.token)
          }
        }
      } catch {
        // No valid session — that's fine
      } finally {
        if (!cancelled) {
          setIsLoading(false)
        }
      }
    }
    tryRefresh()
    return () => {
      cancelled = true
    }
  }, [setAuth])

  // Cleanup timer on unmount
  useEffect(() => {
    return () => {
      if (refreshTimerRef.current) {
        clearTimeout(refreshTimerRef.current)
      }
    }
  }, [])

  // Global 401 interceptor
  useEffect(() => {
    const originalFetch = window.fetch
    window.fetch = async (...args) => {
      const res = await originalFetch(...args)
      const url = typeof args[0] === 'string' ? args[0] : args[0] instanceof Request ? args[0].url : ''
      if (res.status === 401 && getAccessToken() && !url.includes('/api/auth/refresh')) {
        clearAuth()
        const currentPath = window.location.pathname
        if (currentPath !== '/login' && currentPath !== '/register') {
          navigateRef.current?.(
            `/login?redirect=${encodeURIComponent(currentPath)}`
          )
        }
      }
      return res
    }
    return () => {
      window.fetch = originalFetch
    }
  }, [clearAuth])

  const login = useCallback(
    async (credentials: LoginCredentials) => {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          identifier: credentials.identifier,
          password: credentials.password,
          rememberMe: credentials.rememberMe,
        }),
      })

      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        throw new Error(data.message || 'Invalid credentials')
      }

      const data = await res.json()
      setAuth(data.token)
    },
    [setAuth]
  )

  const register = useCallback(
    async (data: RegisterData) => {
      const res = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(data),
      })

      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        throw new Error(body.message || 'Registration failed')
      }

      const body = await res.json()
      setAuth(body.token)
    },
    [setAuth]
  )

  const logout = useCallback(async () => {
    try {
      await fetch('/api/auth/logout', {
        method: 'POST',
        credentials: 'include',
      })
    } catch {
      // Logout endpoint failure is non-critical
    }
    clearAuth()
    navigateRef.current?.('/')
  }, [clearAuth])

  return (
    <AuthContext.Provider value={{ user, isAuthenticated: !!user, isLoading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return ctx
}
