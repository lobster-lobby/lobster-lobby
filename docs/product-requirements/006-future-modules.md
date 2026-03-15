# PRD 006: Future Modules (Phase 2-4)

**Status:** Draft
**Author:** Victor Brechbill & Nova
**Date:** 2026-03-15

---

This document captures the design for modules planned for Phase 2 and beyond.

## 1. Policy Draft Module (Phase 2)

Collaboratively draft legislative text that represents the community's position.

**[LL-006-R01]** Each policy MAY have one or more community drafts of legislative text.

**[LL-006-R02]** Drafts MUST support versioning — each edit creates a new version with diff tracking.

**[LL-006-R03]** Drafts MUST support inline comments on specific sections of the text.

**[LL-006-R04]** The community MAY vote on draft versions to signal which best represents their position.

**[LL-006-R05]** For policies of type `existing_law` or `active_bill`, drafts SHOULD be presented as amendments (additions, removals, modifications to the original text with redline formatting).

## 2. Amendment Tracker (Phase 2)

Track how active legislation evolves through the legislative process.

**[LL-006-R06]** For `active_bill` policies, the platform SHOULD auto-track amendments, committee markups, and status changes from Congress.gov API data.

**[LL-006-R07]** Each amendment MUST show: sponsor, date, summary, and full text.

**[LL-006-R08]** Users MUST receive notifications when tracked policies receive amendments or status changes.

## 3. Impact Analysis Module (Phase 3)

Crowdsource economic, social, and environmental projections.

**[LL-006-R09]** Impact analyses MUST declare a type: `economic`, `social`, `environmental`, `legal`.

**[LL-006-R10]** Analyses MUST include methodology description and source citations.

**[LL-006-R11]** Analyses MAY include structured projections: metric name, baseline value, projected value, timeframe.

**[LL-006-R12]** Analyses are subject to fact-checking and community voting.

## 4. Fact-Check System (Phase 3)

Community-driven verification of claims made anywhere on the platform.

**[LL-006-R13]** Users MAY submit fact-checks against: debate comments, research items, impact analyses, or poll claims.

**[LL-006-R14]** Fact-checks MUST include: the specific claim being checked, a verdict (`true`, `mostly_true`, `mixed`, `mostly_false`, `false`), and supporting evidence with sources.

**[LL-006-R15]** Fact-checks MUST be endorsable by other users (agree/disagree with the verdict).

**[LL-006-R16]** Content with disputed fact-checks MUST display a visual indicator linking to the fact-check.

## 5. Coalition Builder (Phase 3)

Connect users and organizations who support the same policy positions.

**[LL-006-R17]** Each policy MAY have support and oppose coalitions.

**[LL-006-R18]** Coalitions MUST show: member count, organizations involved, shared actions taken.

**[LL-006-R19]** The platform SHOULD surface "coalition overlap" — users who support Policy A also tend to support Policy B.

## 6. Public Comment Portal (Phase 3)

Help users participate in federal rulemaking via regulations.gov.

**[LL-006-R20]** The platform SHOULD track open comment periods on regulations.gov for policies tracked on the platform.

**[LL-006-R21]** Users SHOULD be able to draft public comments collaboratively and submit them to regulations.gov.

**[LL-006-R22]** The platform MUST link to the official regulations.gov page for each open comment period.

## 7. Lobbying Fundraiser Module (Phase 4 — Pending Legal Review)

Each policy has its own GoFundMe-style fundraiser to fund professional lobbying.

**[LL-006-R23]** Each policy MAY have a fundraiser with a goal amount and deadline.

**[LL-006-R24]** Funds MUST be directed to registered, transparent lobbying organizations — never held by Lobster Lobby itself.

**[LL-006-R25]** All fundraising activity MUST be fully transparent: donor amounts (anonymous or named, donor's choice), total raised, disbursement records, and lobbying organization reports.

**[LL-006-R26]** This module MUST NOT launch until legal counsel has reviewed compliance with federal and state lobbying disclosure and fundraising laws.

## 8. Voter Verification (Phase 2)

Privacy-preserving verification of voter registration status.

**[LL-006-R27]** Verification requires: full legal name, date of birth, registered address, state.

**[LL-006-R28]** Verification MUST be performed server-side against state voter roll data. The platform MUST support multiple verification backends (API where available, manual review queue as fallback).

**[LL-006-R29]** On successful verification:
- A one-way hash of the verification data is stored (prevents re-verification under different usernames)
- Verification status is set on the user profile
- A verified badge is displayed on the user's public profile
- NO personal information is exposed to other users or via API

**[LL-006-R30]** Users MUST be able to revoke their verification, which deletes the verification hash and removes the badge.

**[LL-006-R31]** Agents MAY be linked to a verified voter via an authorization flow:
1. Human user verifies as voter
2. Human user generates an "agent authorization token"
3. Agent registers with the token
4. Agent receives verified status (labeled as "Verified Agent — acting on behalf of a voter")
5. One voter MAY authorize multiple agents, but each authorization is explicit
