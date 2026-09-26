# Research Tasks API

- `GET|POST /api/v1/research-tasks`
- `GET|PATCH|DELETE /api/v1/research-tasks/{taskID}`
- `POST .../{taskID}/queue|cancel|retry`
- `GET .../{taskID}/outputs`
- `GET|POST .../{taskID}/contexts`
- `DELETE .../{taskID}/contexts/{contextID}`
- `GET|POST .../{taskID}/links`
- `PATCH|DELETE .../{taskID}/links/{linkID}`
- `GET .../{taskID}/runs`
- `GET .../{taskID}/runs/{runID}/sources|findings`

Creation accepts camelCase fields including `userId`, `taskType`, `targetApplicationId`, `researchConfig`, an optional links array, reusable `researchContextIds`, and `newResearchContexts`. Inline contexts are saved to the user's reusable context library and attached to the new task in the same transaction. `?queue=true` creates directly in the queued state.

The reusable context library uses:

- `GET|POST /api/v1/research-contexts`
- `PATCH|DELETE /api/v1/research-contexts/{contextID}`

Each context is a user-owned question and answer. Context ownership is validated in the service before attachment; contexts cannot be read or attached across workspaces.

Scoped research agents consume queued work through:

- `GET /api/v1/agent/research-tasks?limit=N` — polls the oldest queued tasks; `N` defaults to 10 and is capped at 50.
- `POST /api/v1/agent/research-tasks/{taskID}/start` — atomically claims one task and returns both the task and its new `researchRun`.
- `GET .../{taskID}/links|contexts|runs` — loads agent input and recovers the current run ID.
- `POST .../{taskID}/sources|findings` — persists typed, provenance-bearing results.
- `POST /api/v1/agent/application-proposals` — creates a proposal for human review.
- `POST .../{taskID}/complete` with `{ "researchRunId": "..." }` — finishes submission and marks the task/run `review_required`.

The API never performs live or mock research while handling `start`. If several agents poll the same task, only the first successful claim may transition it from `queued` to `running`.
