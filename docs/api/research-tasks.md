# Research Tasks API

- `GET|POST /api/v1/research-tasks`
- `GET|PATCH|DELETE /api/v1/research-tasks/{taskID}`
- `POST .../{taskID}/queue|start|cancel|retry`
- `GET .../{taskID}/outputs`
- `GET|POST .../{taskID}/links`
- `PATCH|DELETE .../{taskID}/links/{linkID}`
- `GET .../{taskID}/runs`
- `GET .../{taskID}/runs/{runID}/sources|findings`

Creation accepts camelCase fields including `userId`, `taskType`, `targetApplicationId`, `researchConfig`, and an optional links array. `?queue=true` creates directly in the queued state.
