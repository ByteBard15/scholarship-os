# Proposal risk assessments

Risk factor is an evidence-based planning index, not a probability of admission, funding success or rejection. Ordinal factor values cannot establish calibrated probabilities. Keep it independent of proposal priority/rank and research evidence confidence.

Every proposal card in Research details displays its latest assessment or “Not assessed”, factor breakdown, unknowns, blockers, computation date and history. A New assessment dialog lets the owner choose the credential profile and enter variables. Suggested variables include academic eligibility, research fit, procedure complexity, document readiness, funding feasibility and **supervisor accessibility**. Add/remove variables to suit the proposal; do not enter numbers without a rationale. No existing proposal receives invented default scores.

## Scoring

Each `variableName` is unique lower_snake_case, with `value` from 0 to 100: higher means greater risk. The stored `riskFactor` is the arithmetic mean of the known values, with methodology `mean_v1`. Blank/null values are unknown and excluded from the denominator; all unknown means null/unscored. Show the unknown count and hard blockers alongside the result. A low average cannot cancel a disqualifying requirement. `blocking` identifies such an unresolved hard blocker independently of the score. Duplicate names, nonfinite values and out-of-range values are rejected.

Suggested evidence-oriented rubric: 0 means this factor is demonstrably satisfied/low risk; 25 minor outstanding work; 50 a substantial unresolved dependency; 75 a serious documented obstacle; 100 a documented unmet requirement. These are qualitative planning anchors, not observed outcome frequencies. Use null where evidence does not establish an assessment, rather than assigning 50 to unknowns.

## Time and provenance

Every assessment is append-only, server-timestamped and attributed to the authenticated principal. It references an immutable EffectiveProfile snapshot created through the profile service and stores a snapshot of the proposal and its source links. Changing credentials, research or factor values requires a new assessment; history is preserved. No edits/deletes are exposed.

Every variable includes `observedAt` and a rationale, with an optional HTTP(S) source URL. A `timeVariable` factor requires `timePredicate` and `validUntil`. Example: supervisor accessibility is assessed only while the faculty member is accepting students for the stated intake, based on their reply, and must be reviewed by the selected date. A review-by date is an assessor's planning decision, not a university deadline. Also reassess immediately if the condition ceases to hold before that date.

Responses flag stale evidence once `now >= validUntil`; original values and their mean remain unchanged. This is an expiry check, not automatic monitoring of supervisors or profile changes. The UI explicitly warns that subsequent credential/evidence changes require reassessment. Times are UTC in API/database and displayed locally. Conditions are explanatory text; the system does not execute natural-language predicates.

## Endpoints

USER/SYSTEM, with existing proposal ownership middleware plus service checks:

- `GET /api/v1/application-proposals/{proposalID}/risk-assessments`
- `POST /api/v1/application-proposals/{proposalID}/risk-assessments`

AGENT: the same paths under `/api/v1/agent/application-proposals/...`. Listing requires `research`; creating additionally requires `profiles` to snapshot owned credentials. Dedicated typed tools `ListProposalRisk` and `CreateProposalRisk` are also available. No proposal approval or application creation occurs when recording a risk assessment.

GET returns up to 50 most recent assessments, newest first, with all variables. Public JSON uses camelCase; database tables `proposal_risk_assessments` and `proposal_risk_variables` use snake_case. All persistent IDs are UUIDs. Historical rows are removed with a deleted proposal; referenced credential snapshots are protected by RESTRICT.

Example request (illustrative only; replace dates, IDs and evidence):

```json
{
  "profileId": "<owned-profile-uuid>",
  "notes": "Planning assessment for the stated intake; not a calibrated success probability.",
  "variables": [
    {
      "variableName": "academic_eligibility",
      "value": null,
      "rationale": "Degree equivalence still requires programme confirmation.",
      "observedAt": "2026-11-07T09:00:00Z",
      "timeVariable": false,
      "blocking": false
    },
    {
      "variableName": "supervisor_accessibility",
      "value": null,
      "rationale": "No current acceptance/funding confirmation has been received.",
      "observedAt": "2026-11-07T09:00:00Z",
      "timeVariable": true,
      "timePredicate": "Only applies while this supervisor's intake availability and funding position remain unchanged.",
      "validUntil": "2026-11-14T09:00:00Z",
      "blocking": false
    }
  ]
}
```

Use the same set of variables when comparing snapshots; adding/removing variables changes the denominator. Always explain that change in assessment notes. Preserve source wording and intake specificity. Do not treat lack of supervisor response as proof of rejection or accessibility.

## Deployment and verification

Apply paired migration `000019_proposal_risk_assessments` with the existing migration script, then deploy backend and frontend together. This implementation does not run migrations or populate live scores. Down migration drops risk history; export it before rollback. Backend tests cover aggregation, unknowns, invalid values, temporal requirements, expiry boundaries, ownership, scopes, snapshot identity and historical preservation. Verified all 19 up migrations in a disposable PostgreSQL 17 database, migration 19 down/up, and database constraints for unknown-value averaging, duplicate variables, temporal predicates, ranges, snapshot foreign keys and cascading proposal deletion. The test container was removed; no live database was changed.
