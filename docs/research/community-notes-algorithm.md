# Community Notes Bridging Algorithm — Research & Design for Lobster Lobby

## How X's Community Notes Works

### The Core Idea: Matrix Factorization

Community Notes uses **matrix factorization** — the same technique Netflix uses for recommendations, but applied to fact-checking. The algorithm decomposes each user's vote into two components:

1. **Polarity factor** — How much of this vote is explained by the user's political bias?
2. **Intercept (helpfulness) factor** — How much of this vote is explained by something *independent* of political bias?

The intercept is the "common ground" — it represents quality that users recognize regardless of their political position.

### How It Works Step by Step

1. **Users rate notes** as "Helpful" or "Not Helpful"
2. **The algorithm assigns each user a polarity score** (roughly: left-leaning to right-leaning, but discovered from data, not self-reported)
3. **For each note, it runs a regression**: predicting votes as a function of user polarity
4. **The slope** tells us the note's political lean (polarizing notes have steep slopes)
5. **The intercept** tells us the note's "bridging" quality — how it would score if political bias didn't exist
6. **Only notes with a high intercept get shown** — they must be rated helpful by users *across* the political spectrum

### Key Insight: It's Not About Balance

The algorithm doesn't require equal numbers of left and right users. Even if 90% of raters are right-leaning, a note that only right-leaning users find helpful will have a high slope but low intercept — it won't be shown. The algorithm separates signal (genuine quality) from noise (political agreement).

### Why This Works

- People are politically biased, but they also have a tendency to upvote genuinely accurate, well-sourced information
- A right-leaning user will upvote right-supporting content, but will *especially* upvote right-supporting content that is also factually accurate
- They'll downvote left-supporting content, but will downvote *less zealously* when that content is well-sourced and fair
- By factoring out the political component, we can extract this quality signal

### The Formula

For each note `n` and user `u`, the predicted rating is:

```
rating(u, n) ≈ μ + intercept_note(n) + factor_user(u) × factor_note(n)
```

Where:
- `μ` = global average rating
- `intercept_note(n)` = the note's "bridging score" (this is what we want)
- `factor_user(u)` = the user's position on the discovered political axis
- `factor_note(n)` = how much the note's rating depends on political position

A note is shown when `intercept_note(n)` exceeds a threshold (currently 0.40 in Community Notes).

### Strengths
- Resistant to brigading — one-sided voting is factored out
- Discovers political positions from data (no self-reporting needed)
- Works even with unbalanced populations
- Completely open source (X publishes all code and data)

### Weaknesses (and how we can improve)
- **Single-axis assumption**: Community Notes models politics as a single left-right axis. Real opinion space is multi-dimensional.
- **Cold start problem**: New users with few ratings have uncertain polarity factors
- **Gaming potential**: Sophisticated actors could vote strategically to manipulate their polarity factor
- **Only binary ratings**: Helpful / Not Helpful. No nuance.

## Design for Lobster Lobby

### Adapting the Algorithm

Our use case is different from Community Notes in important ways:

| Community Notes | Lobster Lobby |
|----------------|---------------|
| Rates notes on posts | Rates summary points in debates |
| Binary: Helpful / Not Helpful | Endorsements with position context |
| Goal: surface accurate fact-checks | Goal: surface strongest arguments from each side + consensus |
| Single political axis | Support / Oppose / Neutral position on specific policy |
| Millions of users | Initially small community |

### Our Approach: Position-Aware Endorsement Scoring

Instead of discovering political positions from voting patterns (which requires massive data), we have an explicit signal: **each user declares their position** on a policy (support, oppose, neutral).

This gives us a huge advantage — we know exactly which "side" each endorser is on, without needing matrix factorization to discover it.

#### Scoring Formula

For each summary point `p`, calculate:

```
bridging_score(p) = (cross_endorsements × W_cross + same_endorsements × W_same) 
                    × reputation_multiplier × verification_multiplier
```

Where:
- `cross_endorsements` = endorsements from users who hold the OPPOSITE position to the point's stance
- `same_endorsements` = endorsements from users who hold the SAME position
- `W_cross = 3.0` (cross-position endorsements worth 3x)
- `W_same = 1.0`
- `reputation_multiplier` = average(endorser_reputation_tier_multiplier)
  - New (0-10): 0.5x
  - Regular (11-50): 1.0x
  - Trusted (51-200): 1.2x
  - Expert (201+): 1.5x
