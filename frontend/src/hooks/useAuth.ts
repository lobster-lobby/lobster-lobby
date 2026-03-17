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
    return {
      id: payload.sub || payload.id,
      username: payload.username,
      email: payload.email,
    }
  } catch {
    return null
  }
}

export function useAuth(): AuthState {
  const token = localStorage.getItem('ll_token')
  const user = token ? parseJwt(token) : null

  const logout = () => {
    localStorage.removeItem('ll_token')
    window.location.href = '/'
  }

  return {
    isAuthenticated: !!user,
    user,
    logout,
  }
}
