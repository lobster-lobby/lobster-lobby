import { useState, useEffect } from 'react'

interface User {
  id: string
  username: string
  email: string
}

interface AuthState {
  isAuthenticated: boolean
  user: User | null
  logout: () => void
}

function parseJwt(token: string): User | null {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const payload = JSON.parse(atob(base64))

    // Check expiry
    if (payload.exp && payload.exp * 1000 < Date.now()) {
      localStorage.removeItem('ll_token')
      return null
    }

    return {
      id: payload.sub || payload.id,
      username: payload.username,
      email: payload.email,
    }
  } catch {
    return null
  }
}

function getUser(): User | null {
  const token = localStorage.getItem('ll_token')
  return token ? parseJwt(token) : null
}

// Simple external store for cross-component reactivity
const listeners = new Set<() => void>()

export function notifyAuthChange() {
  listeners.forEach((fn) => fn())
}

export function useAuth(): AuthState {
  const [user, setUser] = useState<User | null>(getUser)

  useEffect(() => {
    const update = () => setUser(getUser())
    listeners.add(update)

    // Also listen for storage events (cross-tab)
    window.addEventListener('storage', update)

    return () => {
      listeners.delete(update)
      window.removeEventListener('storage', update)
    }
  }, [])

  const logout = () => {
    localStorage.removeItem('ll_token')
    notifyAuthChange()
    window.location.href = '/'
  }

  return {
    isAuthenticated: !!user,
    user,
    logout,
  }
}