- `verification_multiplier`:
  - Unverified: 1.0x
  - Verified voter: 1.3x

#### Visibility Threshold

A summary point becomes visible when:
- `bridging_score >= 5.0` (minimum threshold)
- At least 2 endorsements from each side (cross-endorsement minimum)

Points are ranked by `bridging_score` descending.

#### Summary Point Lifecycle

1. **Nomination**: Any user can nominate a debate comment OR write an original summary point
2. **Endorsement**: Users endorse points they agree are well-stated, regardless of whether they agree with the conclusion
3. **Scoring**: Points are scored using the bridging formula
4. **Display**: Points above threshold appear in the Community Summary, ranked by score
5. **Decay**: Points gradually decay in score (half-life ~30 days) to keep summaries fresh

### Advantages Over Community Notes Approach

1. **No cold start**: We use explicit position declarations, not inferred positions
2. **Works at small scale**: Community Notes needs thousands of raters. Our system works with dozens.
3. **Multi-dimensional**: Each policy has its own support/oppose axis — no assumption of a single global political spectrum
4. **Richer signal**: Users declare positions, so we know exactly what a cross-endorsement means
5. **Transparent**: Users can see exactly why a point ranks high ("endorsed by 8 supporters and 5 opponents")

### Future Enhancement: Matrix Factorization

As the platform grows, we can layer on proper matrix factorization as a secondary signal:
- Discover multi-dimensional opinion clusters across policies
- Detect users who game the system by declaring false positions
- Find deeper consensus patterns across the full user base
- Weight endorsements by users who have historically been good-faith participants

This would be Phase 3+ work, implemented when we have enough data to make it meaningful.

### Implementation Notes

**Data model additions for endorsements:**

```go
type Endorsement struct {
    ID              ObjectID   `bson:"_id"`
    SummaryPointID  ObjectID   `bson:"summaryPointId"`
    UserID          ObjectID   `bson:"userId"`
    UserPosition    string     `bson:"userPosition"`  // "support" | "oppose" | "neutral"
    PointPosition   string     `bson:"pointPosition"` // position of the summary point
    CreatedAt       time.Time  `bson:"createdAt"`
}

type SummaryPoint struct {
    ID                  ObjectID   `bson:"_id"`
    PolicyID            ObjectID   `bson:"policyId"`
    Position            string     `bson:"position"` // "support" | "oppose"
    Content             string     `bson:"content"`
    NominatedBy         ObjectID   `bson:"nominatedBy"`
    SourceCommentID     *ObjectID  `bson:"sourceCommentId"` // if nominated from a debate comment
    BridgingScore       float64    `bson:"bridgingScore"`
    CrossEndorsements   int        `bson:"crossEndorsements"`
    SameEndorsements    int        `bson:"sameEndorsements"`
    TotalEndorsements   int        `bson:"totalEndorsements"`
    Visible             bool       `bson:"visible"`
    CreatedAt           time.Time  `bson:"createdAt"`
    LastScoreUpdate     time.Time  `bson:"lastScoreUpdate"`
}
```

**Score recalculation**: Triggered on each new endorsement. Could be batched for performance at scale.

## References

- [X Community Notes Algorithm Documentation](https://communitynotes.x.com/guide/en/under-the-hood/ranking-notes)
- [Birdwatch Paper (original algorithm)](https://github.com/twitter/communitynotes/blob/main/birdwatch_paper_2022_10_27.pdf)
- [Understanding Community Notes — Jonathan Warden](https://jonathanwarden.com/understanding-community-notes/)
- [Vitalik Buterin's Analysis](https://vitalik.eth.limo/general/2023/08/16/communitynotes.html)
- [Bridging-Based Ranking — Aviv Ovadya, Harvard Kennedy School](https://www.belfercenter.org/publication/bridging-based-ranking)
- [Threats to Sustainability of Community Notes — arXiv](https://arxiv.org/html/2510.00650v1)
- [pol.is — Bridging-based consensus platform](https://pol.is/home)
