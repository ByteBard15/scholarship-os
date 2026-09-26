# Application Prefill API

- `POST /api/v1/applications/{applicationID}/prefill`
- `GET /api/v1/applications/{applicationID}/preparation`
- `GET|POST /api/v1/applications/{applicationID}/fields`
- `PATCH|DELETE /api/v1/applications/{applicationID}/fields/{fieldID}`

The prefill body accepts `trigger` and `regenerate`. Default behavior is conservative: preserve approved/user answers, skip completed Information Requests, and evaluate unresolved targets only.
