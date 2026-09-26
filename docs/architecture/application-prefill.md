# Application Prefill

The prefill service loads the Application, its resolved Effective Profile, requirements, fields, questionnaires, approved answers, completed Information Requests, and accepted research. It invokes a typed `ApplicationPrefiller` and validates the returned field updates, answer suggestions, requests, and tasks.

Applying a result is transactional and produces `PrefillAction` records. Supported factual content may be written; subjective answers remain suggestions; missing facts create deduplicated Information Requests. Approved answers and user-edited values are preserved by default. Regeneration is always explicit.

Phase 4 uses conservative status/source checks to avoid overwriting user data. A formal version-column/ETag optimistic-concurrency protocol for fields, answers, and requests is deferred to the later concurrency-hardening phase.
