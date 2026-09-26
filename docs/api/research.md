# Research API

- `POST /api/v1/applications/{applicationID}/research` runs the configured provider synchronously.
- `GET /api/v1/applications/{applicationID}/research` lists runs.
- `GET /api/v1/applications/{applicationID}/research/{runID}` gets one run.
- `GET .../{runID}/sources` and `GET .../{runID}/findings` expose provenance and proposals.
- `PATCH .../{runID}/findings/{findingID}` accepts `{ "reviewStatus": "accepted" }` or `rejected`.
- `POST .../{runID}/apply` applies accepted findings after every pending finding is reviewed.
- `POST /api/v1/applications/{applicationID}/research/mark-stale` marks research stale manually.

The development provider is a deterministic mock. Live web access and external AI credentials are not required. Applying research inserts new structured records; it does not overwrite conflicts or existing facts.
