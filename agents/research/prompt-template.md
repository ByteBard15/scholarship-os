# Scholarship research structured-output template

Research the named opportunity using the source priority in `AGENTS.md`. Return typed structured data for scholarship identity, institution, programme, eligibility, application process, admission requirements, scholarship requirements, required documents, deadlines, funding, contacts, supervisors, URLs, provenance, conflicts, and unknown values.

Return a separate application proposal for every viable programme/scholarship combination. Do not collapse distinct opportunities into one proposal. Assign each proposal a priority of `highest`, `high`, `medium`, or `low`; add a one-based rank when proposals are meaningfully comparable. Explain the priority using verified fit, funding, deadline feasibility, and unresolved eligibility—not admission probability.

For every claim include its source reference, confidence, verification status, and a short supporting excerpt when available. Keep admission and scholarship requirements separate. Preserve vague deadline wording and date precision. If two sources disagree, return both as conflicting findings.

Use `null` or `unknown` when information cannot be verified. Do not guess. Do not infer exact dates, amounts, applicant achievements, or eligibility. Never return free-form instructions that require direct database access; return the typed `ResearchResult` schema only.
