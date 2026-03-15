# PRD 004: Representative Lookup & Legislator Scorecards

**Status:** Draft
**Author:** Victor Brechbill & Nova
**Date:** 2026-03-15

---

## 1. Overview

Connect users to their elected representatives. Show contact information, voting records on platform issues, and community-generated alignment scores. Help users take direct action.

## 2. Representative Lookup

**[LL-004-R01]** Users MUST be able to find representatives by:
- State (returns U.S. senators + governor)
- Congressional district number (returns House representative)
- Street address (returns all applicable federal + state representatives)

**[LL-004-R02]** Address-based lookup MUST use the Google Civic Information API (or equivalent) to resolve district from address.

**[LL-004-R03]** The platform MUST NOT store user addresses used for lookup. Lookups are stateless unless the user explicitly saves their district to their profile.

## 3. Representative Profiles

**[LL-004-R04]** Each representative MUST have a profile page showing:
- Name, title, party, state/district
- Official photo (sourced from congress.gov or equivalent)
- Contact information: office phone, email, website, social media
- Committee and subcommittee memberships

**[LL-004-R05]** Representative profiles MUST show voting record on platform policies:
- How they voted on each `active_bill` or `existing_law` tracked on the platform
- Whether they sponsored or co-sponsored the legislation
- Visual indicator: ✅ Voted Yes, ❌ Voted No, ⬜ Abstained, — No Vote

**[LL-004-R06]** Representative profiles MUST display a community alignment score:
- Calculated as: (votes matching community poll majority) / (total tracked votes)
- Displayed as a percentage with visual indicator (e.g., "73% aligned with community positions")

## 4. Policy ↔ Representative Connection

**[LL-004-R07]** Each policy page MUST show relevant representatives:
- Bill sponsors and co-sponsors
- Committee members with jurisdiction
- Recent voters (if the bill has had floor votes)

**[LL-004-R08]** Each representative on a policy page MUST show their known position (voted yes/no, sponsored, or unknown).

## 5. Campaign Finance (Phase 3)

**[LL-004-R09]** Representative profiles SHOULD display campaign finance data from OpenSecrets:
- Top donors (organizations)
- Industry funding breakdown
- Total fundraising for current cycle

**[LL-004-R10]** On each policy page, relevant industry donors SHOULD be highlighted (e.g., if the policy is about pharmaceutical regulation, show pharma industry donations to relevant committee members).

## 6. Action Center (Phase 3)

**[LL-004-R11]** Users MUST be able to take action directly from the platform:
- Pre-drafted email templates for each policy (customizable)
- Representative phone numbers with call scripts
- Links to upcoming town halls and public events
- Voter registration links (state-specific)

**[LL-004-R12]** Action templates MUST be politically neutral — they present the user's position without editorial bias.

## 7. API

**[LL-004-R13]** Representative endpoints:
- `GET /api/representatives?state=MI` — By state
- `GET /api/representatives?district=MI-13` — By district
- `GET /api/representatives?address=...` — By address (proxied to Civic API)
- `GET /api/representatives/:id` — Full profile
- `GET /api/representatives/:id/votes` — Voting record
- `GET /api/policies/:id/representatives` — Representatives relevant to a policy
