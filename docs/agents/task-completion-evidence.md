# Task completion criteria and evidence

`completionCriteria` is optional text on an Application Task describing the expected result. Set it through existing task create/update requests; omission on update preserves it, and an empty string clears it. Existing description guides and completion statuses are preserved by migration 18.

`completionEvidence` is the concept shown as **Completion evidence** in the task dialog. Evidence is stored as append-only relational records, separately from criteria and task status. Corrections should be appended with an explanation referencing the earlier evidence ID. Evidence is an attributed observation, not verification, user approval or authority to submit an application.

USER routes (existing application ownership applies):

- `GET /api/v1/applications/{applicationID}/tasks/{taskID}/evidence`
- `POST /api/v1/applications/{applicationID}/tasks/{taskID}/evidence`

AGENT routes are the same under `/api/v1/agent/applications/...`, requiring the `tasks` scope and workspace ownership. SYSTEM retains its existing access. Services enforce authentication, scope, application ownership and task membership. IDs and recording actor/time are server-assigned. Evidence writes and their audit entry are transactional; task status is unchanged.

Example POST body (illustrative, never an assertion of an actual issued ID):

```json
{
  "kind": "identifier",
  "label": "University applicant ID",
  "value": "<actual ID from the university>",
  "sourceUrl": "https://my.ucalgary.ca/"
}
```

Supported kinds: `link`, `identifier`, `document`, `confirmation`, `note`. Labels are limited to 200 characters and values to 10,000; source URLs to 2,000. Link values and optional source URLs must use HTTP(S) without embedded credentials. Document values are owned document/writing references or explanatory text; this endpoint does not upload files. Multiple pieces of domain-specific evidence can be recorded for one task. Do not store secrets, one-time codes, signed download links or session credentials here.

The UI lists evidence with actor type and time and lets the applicant add records. Completion remains a separate action: a bare link is insufficient evidence that an application was submitted.

Deploy migration `000018_task_completion_evidence` before the backend; then deploy the frontend. Use the existing `scripts/migrate -m up` with the intended environment. The down migration removes evidence and criteria; back them up before rollback. This change does not run migrations against a live database.

For staged external automation and the applicant audit checkpoint, see [the Calgary plan](calgary-assisted-application-plan.md).
