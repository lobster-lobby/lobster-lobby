import { useEffect, useRef } from 'react'

/**
 * Adds a CSS class when the element scrolls into view.
 * Uses IntersectionObserver — no JS animation libraries.
 */
export function useScrollReveal<T extends HTMLElement>(
  className = 'revealed',
  options?: IntersectionObserverInit,
) {
  const ref = useRef<T>(null)

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
      { threshold: 0.15, ...options },
    )

    observer.observe(el)
    return () => observer.disconnect()
  }, [className, options])

  return ref
}
