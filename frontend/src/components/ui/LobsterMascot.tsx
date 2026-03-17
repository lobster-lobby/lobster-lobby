import type { SVGAttributes } from 'react'

export type MascotPose = 'waving' | 'thinking' | 'celebrating' | 'reading' | 'debating'

export interface LobsterMascotProps extends SVGAttributes<SVGSVGElement> {
  pose?: MascotPose
  width?: number | string
  height?: number | string
}

export function LobsterMascot({
  pose = 'waving',
  width = 200,
  height = 200,
  ...props
}: LobsterMascotProps) {
  return (
    <svg
      viewBox="0 0 100 100"
      width={width}
      height={height}
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      role="img"
      aria-label={`Lobster mascot ${pose}`}
      {...props}
    >
      {/* Base lobster body - shared across all poses */}
      <g>
        {/* Tail segments */}
        <ellipse cx="50" cy="78" rx="10" ry="6" fill="var(--ll-primary)" opacity="0.8" />
        <ellipse cx="50" cy="84" rx="8" ry="5" fill="var(--ll-primary)" opacity="0.7" />
        <path
          d="M42 88 Q50 98 58 88"
          fill="var(--ll-primary)"
          opacity="0.6"
        />

        {/* Main body */}
        <ellipse cx="50" cy="55" rx="18" ry="25" fill="var(--ll-primary)" />

        {/* Body highlight */}
        <ellipse cx="46" cy="50" rx="8" ry="12" fill="var(--ll-primary-hover)" opacity="0.3" />

        {/* Head */}
        <circle cx="50" cy="28" r="14" fill="var(--ll-primary)" />

        {/* Face highlight */}
        <circle cx="47" cy="26" r="6" fill="var(--ll-primary-hover)" opacity="0.2" />

        {/* Eyes */}
        <circle cx="44" cy="24" r="4" fill="white" />
        <circle cx="56" cy="24" r="4" fill="white" />
        <circle cx="45" cy="24" r="2" fill="var(--ll-text)" />
        <circle cx="57" cy="24" r="2" fill="var(--ll-text)" />
        <circle cx="45.5" cy="23.5" r="0.8" fill="white" />
        <circle cx="57.5" cy="23.5" r="0.8" fill="white" />

        {/* Antennae */}
        <path
          d="M42 16 Q38 8 32 4"
          stroke="var(--ll-primary)"
          strokeWidth="2"
          strokeLinecap="round"
          fill="none"
        />
        <path
          d="M58 16 Q62 8 68 4"
          stroke="var(--ll-primary)"
          strokeWidth="2"
          strokeLinecap="round"
          fill="none"
        />
        <circle cx="32" cy="4" r="2" fill="var(--ll-primary)" />
        <circle cx="68" cy="4" r="2" fill="var(--ll-primary)" />

        {/* Legs */}
        <g stroke="var(--ll-primary)" strokeWidth="2.5" strokeLinecap="round">
          <path d="M35 58 L26 65" />
          <path d="M35 65 L26 72" />
          <path d="M35 72 L28 80" />
          <path d="M65 58 L74 65" />
          <path d="M65 65 L74 72" />
          <path d="M65 72 L72 80" />
        </g>
      </g>

      {/* Pose-specific elements */}
      {pose === 'waving' && <WavingPose />}
      {pose === 'thinking' && <ThinkingPose />}
      {pose === 'celebrating' && <CelebratingPose />}
      {pose === 'reading' && <ReadingPose />}
      {pose === 'debating' && <DebatingPose />}

      {/* Mouth - varies by pose */}
      <PoseMouth pose={pose} />
    </svg>
  )
}

function PoseMouth({ pose }: { pose: MascotPose }) {
  switch (pose) {
    case 'celebrating':
      return (
        <path
          d="M44 32 Q50 38 56 32"
          stroke="var(--ll-text)"
          strokeWidth="2"
          strokeLinecap="round"
          fill="none"
        />
      )
    case 'thinking':
      return (
        <ellipse cx="50" cy="33" rx="3" ry="2" fill="var(--ll-text)" opacity="0.6" />
      )
    case 'debating':
      return (
        <path
          d="M45 32 L55 32 Q55 36 50 36 Q45 36 45 32"
          fill="var(--ll-text)"
          opacity="0.7"
        />
      )
    default:
      return (
        <path
          d="M46 32 Q50 35 54 32"
          stroke="var(--ll-text)"
          strokeWidth="1.5"
          strokeLinecap="round"
          fill="none"
        />
      )
  }
}

