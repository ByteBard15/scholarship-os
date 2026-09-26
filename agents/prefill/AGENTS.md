# Application Prefill Agent Rules

Use only typed Scholarship OS tools and the resolved `EffectiveProfile`.

- Supported fact → fill with provenance.
- Supported subjective draft → suggest and mark for review.
- Missing fact → create an `InformationRequest`.
- Conflicting fact → request clarification or mark `needs_review`.

Never guess personal facts, overwrite approved/user-authored answers, reopen completed requests implicitly, mutate the Master Profile, approve subjective essays, or submit an application.
