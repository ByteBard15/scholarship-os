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

Creation accepts camelCase fields including `userId`, `profileId`, `taskType`, `targetApplicationId`, `researchConfig`, an optional links array, reusable `researchContextIds`, and `newResearchContexts`. Inline contexts are saved to the user's reusable context library and attached to the new task in the same transaction. The selected profile must belong to the task owner. A profile is required before the task can be queued.

`PATCH /api/v1/research-tasks/{taskID}` updates core task fields, including `profileId`, only while the task is in `draft`. Links and context attachments are likewise mutable only in `draft`. Once queued, research input is frozen.

The reusable context library uses:

- `GET|POST /api/v1/research-contexts`
- `PATCH|DELETE /api/v1/research-contexts/{contextID}`

Each context is a user-owned question and answer. Context ownership is validated in the service before attachment; contexts cannot be read or attached across workspaces.

Scoped research agents consume queued work through:

- `GET /api/v1/agent/research-tasks?limit=N` — polls queued tasks by priority, earliest deadline, then oldest creation time; `N` defaults to 10 and is capped at 50.
- `POST /api/v1/agent/research-tasks/{taskID}/start` — atomically claims one task and returns both the task and its new `researchRun`.
- `GET .../{taskID}/links|contexts|effective-profile|runs` — loads agent input, resolves the selected profile through the profile service, and recovers the current run ID.
- `POST .../{taskID}/sources|findings` — persists typed, provenance-bearing results.
- `POST /api/v1/agent/application-proposals` — creates one proposal for human review. Call it once per viable opportunity; include `priority` and optional `rank` to present a useful shortlist.
- `POST .../{taskID}/complete` with `{ "researchRunId": "..." }` — finishes submission and marks the task/run `review_required`.

The API never performs live or mock research while handling `start`. If several agents poll the same task, only the first successful claim may transition it from `queued` to `running`.

## Listing, scheduling, and pagination

`GET /api/v1/research-tasks` accepts `page` (default 1, maximum 1000000),
`pageSize` (default 20, maximum 100), and `sortBy` (`createdAt`, `priority`,
or `dueAt`). Filters are `userId`, `parentTaskId`, `status`, `taskType`, and
`priority`. Ownership applies to both the rows and the total count.

The response has `data` and `meta`: `count` is the number of rows on this
page, `total` is the filtered total, and `page`, `pageSize`, and `totalPages`
describe pagination. An empty result has `totalPages: 0`. Out-of-range pages
return an empty page; invalid page, page size, or sort values return HTTP 400.

Each task exposes `priority` (defaults to `normal`) and `dueAt` (UTC timestamp,
or `null` when unknown). This is the deadline for completing research, separate
from sourced scholarship/application deadlines. Creation accepts both fields.
Draft updates accept `dueAt` or `clearDueAt: true` to remove a deadline.

Priority sorting puts urgent, high, normal, then low tasks first, with earliest
known deadlines first within a priority. Deadline sorting puts unknown deadlines
last. Creation time and UUID break ties consistently.

Agent polling retains its `limit` contract, returning queued tasks by priority,
then earliest known deadline, then oldest creation time and UUID. The export
endpoint returns the complete filtered set, without page or page-size limits.

### Researched scholarship deadlines

The research list and detail responses include `scholarshipDeadlines`, derived
from sourced, non-rejected `deadline` findings in the task's latest run, only
when both the task and latest run are ready for review or completed. Unfinished
reruns do not expose older deadlines. Each entry retains `findingId`, `field`,
`value`, optional `rawText`/`notes`, `sourceUrl`, `verificationStatus`, and
`reviewStatus`. Multiple nomination, funding, or admission deadlines stay
separate. Vague wording and conflicts are preserved; pending findings remain
marked as awaiting review. An empty array means research has not established a
sourced deadline. The UI displays these findings, not the internal `dueAt`
scheduling field, and does not ask users to enter a research deadline.
