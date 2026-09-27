# Application Proposals API

- `GET /api/v1/application-proposals?userId=&status=`
- `GET /api/v1/application-proposals/{proposalID}`
- `POST .../{proposalID}/approve` with `{ "parentProfileId": "..." }`
- `POST .../{proposalID}/reject`
- `POST .../{proposalID}/reopen`

A single Research Task or Research Run may create multiple proposals. Each proposal carries `priority` (`highest`, `high`, `medium`, or `low`) and an optional one-based `rank`. Lists are returned by priority and rank. `confidence` remains evidence confidence and is intentionally separate from recommendation priority. Users independently approve or reject each proposal and may approve more than one.

Human approval records the selected Master or Domain parent and changes the proposal to `approved`; it does not create an Application. An authenticated agent with the `applications` scope then polls `GET /api/v1/agent/application-proposals/approved` and calls `POST /api/v1/agent/application-proposals/{proposalID}/application`. That operation transactionally creates the catalog records where necessary, isolated Application Profile, Application, checklist, research links, outputs, and audit records. It is idempotency-guarded by the proposal's `applicationId`.
