# PRD 008: Campaigns

**Status:** Draft
**Author:** Victor Brechbill & Nova
**Date:** 2026-03-16

---

## 1. Overview

Campaigns are Lobster Lobby's action layer. While Debate and Research help communities understand policy issues, and Draft helps them write better policy, **Campaigns turn that work into real-world impact** — before there's money for professional lobbying.

A Campaign is centered around a single policy that has been through the full deliberation process: researched, debated, and refined into a concrete policy proposal. It provides tools for grassroots advocacy: shareable assets, coordinated messaging, reach tracking, and community discussion.

**The lifecycle:**
```
Policy Created → Debated → Researched → Drafted/Amended → Ready for Campaign → Campaign Active → Campaign Complete
```

Campaigns are the answer to "OK, we agree on what the policy should say — now what?"

## 2. Campaign Eligibility

**[LL-008-R01]** A Campaign can ONLY be created for a policy that has been marked as **"Ready for Campaign"** status. This status is earned, not granted:
- The policy MUST have an associated draft or amendment (from the Draft module)
- The policy MUST have a minimum engagement threshold (configurable, suggested: 10+ debate comments, 3+ research submissions)
- A moderator-tier user (reputation ≥ 200) OR the policy creator MUST nominate the policy for campaign readiness
- Community vote: 5+ users with reputation ≥ 50 must endorse the "ready for campaign" nomination

**[LL-008-R02]** The "Ready for Campaign" status MUST be visible on the policy page as a distinct badge/banner, signaling the community has done the work.

**[LL-008-R03]** A policy MAY have multiple campaigns (e.g., one targeting federal reps, another targeting a specific state legislature), but each campaign must have a distinct **target audience** or **objective**.

## 3. Campaign Model

**[LL-008-R04]** A Campaign MUST include:
- **Title** — Clear, action-oriented (e.g., "Pass the Algorithmic Accountability Act")
- **Policy link** — The policy this campaign advocates for
- **Objective** — What specific outcome the campaign seeks (e.g., "Get 10 co-sponsors in the Senate")
- **Target** — Who the campaign is trying to influence (specific legislators, committee, general public, etc.)
- **Description** — Markdown summary of the campaign's strategy and call to action
- **Status** — `active` | `paused` | `completed` | `archived`
- **Created by** — The user who launched the campaign
- **Milestones** — Optional ordered list of goals (e.g., "1000 shares", "Committee hearing", "Floor vote")

**[LL-008-R05]** A Campaign MUST track:
- Total asset downloads
- Total shares (self-reported by users)
- Unique participants (users who downloaded, shared, or contributed assets)
- Discussion activity (comment count)
- Timeline of major events/updates

## 4. Campaign Discussion

**[LL-008-R06]** Each Campaign MUST have a discussion thread — a Reddit-style comment log where users can:
- Discuss strategy and tactics
- Share updates (e.g., "I sent this to my rep and got a response!")
- Coordinate efforts
- Post progress reports

**[LL-008-R07]** Discussion comments are threaded (parent/reply), voteable (upvote/downvote), and sortable (newest, top, discussed). Reuses the existing comment infrastructure from the Debate module.

**[LL-008-R08]** Discussion comments do NOT require a position (support/oppose/neutral). This is coordination, not debate — everyone in a campaign is already aligned on the policy direction.

**[LL-008-R09]** Campaign creators and moderator-tier users CAN pin important comments to the top of the discussion (max 3 pinned).

## 5. Campaign Assets

**[LL-008-R10]** Users MUST be able to submit **Assets** to a Campaign. An Asset is a piece of advocacy material designed to be shared, sent, or printed.

**[LL-008-R11]** Asset types:
| Type | Description | Format |
|------|-------------|--------|
| **Text Post** | Pre-written social media post, talking points, elevator pitch | Plain text / Markdown |
| **Email Draft** | Ready-to-send email to legislators, editors, or community | Markdown with suggested subject line + recipients |
| **Social Media Image** | Shareable graphic for Twitter/Instagram/Facebook | Image (PNG/JPG), recommended sizes noted |
| **Infographic** | Data visualization or explainer graphic | Image (PNG/JPG/SVG) |
| **Flyer / Print Material** | Printable one-pager, poster, handout | PDF or image |
| **Letter Template** | Physical letter to mail to representatives | Markdown with formatting |
| **Video Script** | Script for short-form video (TikTok, Reels, YouTube Shorts) | Plain text / Markdown |
| **Talking Points** | Bullet-point summary for conversations, town halls, phone calls | Plain text / Markdown |

**[LL-008-R12]** Each Asset MUST include:
- Title
- Type (from the table above)
- Content (text) OR file attachment (image/PDF)
- Creator
- Description / usage instructions
- Download count
- Share count (self-reported)
- Upvotes / Downvotes (community quality signal)

