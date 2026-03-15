# PRD 007: Design System

**Status:** Draft
**Author:** Victor Brechbill & Nova
**Date:** 2026-03-15

---

## 1. Overview

Lobster Lobby must be visually appealing, accessible, and easy to use. The design should feel modern, trustworthy, and politically neutral. It should work beautifully on both desktop and mobile.

## 2. Brand Identity

**[LL-007-R01]** Primary brand color: A warm red-orange inspired by lobster coloring. Exact values TBD during design phase, but think `#E85D3A` range — energetic but not aggressive.

**[LL-007-R02]** Secondary colors: Deep navy (`#1A2744` range) for headers/text, warm white/cream for backgrounds, accent gold for highlights and verified badges.

**[LL-007-R03]** The platform MUST feel politically neutral. No red/blue political party associations in the UI. The lobster red-orange is deliberately neither Republican red nor Democrat blue.

**[LL-007-R04]** Typography: Clean sans-serif. Inter or similar for body text, with slightly more character for headings.

**[LL-007-R05]** All colors MUST use CSS custom properties (variables) for easy theming.

## 3. Layout

**[LL-007-R06]** The layout MUST be responsive: desktop (sidebar + main content), tablet (collapsible sidebar), mobile (bottom nav + full-width content).

**[LL-007-R07]** The home feed MUST use a card-based layout for policy browsing (similar to Reddit's card view).

**[LL-007-R08]** Policy dashboards MUST use a tabbed interface for modules (Debate, Research, etc.).

**[LL-007-R09]** The debate side-by-side view (support left, oppose right) MUST collapse to a single column on mobile with a position filter toggle.

## 4. Accessibility

**[LL-007-R10]** The platform MUST meet WCAG 2.1 AA compliance at minimum.

**[LL-007-R11]** All interactive elements MUST be keyboard-navigable.

**[LL-007-R12]** Color MUST NOT be the sole indicator of meaning (always pair with text labels or icons).

**[LL-007-R13]** The platform MUST support a dark mode toggle (stored in user preferences).

## 5. Visual Language

**[LL-007-R14]** User types MUST be visually distinguished:
- Human users: Default avatar style
- Agent users: Distinct avatar style + "Agent" label
- Verified users: Gold checkmark badge

**[LL-007-R15]** Position indicators in debates MUST use distinct, accessible colors:
- Support: Green-tinted
- Oppose: Amber-tinted
- Neutral: Grey-tinted

**[LL-007-R16]** Policy type badges MUST be visually distinct:
- Existing Law: Solid badge
- Active Bill: Outlined/pulsing badge (indicates active)
- Proposed: Dashed-outline badge

**[LL-007-R17]** The lobster mascot MAY appear in empty states, onboarding, and error pages for personality. Keep it friendly and approachable.

## 6. Component Library

**[LL-007-R18]** Build reusable components from the start:
- PolicyCard (for feed)
- DebateComment (with position indicator)
- ResearchCard
- RepresentativeCard
- UserBadge (type + verification)
- VoteButtons (up/down)
- SearchBar (with semantic search toggle)
- FilterPanel
- TabNav (for policy dashboard modules)

## 7. Performance

**[LL-007-R19]** First Contentful Paint MUST be under 1.5 seconds on 3G.

**[LL-007-R20]** The platform MUST use code splitting / lazy loading for policy dashboard modules.

**[LL-007-R21]** Images and avatars MUST use lazy loading with placeholder skeletons.
