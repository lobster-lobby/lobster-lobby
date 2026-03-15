# PRD 003: Research Module

**Status:** Draft
**Author:** Victor Brechbill & Nova
**Date:** 2026-03-15

---

## 1. Overview

The research module allows users to submit findings, news, data, and analysis related to a policy. Think of it as a curated knowledge base for each policy issue.

## 2. Research Submissions

**[LL-003-R01]** Research submissions MUST include:
- Title
- Type: `analysis`, `news`, `data`, `academic`, `government`
- Content (Markdown)
- At least one source citation

**[LL-003-R02]** Sources MUST include: URL, title, and optionally publisher and publication date.

**[LL-003-R03]** Research MAY include file attachments: charts, PDFs, images, datasets (CSV/JSON).

**[LL-003-R04]** Research MUST support upvotes and downvotes for community ranking.

**[LL-003-R05]** Research MUST be sortable by: Newest, Top (most upvoted), Most Cited (referenced by other research/debates), Type.

**[LL-003-R06]** Research MUST be filterable by type, date range, and fact-check status.

## 3. Source Quality

**[LL-003-R07]** Sources from government domains (.gov), academic institutions (.edu), and known research organizations SHOULD receive a visual "institutional source" indicator.

**[LL-003-R08]** Users MAY challenge the reliability of a source by flagging it with a reason.

## 4. Cross-Referencing

**[LL-003-R09]** Research from one policy MAY be cross-referenced to related policies. This creates a link visible on both policy dashboards.

**[LL-003-R10]** Debate comments MAY cite research items by ID, creating a traceable evidence chain.

## 5. API

**[LL-003-R11]** Research endpoints:
- `GET /api/policies/:id/research` — List research with filtering/sorting
- `POST /api/policies/:id/research` — Submit research (requires title, type, content, sources)
- `GET /api/policies/:id/research/:researchId` — Get single research item
- `POST /api/policies/:id/research/:researchId/vote` — Upvote/downvote
- `POST /api/policies/:id/research/:researchId/cite` — Cross-reference to another policy
