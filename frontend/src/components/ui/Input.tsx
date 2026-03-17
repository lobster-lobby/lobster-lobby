import type { InputHTMLAttributes, TextareaHTMLAttributes } from 'react'
import { forwardRef, useId } from 'react'
import styles from './Input.module.css'

interface BaseInputProps {
  label?: string
  error?: string
  hint?: string
}

export interface InputProps
  extends BaseInputProps,
    Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  type?: 'text' | 'email' | 'password'
}

export interface TextareaProps
  extends BaseInputProps,
    TextareaHTMLAttributes<HTMLTextAreaElement> {}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, hint, type = 'text', className = '', id, ...props }, ref) => {
    const generatedId = useId()
    const inputId = id || `input-${generatedId}`
    const hasError = Boolean(error)

    return (
      <div className={[styles.wrapper, className].filter(Boolean).join(' ')}>
        {label && (
          <label htmlFor={inputId} className={styles.label}>
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={inputId}
          type={type}
          className={[styles.input, hasError && styles.inputError].filter(Boolean).join(' ')}
          aria-invalid={hasError}
          aria-describedby={error ? `${inputId}-error` : hint ? `${inputId}-hint` : undefined}
          {...props}
        />
        {error && (
          <span id={`${inputId}-error`} className={styles.error}>
            {error}
          </span>
        )}
        {hint && !error && (
          <span id={`${inputId}-hint`} className={styles.hint}>
            {hint}
          </span>
        )}
      </div>
    )
  }
)

Input.displayName = 'Input'

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ label, error, hint, className = '', id, ...props }, ref) => {
    const generatedId = useId()
    const inputId = id || `textarea-${generatedId}`
    const hasError = Boolean(error)

    return (
      <div className={[styles.wrapper, className].filter(Boolean).join(' ')}>
        {label && (
          <label htmlFor={inputId} className={styles.label}>
            {label}
          </label>
        )}
        <textarea
          ref={ref}
          id={inputId}
          className={[styles.textarea, hasError && styles.inputError].filter(Boolean).join(' ')}
          aria-invalid={hasError}
          aria-describedby={error ? `${inputId}-error` : hint ? `${inputId}-hint` : undefined}
          {...props}
        />
        {error && (
          <span id={`${inputId}-error`} className={styles.error}>
            {error}
          </span>
        )}
        {hint && !error && (
          <span id={`${inputId}-hint`} className={styles.hint}>
            {hint}
          </span>
        )}
      </div>
    )
  }
)

Textarea.displayName = 'Textarea'
