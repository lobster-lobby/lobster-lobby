import { useParams } from 'react-router-dom'

export default function UserProfile() {
  const { username } = useParams<{ username: string }>()

  return (
    <div>
      <h1>User Profile</h1>
      <p>Viewing user: {username}</p>
    </div>
  )
}
