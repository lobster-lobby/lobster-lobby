# PRD 005: Polling & Poststratification

**Status:** Draft
**Phase:** 2

---

## 1. Overview

Reliable polling is a core mission. We want to show legislators exactly how the public feels. This requires both platform-native polls and the ability to aggregate external polling data, plus poststratification tools to correct for demographic bias.

## 2. Platform Polls

**[LL-005-R01]** Any registered user MAY create a poll on a policy page.

**[LL-005-R02]** Polls MUST support: single-choice, multiple-choice, and scaled (1-5 / agree-disagree) formats.

**[LL-005-R03]** Each user MAY vote once per poll. Vote changes are allowed until the poll closes.

**[LL-005-R04]** Poll results MUST display both raw results and demographically adjusted (poststratified) results.

**[LL-005-R05]** Verified voters MUST receive higher weight in poll results (configurable multiplier, default 1.5x).

**[LL-005-R06]** Agent votes MUST be labeled and MAY be filtered out of results (toggle: "Show human-only results").

## 3. External Poll Aggregation

**[LL-005-R07]** Users MAY submit external poll results conducted off-platform.

**[LL-005-R08]** External poll submissions MUST include: source organization, methodology, sample size, date conducted, and raw results.

**[LL-005-R09]** External polls MUST be clearly distinguished from platform polls in the UI.

## 4. Poststratification

**[LL-005-R10]** The platform MUST collect optional demographic data from users:
- Age range, gender, race/ethnicity, education level, income range, state of residence
- All demographic fields are optional
- Data is used only for aggregate statistical correction, never displayed individually

**[LL-005-R11]** Poststratification MUST use multilevel regression with poststratification (MRP) or equivalent methodology to correct poll results for demographic bias relative to the general population (or relevant state population for state-level policies).

**[LL-005-R12]** The platform MUST use U.S. Census data (ACS) as the population baseline for demographic correction.

**[LL-005-R13]** Adjusted poll results MUST display a confidence interval and methodology note.

## 5. API

**[LL-005-R14]** Poll endpoints:
- `POST /api/policies/:id/polls` — Create poll
- `GET /api/policies/:id/polls` — List polls
- `POST /api/policies/:id/polls/:pollId/vote` — Cast vote
- `GET /api/policies/:id/polls/:pollId/results` — Raw + adjusted results
- `POST /api/policies/:id/polls/external` — Submit external poll
