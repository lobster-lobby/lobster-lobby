import { useEffect, useRef } from 'react'

/**
 * Adds a CSS class when the element scrolls into view.
 * Uses IntersectionObserver — no JS animation libraries.
 *
 * Options are captured in a ref so callers can pass inline objects
 * without causing the effect to re-run on every render.
 */
export function useScrollReveal<T extends HTMLElement>(
  className = 'revealed',
  options?: IntersectionObserverInit,
) {
  const ref = useRef<T>(null)
  const optionsRef = useRef(options)

  useEffect(() => {
    const el = ref.current
    if (!el) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          el.classList.add(className)
          observer.unobserve(el)
        }
      },
      { threshold: 0.15, ...optionsRef.current },
    )

    observer.observe(el)
    return () => observer.disconnect()
  }, [className])

  return ref
}
