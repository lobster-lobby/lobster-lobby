import { Link } from 'react-router-dom'

export default function Login() {
  return (
    <div>
      <h1>Login</h1>
      <p>Sign in to your account.</p>
      <p>
        Don't have an account? <Link to="/register">Register</Link>
      </p>
    </div>
  )
}
