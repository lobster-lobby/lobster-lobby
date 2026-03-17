import type { SVGAttributes } from 'react'

export interface LogoProps extends SVGAttributes<SVGSVGElement> {
  width?: number | string
  height?: number | string
  showText?: boolean
}

export function Logo({
  width = 180,
  height = 40,
  showText = true,
  ...props
}: LogoProps) {
  if (!showText) {
    return (
      <svg
        viewBox="0 0 40 40"
        width={typeof width === 'number' ? width * 0.22 : 40}
        height={height}
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        role="img"
        aria-label="Lobster Lobby logo"
        {...props}
      >
        <LogoIcon />
      </svg>
    )
  }

  return (
    <svg
      viewBox="0 0 180 40"
      width={width}
      height={height}
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      role="img"
      aria-label="Lobster Lobby logo"
      {...props}
    >
      <g transform="translate(0, 0)">
        <LogoIcon />
      </g>
      <g transform="translate(46, 0)">
        <text
          x="0"
          y="26"
          fontFamily="system-ui, -apple-system, sans-serif"
          fontSize="18"
          fontWeight="700"
          fill="var(--ll-text)"
        >
          Lobster
        </text>
        <text
          x="68"
          y="26"
          fontFamily="system-ui, -apple-system, sans-serif"
          fontSize="18"
          fontWeight="400"
          fill="var(--ll-text-secondary)"
        >
          Lobby
        </text>
      </g>
    </svg>
  )
}

function LogoIcon() {
  return (
    <g>
      {/* Lobster body */}
      <ellipse cx="20" cy="24" rx="8" ry="10" fill="var(--ll-primary)" />

      {/* Head */}
      <circle cx="20" cy="12" r="6" fill="var(--ll-primary)" />

      {/* Eyes */}
      <circle cx="18" cy="11" r="1.5" fill="white" />
      <circle cx="22" cy="11" r="1.5" fill="white" />
      <circle cx="18.3" cy="11" r="0.7" fill="var(--ll-text)" />
      <circle cx="22.3" cy="11" r="0.7" fill="var(--ll-text)" />

      {/* Antennae */}
      <path
        d="M17 7 Q14 3 11 2"
        stroke="var(--ll-primary)"
        strokeWidth="1.5"
        strokeLinecap="round"
        fill="none"
      />
      <path
        d="M23 7 Q26 3 29 2"
        stroke="var(--ll-primary)"
        strokeWidth="1.5"
        strokeLinecap="round"
        fill="none"
      />

      {/* Claws */}
      <g transform="translate(6, 16)">
        <ellipse cx="0" cy="4" rx="3" ry="2" fill="var(--ll-primary)" />
        <path d="M-3 2 Q-5 0 -4 -1.5 Q-2 -3 0 -1 L1 2" fill="var(--ll-primary)" />
        <path d="M3 2 Q5 0 4 -1.5 Q2 -3 0 -1 L-1 2" fill="var(--ll-primary)" />
      </g>
      <path
        d="M12 20 Q8 18 6 20"
        stroke="var(--ll-primary)"
        strokeWidth="3"
        strokeLinecap="round"
        fill="none"
      />

      <g transform="translate(34, 16)">
        <ellipse cx="0" cy="4" rx="3" ry="2" fill="var(--ll-primary)" />
        <path d="M-3 2 Q-5 0 -4 -1.5 Q-2 -3 0 -1 L1 2" fill="var(--ll-primary)" />
        <path d="M3 2 Q5 0 4 -1.5 Q2 -3 0 -1 L-1 2" fill="var(--ll-primary)" />
      </g>
      <path
        d="M28 20 Q32 18 34 20"
        stroke="var(--ll-primary)"
        strokeWidth="3"
        strokeLinecap="round"
        fill="none"
      />

      {/* Tail */}
      <ellipse cx="20" cy="34" rx="5" ry="3" fill="var(--ll-primary)" opacity="0.8" />
      <path d="M16 36 Q20 40 24 36" fill="var(--ll-primary)" opacity="0.6" />

      {/* Smile */}
      <path
        d="M18 14 Q20 16 22 14"
        stroke="var(--ll-text)"
        strokeWidth="1"
        strokeLinecap="round"
        fill="none"
      />
    </g>
  )
}

export function LogoMark(props: Omit<LogoProps, 'showText'>) {
  return <Logo showText={false} {...props} />
}
