# Applications API

Catalog endpoints provide list/create/get/update for `/api/v1/institutions`, `/programmes`, and `/scholarships`. Programme lists accept `institutionId`; scholarship lists accept `institutionId`, `country`, and `degreeLevel`.

Applications:

- `GET|POST /api/v1/applications`
- `GET|PATCH|DELETE /api/v1/applications/{applicationID}`
- `GET /api/v1/applications/{applicationID}/summary`
- `GET /api/v1/applications/{applicationID}/readiness`
- `GET /api/v1/applications/{applicationID}/eligibility`

Creation takes `name` and `parentProfileId`, with optional institution, programme, scholarship, intake, country, and priority fields. It creates and links a derived Application Profile transactionally. Filters are `status`, `country`, and `intakeYear`.

Application-scoped CRUD routes exist for requirements, deadlines, funding, contacts, supervisors, URLs, and tasks. Requirement evidence supports list/create plus `evidence-suggestions`. `GET /api/v1/deadlines/upcoming?days=30` returns exact deadlines in chronological order. `GET /api/v1/dashboard` returns compact operational attention data.
