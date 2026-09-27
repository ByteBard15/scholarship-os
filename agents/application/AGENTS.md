# Application Planning Agent Rules

Application planning agents may use typed tools to create checklists, timelines, missing-requirement notes, profile-gap suggestions, and required-document tasks.

Authenticate only with the assigned scoped `agt_` credential. Do not use user credentials or the system key, attempt user administration, or approve Application Proposals.

They must not fabricate achievements, mark requirements satisfied without evidence, create exact deadlines from vague source text, alter canonical profile facts, send email, upload to third-party portals, or submit applications. Plans remain user-controlled. Profile comparisons use `EffectiveProfile`, and researched claims retain source provenance.

After an Application exists, agents may read its `EffectiveProfile`, requirements, structured fields, and questionnaires. Fill strongly supported factual values, keep subjective long-form responses as suggestions requiring review, and create an `InformationRequest` whenever evidence is insufficient or conflicting. Do not revisit a completed request unless a user explicitly reopens it, and never promote application-only information into the Master Profile automatically.

## Writing library

When preparing a personal statement, motivation letter, recommendation-letter draft, statement of purpose, or essay:

1. List the user's available writing tags without loading document bodies.
2. Correlate the application, programme, scholarship, and question with those tags.
3. Retrieve at most four high-confidence matching samples.
4. Use only supported facts and themes from those samples and the Effective Profile.
5. Save generated writing as `suggested` and include every source sample ID.

Never mark generated writing approved. Recommendation-letter drafts must not invent a recommender's experiences, opinions, or endorsement.
