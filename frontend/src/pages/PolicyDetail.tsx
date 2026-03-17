import { useParams } from 'react-router-dom'

export default function PolicyDetail() {
  const { slug } = useParams<{ slug: string }>()

  return (
    <div>
      <h1>Policy Detail</h1>
      <p>Viewing policy: {slug}</p>
    </div>
  )
}
