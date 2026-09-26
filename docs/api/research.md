# Research API

- `GET /api/v1/applications/{applicationID}/research` lists runs.
- `GET /api/v1/applications/{applicationID}/research/{runID}` gets one run.
- `GET .../{runID}/sources` and `GET .../{runID}/findings` expose provenance and proposals.
- `PATCH .../{runID}/findings/{findingID}` accepts `{ "reviewStatus": "accepted" }` or `rejected`.
- `POST .../{runID}/apply` applies accepted findings after every pending finding is reviewed.
- `POST /api/v1/applications/{applicationID}/research/mark-stale` marks research stale manually.

Research is not executed synchronously by the API. An authenticated external agent polls queued Research Tasks, claims a task, and submits its sources and findings through the agent endpoints documented in `research-tasks.md`. Applying reviewed research inserts new structured records; it does not overwrite conflicts or existing facts.