function WavingPose() {
  return (
    <g>
      {/* Left claw - down */}
      <g transform="translate(20, 40)">
        <ellipse cx="0" cy="8" rx="6" ry="4" fill="var(--ll-primary)" />
        <path
          d="M-6 4 Q-12 0 -8 -4 Q-4 -8 0 -4 L2 4"
          fill="var(--ll-primary)"
        />
        <path
          d="M6 4 Q12 0 8 -4 Q4 -8 0 -4 L-2 4"
          fill="var(--ll-primary)"
        />
      </g>
      {/* Arm to left claw */}
      <path
        d="M35 45 Q28 42 20 48"
        stroke="var(--ll-primary)"
        strokeWidth="6"
        strokeLinecap="round"
        fill="none"
      />

      {/* Right claw - waving up */}
      <g transform="translate(82, 22) rotate(30)">
        <ellipse cx="0" cy="8" rx="6" ry="4" fill="var(--ll-primary)" />
        <path
          d="M-6 4 Q-12 0 -8 -4 Q-4 -8 0 -4 L2 4"
          fill="var(--ll-primary)"
        />
        <path
          d="M6 4 Q12 0 8 -4 Q4 -8 0 -4 L-2 4"
          fill="var(--ll-primary)"
        />
      </g>
      {/* Arm to right claw */}
      <path
        d="M65 45 Q75 35 78 25"
        stroke="var(--ll-primary)"
        strokeWidth="6"
        strokeLinecap="round"
        fill="none"
      />

      {/* Motion lines for waving */}
      <g stroke="var(--ll-primary)" strokeWidth="1.5" opacity="0.4" strokeLinecap="round">
        <path d="M88 18 L94 14" />
        <path d="M90 24 L96 22" />
        <path d="M88 30 L94 32" />
      </g>
    </g>
  )
}

function ThinkingPose() {
  return (
    <g>
      {/* Left claw - resting */}
      <g transform="translate(18, 45)">
        <ellipse cx="0" cy="8" rx="6" ry="4" fill="var(--ll-primary)" />
        <path
          d="M-6 4 Q-12 0 -8 -4 Q-4 -8 0 -4 L2 4"
          fill="var(--ll-primary)"
        />
        <path
          d="M6 4 Q12 0 8 -4 Q4 -8 0 -4 L-2 4"
          fill="var(--ll-primary)"
        />
      </g>
      <path
        d="M35 48 Q26 46 18 52"
        stroke="var(--ll-primary)"
        strokeWidth="6"
        strokeLinecap="round"
        fill="none"
      />

      {/* Right claw - near chin, thinking pose */}
      <g transform="translate(68, 32) rotate(-20)">
        <ellipse cx="0" cy="8" rx="6" ry="4" fill="var(--ll-primary)" />
        <path
          d="M-6 4 Q-12 0 -8 -4 Q-4 -8 0 -4 L2 4"
          fill="var(--ll-primary)"
        />
        <path
          d="M6 4 Q12 0 8 -4 Q4 -8 0 -4 L-2 4"
          fill="var(--ll-primary)"
        />
      </g>
      <path
        d="M65 45 Q68 38 66 34"
        stroke="var(--ll-primary)"
        strokeWidth="6"
        strokeLinecap="round"
        fill="none"
      />

      {/* Thought bubbles */}
      <circle cx="78" cy="12" r="3" fill="var(--ll-text-muted)" opacity="0.3" />
      <circle cx="84" cy="6" r="4" fill="var(--ll-text-muted)" opacity="0.3" />
      <circle cx="92" cy="2" r="5" fill="var(--ll-text-muted)" opacity="0.3" />
    </g>
  )
}

function CelebratingPose() {
  return (
    <g>
      {/* Left claw - raised up */}
      <g transform="translate(16, 18) rotate(-30)">
        <ellipse cx="0" cy="8" rx="6" ry="4" fill="var(--ll-primary)" />
        <path
          d="M-6 4 Q-12 0 -8 -4 Q-4 -8 0 -4 L2 4"
          fill="var(--ll-primary)"
        />
        <path
          d="M6 4 Q12 0 8 -4 Q4 -8 0 -4 L-2 4"
          fill="var(--ll-primary)"
        />
      </g>
      <path
        d="M35 45 Q25 32 20 22"
        stroke="var(--ll-primary)"
        strokeWidth="6"
        strokeLinecap="round"
        fill="none"
      />

      {/* Right claw - raised up */}
      <g transform="translate(84, 18) rotate(30)">
        <ellipse cx="0" cy="8" rx="6" ry="4" fill="var(--ll-primary)" />
        <path
          d="M-6 4 Q-12 0 -8 -4 Q-4 -8 0 -4 L2 4"
          fill="var(--ll-primary)"
        />
        <path
          d="M6 4 Q12 0 8 -4 Q4 -8 0 -4 L-2 4"
          fill="var(--ll-primary)"
        />
      </g>
      <path
        d="M65 45 Q75 32 80 22"
        stroke="var(--ll-primary)"
        strokeWidth="6"
        strokeLinecap="round"
        fill="none"
      />

      {/* Celebration sparkles */}
      <g fill="var(--ll-support)" opacity="0.8">
        <circle cx="8" cy="8" r="2" />
        <circle cx="92" cy="8" r="2" />
        <circle cx="50" cy="2" r="1.5" />
      </g>
      <g stroke="var(--ll-support)" strokeWidth="1.5" opacity="0.6" strokeLinecap="round">
        <path d="M6 14 L2 18" />
        <path d="M10 10 L6 6" />
        <path d="M94 14 L98 18" />
        <path d="M90 10 L94 6" />
      </g>
    </g>
  )
}

