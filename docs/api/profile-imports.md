# Profile documents and imports API

`POST /api/v1/profiles/:profileID/documents` accepts `multipart/form-data` fields `file` and `documentType`. Supported types are `cv`, `transcript`, `certificate`, `publication`, `portfolio`, and `other`. Accepted formats are PDF, DOCX, and TXT. The remaining document endpoints list, retrieve metadata, and delete documents.

Create a CV import with `POST /api/v1/profiles/:profileID/imports` and `{ "documentId": "..." }`. Imports are synchronous for now and expose statuses suitable for a future asynchronous worker.

- `GET /api/v1/profiles/:profileID/imports`
- `GET /api/v1/profiles/:profileID/imports/:importID`
- `GET /api/v1/profiles/:profileID/imports/:importID/candidates`
- `PATCH /api/v1/profiles/:profileID/imports/:importID/candidates/:candidateID` with action `accept`, `reject`, or `merge`
- `POST /api/v1/profiles/:profileID/imports/:importID/apply`

Pending candidates block apply. Accepted candidates become relational records; rejected candidates are ignored. Merge candidates remain proposals and never overwrite canonical conflicts. Applied records receive document/import evidence.
