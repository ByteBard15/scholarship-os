# Scholarship Research Agent Rules

Research agents interact with Scholarship OS only through typed application tools. They never execute SQL, access GORM models, or reconstruct profile inheritance. Applicant comparisons use the profile service's `EffectiveProfile`.

Authenticate with the assigned `agt_` credential only. Never use a user's password/session or the root-equivalent system key, and never call proposal approval or user administration endpoints.

## Source priority

1. Official scholarship provider
2. Official university
3. Official programme or faculty
4. Official government source
5. Official supervisor profile
6. Trusted secondary source
7. General web source

Prefer higher-ranked sources but retain conflicting claims and their provenance. Record every URL, publisher, source type, retrieval date, and supporting text where practical.

## Research rules

- Never invent deadlines, eligibility rules, funding, contacts, supervisors, or required documents.
- Never infer an exact date from wording such as “January 2027” or “early December.” Preserve raw text and the correct precision.
- Distinguish verified and supported facts from inference, uncertainty, and conflicts.
- Separate programme admission requirements from scholarship requirements.
- Separate admission, funding, nomination, and internal deadlines.
- Separate funding terms from admission terms.
- Use null or unknown when a fact cannot be verified.
- Do not assess admission probability.
- Do not modify canonical applicant facts.
- Research findings remain pending until a controlled review accepts or rejects them.
- Applying accepted findings must preserve the `ResearchSource` relationship.

The current provider is deterministic and local. Future live web or AI providers must return the same typed `ResearchResult` and follow these rules.

## Research Task workflow

1. Fetch the `ResearchTask` through typed tools.
2. Read the title, description, custom instructions, and every supplied link.
3. Research official sources first and persist sources before claims.
4. Persist typed findings with confidence, supporting text, and verification state.
5. Create an `ApplicationProposal` with source links when an opportunity is found.
6. Do not create an `Application`.
7. Mark the task `review_required` and stop until a human approves or rejects the proposal.

Rejected proposals and their research history remain available. Never bypass the proposal approval boundary.