function ReadingPose() {
  return (
    <g>
      {/* Book/document */}
      <g transform="translate(30, 48)">
        <rect x="0" y="0" width="40" height="28" rx="2" fill="var(--ll-bg-card)" stroke="var(--ll-border)" strokeWidth="1" />
        <line x1="20" y1="2" x2="20" y2="26" stroke="var(--ll-border)" strokeWidth="1" />
        <g fill="var(--ll-text-muted)" opacity="0.4">
          <rect x="4" y="6" width="12" height="2" rx="1" />
          <rect x="4" y="11" width="10" height="2" rx="1" />
          <rect x="4" y="16" width="12" height="2" rx="1" />
          <rect x="24" y="6" width="12" height="2" rx="1" />
          <rect x="24" y="11" width="10" height="2" rx="1" />
          <rect x="24" y="16" width="12" height="2" rx="1" />
        </g>
      </g>

      {/* Left claw - holding book */}
      <g transform="translate(24, 58) rotate(10)">
        <ellipse cx="0" cy="4" rx="5" ry="3" fill="var(--ll-primary)" />
        <path
          d="M-4 2 Q-8 0 -6 -3 Q-3 -5 0 -2 L1 2"
          fill="var(--ll-primary)"
        />
        <path
          d="M4 2 Q8 0 6 -3 Q3 -5 0 -2 L-1 2"
          fill="var(--ll-primary)"
        />
      </g>
      <path
        d="M35 48 Q30 52 26 58"
        stroke="var(--ll-primary)"
        strokeWidth="5"
        strokeLinecap="round"
        fill="none"
      />

      {/* Right claw - holding book */}
      <g transform="translate(76, 58) rotate(-10)">
        <ellipse cx="0" cy="4" rx="5" ry="3" fill="var(--ll-primary)" />
        <path
          d="M-4 2 Q-8 0 -6 -3 Q-3 -5 0 -2 L1 2"
          fill="var(--ll-primary)"
        />
        <path
          d="M4 2 Q8 0 6 -3 Q3 -5 0 -2 L-1 2"
          fill="var(--ll-primary)"
        />
      </g>
      <path
        d="M65 48 Q70 52 74 58"
        stroke="var(--ll-primary)"
        strokeWidth="5"
        strokeLinecap="round"
        fill="none"
      />

      {/* Reading glasses */}
      <g stroke="var(--ll-text)" strokeWidth="1" fill="none" opacity="0.6">
        <circle cx="44" cy="24" r="5" />
        <circle cx="56" cy="24" r="5" />
        <path d="M49 24 L51 24" />
        <path d="M39 24 L36 22" />
        <path d="M61 24 L64 22" />
      </g>
    </g>
  )
}

function DebatingPose() {
  return (
    <g>
      {/* Left claw - gesturing */}
      <g transform="translate(12, 32) rotate(-45)">
        <ellipse cx="0" cy="8" rx="6" ry="4" fill="var(--ll-primary)" />
        <path
          d="M-6 4 Q-12 0 -8 -4 Q-4 -8 0 -4 L2 4"
          fill="var(--ll-primary)"
        />
        <path
          d="M6 4 Q12 0 8 -4 Q4 -8 0 -4 L-2 4"
          fill="var(--ll-primary)"
        />
      </g>
      <path
        d="M35 45 Q22 38 16 30"
        stroke="var(--ll-primary)"
        strokeWidth="6"
        strokeLinecap="round"
        fill="none"
      />

      {/* Right claw - pointing */}
      <g transform="translate(88, 38) rotate(10)">
        <ellipse cx="0" cy="6" rx="5" ry="3" fill="var(--ll-primary)" />
        <path
          d="M-5 3 Q-10 0 -7 -3 Q-4 -6 0 -3 L1 3"
          fill="var(--ll-primary)"
        />
        <path
          d="M5 3 Q10 0 7 -3 Q4 -6 0 -3 L-1 3"
          fill="var(--ll-primary)"
        />
      </g>
      <path
        d="M65 45 Q78 42 84 40"
        stroke="var(--ll-primary)"
        strokeWidth="6"
        strokeLinecap="round"
        fill="none"
      />

      {/* Speech bubble */}
      <g>
        <path
          d="M4 10 Q4 2 14 2 L26 2 Q36 2 36 10 L36 18 Q36 26 26 26 L18 26 L12 32 L14 26 L14 26 Q4 26 4 18 Z"
          fill="var(--ll-bg-card)"
          stroke="var(--ll-border)"
          strokeWidth="1"
        />
        <g fill="var(--ll-text-muted)" opacity="0.5">
          <rect x="8" y="8" width="24" height="2" rx="1" />
          <rect x="8" y="14" width="18" height="2" rx="1" />
          <rect x="8" y="20" width="20" height="2" rx="1" />
        </g>
      </g>
    </g>
  )
}
