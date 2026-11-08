# Cascading research task deletion

The research listing and detail screen offer **Delete research task**. The dialog previews every affected task (including descendants), proposal and application. Type `DELETE` to confirm permanent deletion. The button shows progress and prevents duplicate requests.

Deletion runs in one database transaction. It removes the selected task subtree, its proposals, applications materialized from those proposals, and their owned research, checklist/evidence, questionnaire, prefill, information request, risk and activity records. A task's `targetApplicationId` alone does not make that application owned by the research task.

Master/domain profiles, reusable research contexts, scholarship catalogue records and writing samples survive. Writing is detached from deleted applications. Isolated application profiles are archived, preserving snapshots and writing provenance. Shared application profiles remain available. No university portal applications are deleted.

Running tasks, research runs or prefill runs block deletion. An application linked to proposals outside the selected subtree also blocks deletion. The dialog refreshes its preview after a failed request. Changed task/proposal/application scope requires confirmation again. Successful deletion is recorded in the audit log.

Authenticated workspace owners (and SYSTEM) can use:

- `GET /api/v1/research-tasks/{taskID}/deletion-preview`
- `POST /api/v1/research-tasks/{taskID}/cascade-delete` with `{ "confirmationToken": "<token from preview>" }`

Agents have no cascade-delete access. The existing `DELETE /api/v1/research-tasks/{taskID}` retains its earlier soft-delete behavior for compatibility.

No database migration is required; this uses existing foreign keys. Deploy the backend and frontend together. To run the destructive integration test, set `WORKFLOW_TEST_DATABASE_URL` to a disposable localhost PostgreSQL database and run `go test ./internal/features/workflow -run TestResearchCascadeDeletion` from `apps/api`. The test creates and removes an isolated schema.