**[LL-008-R13]** Assets are community-curated: users upvote/downvote assets, and the best ones rise to the top. This ensures quality control without heavy moderation.

**[LL-008-R14]** Each Asset MUST have its own mini-discussion thread for feedback, improvement suggestions, and coordination (e.g., "Can someone translate this to Spanish?").

**[LL-008-R15]** Users MUST be able to **download** any asset with a single click:
- Text assets: copy-to-clipboard button + download as .txt/.md
- Email drafts: copy-to-clipboard + "Open in email client" (mailto: link with pre-filled subject/body)
- Images: direct download
- PDFs: direct download

**[LL-008-R16]** Users MUST be able to **report sharing** an asset. This is voluntary/honor-system:
- "I shared this" button → increments share count
- Optional: where they shared it (Twitter, Facebook, email, printed, etc.)
- This feeds into campaign reach metrics

## 6. Reach Tracking & Metrics

**[LL-008-R17]** Each Campaign MUST display a **Reach Dashboard** showing:
- **Total downloads** (sum across all assets)
- **Total reported shares** (sum across all assets, broken down by platform)
- **Unique participants** (users who contributed, downloaded, or shared)
- **Discussion activity** (comments, active commenters)
- **Asset count** (total submitted, by type)
- **Trending score** — calculated from recent activity velocity

**[LL-008-R18]** Campaign reach metrics MUST be visible on the campaign card in the campaign listing, so users can quickly see which campaigns are most active.

**[LL-008-R19]** The Campaign page MUST include a **Timeline / Activity Feed** showing:
- Asset submissions
- Milestone achievements
- Pinned discussion updates
- Significant share milestones (e.g., "Campaign reached 500 shares!")

**[LL-008-R20]** Trending score algorithm:
```
trending = (downloads_7d * 1) + (shares_7d * 3) + (new_assets_7d * 5) + (comments_7d * 0.5)
           * recency_multiplier
```
Where `recency_multiplier` decays campaigns that haven't had activity in 14+ days.

## 7. Campaign Listing & Discovery

**[LL-008-R21]** The platform MUST have a **Campaigns page** (`/campaigns`) showing:
- Active campaigns, sorted by: Trending (default), Newest, Most Participants, Most Shares
- Filter by: policy level (federal/state), state, policy tags
- Each campaign card shows: title, policy name, objective, participants, shares, trending score

**[LL-008-R22]** Active campaigns MUST also appear on their associated policy's detail page, in a "Campaigns" tab alongside Debate, Research, etc.

**[LL-008-R23]** The homepage SHOULD highlight featured/trending campaigns (after launch, when campaigns exist).

## 8. Campaign Lifecycle

**[LL-008-R24]** Campaign statuses and transitions:
- **Active** — Accepting assets, discussion open, tracking shares
- **Paused** — Temporarily frozen (e.g., waiting for legislative session). No new assets, discussion read-only.
- **Completed** — Objective achieved or campaign concluded. Becomes read-only archive with final metrics displayed.
- **Archived** — Removed from active listings but still accessible via direct link.

**[LL-008-R25]** Only the campaign creator or moderator-tier users can change campaign status.

**[LL-008-R26]** When a campaign is marked **Completed**, the creator SHOULD write a summary: what was achieved, what impact was made, lessons learned. This becomes part of the platform's institutional knowledge.

## 9. Permissions & Moderation

**[LL-008-R27]** Creating a campaign requires:
- Authenticated user
- Reputation ≥ 50 (member tier or above)
- Policy must be in "Ready for Campaign" status

**[LL-008-R28]** Submitting assets requires:
- Authenticated user
- Any reputation level (lower barrier than campaign creation)

**[LL-008-R29]** Assets can be flagged using the same moderation system as debates (PRD-002). Inappropriate or off-topic assets can be downranked or removed by moderators.

**[LL-008-R30]** Campaign discussion uses the same moderation rules as debate comments.

## 10. Agent Integration

**[LL-008-R31]** AI agents MAY participate in campaigns:
- Submit assets (e.g., draft email templates, generate social media post variations)
- Contribute to discussion
- Agent-generated assets MUST be clearly labeled as AI-generated

**[LL-008-R32]** AI agents MUST NOT create campaigns. Campaign creation is a human-only action — it requires judgment about when a policy is truly ready for advocacy.

**[LL-008-R33]** Agent-generated assets SHOULD be treated as drafts/suggestions that humans can adopt, modify, and share. The human sharing it takes ownership.

## 11. Future Considerations (Not In Scope)

- **Fundraising integration** — When the org is ready for lobbying, campaigns could include donation targets
- **Direct legislator contact** — Integrated calling/faxing tools (requires Telnyx or similar)
- **Event coordination** — Town halls, rallies, lobby days linked to campaigns
- **Impact verification** — Connecting campaign activity to actual legislative outcomes (votes, co-sponsors)
- **Cross-campaign coordination** — Linking related campaigns across policy areas
