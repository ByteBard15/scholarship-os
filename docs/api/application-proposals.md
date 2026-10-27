# Application Proposals API

- `GET /api/v1/application-proposals?userId=&status=`
- `GET /api/v1/application-proposals/{proposalID}`
- `POST .../{proposalID}/approve` with `{ "parentProfileId": "..." }`
- `POST .../{proposalID}/reject`
- `POST .../{proposalID}/reopen`

A single Research Task or Research Run may create multiple proposals. Each proposal carries `priority` (`highest`, `high`, `medium`, or `low`) and an optional one-based `rank`. Lists are returned by priority and rank. `confidence` remains evidence confidence and is intentionally separate from recommendation priority. Users independently approve or reject each proposal and may approve more than one.

Human approval records the selected Master or Domain parent and changes the proposal to `approved`; it does not create an Application. An authenticated agent with the `applications` scope then polls `GET /api/v1/agent/application-proposals/approved` and calls `POST /api/v1/agent/application-proposals/{proposalID}/application`. That operation transactionally creates the catalog records where necessary, isolated Application Profile, Application, checklist, research links, outputs, and audit records. It is idempotency-guarded by the proposal's `applicationId`.

## Research finding transfer during Application creation

Research finding values may use flexible evidence shapes; they are not assumed
to be valid application DTOs. Materialization validates supported records before
inserting them. Unsupported shapes remain linked to the Application as research
findings with their existing review status, rather than creating empty records
or aborting the workspace. Rejected findings are never applied or reaccepted.

Deadline transfer accepts research `rawText` as `rawDeadlineText`, and normalizes
`date` precision to `day`. Unrecognized precision values become `unknown` while
the original research value remains intact. Month-level, approximate and unknown
deadlines retain source wording without an exact timestamp; no timestamp is
invented from `localDate`, recurring rules or prose. All transferred records
retain their source references. No database migration is needed for this change.

An opt-in PostgreSQL regression test applies the real migrations inside an
isolated schema on a disposable local database:

```sh
cd apps/api
WORKFLOW_TEST_DATABASE_URL='postgres://postgres@127.0.0.1:5432/test?sslmode=disable' \
  go test ./internal/features/workflow -run TestProposalMaterializationWithResearchDeadlineValues -count=1
```

The test requires localhost and removes its generated schema afterward. Use a
disposable test instance. Ordinary `go test ./...` runs the adapter unit tests
and skips this integration test when the variable is unset.
