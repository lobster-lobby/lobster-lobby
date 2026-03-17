import { Link } from 'react-router-dom'

export default function Register() {
  return (
    <div>
      <h1>Register</h1>
      <p>Create your account.</p>
      <p>
        Already have an account? <Link to="/login">Login</Link>
      </p>
    </div>
  )
}
